// Date: 2026-05-30
// Author: XinYang Li

package agent

import (
	"log/slog"
	"sync"
	"time"
)

// PermissionBroker manages the bidirectional flow of permission requests between
// the Python agent process and frontend clients. It holds pending request channels
// that the agent goroutine blocks on, and SSE subscriber channels that push
// requests to connected frontend clients.
type PermissionBroker struct {
	mu          sync.Mutex
	pending     map[string]chan PermissionResponse
	sseClients  map[string]map[chan SSERequest]struct{}
	logger      *slog.Logger
	responseTTL time.Duration
}

/**
 * NewPermissionBroker creates a broker for managing agent permission requests.
 * Params:
 * - logger: the shared backend logger used for diagnostics.
 */
func NewPermissionBroker(logger *slog.Logger) *PermissionBroker {
	return &PermissionBroker{
		pending:     make(map[string]chan PermissionResponse),
		sseClients:  make(map[string]map[chan SSERequest]struct{}),
		logger:      logger.With("service", "permission-broker"),
		responseTTL: 5 * time.Minute,
	}
}

/**
 * WaitForResponse blocks until the frontend responds or the TTL expires.
 * Params:
 * - sessionID: the session that owns the permission request.
 * - request: the SSE-ready permission request to push to frontend.
 * Returns:
 * - the permission response from the frontend, or a timeout-denied default.
 */
func (b *PermissionBroker) WaitForResponse(sessionID string, request SSERequest) PermissionResponse {
	responseCh := make(chan PermissionResponse, 1)

	b.mu.Lock()
	b.pending[request.RequestID] = responseCh
	b.mu.Unlock()

	b.pushToSSEClients(sessionID, request)

	b.logger.Info(
		"waiting for permission response",
		"requestId", request.RequestID,
		"sessionId", sessionID,
		"toolName", request.ToolName,
	)

	select {
	case response := <-responseCh:
		b.mu.Lock()
		delete(b.pending, request.RequestID)
		b.mu.Unlock()
		b.logger.Info(
			"permission response received",
			"requestId", request.RequestID,
			"behavior", response.Behavior,
		)
		return response
	case <-time.After(b.responseTTL):
		b.mu.Lock()
		delete(b.pending, request.RequestID)
		b.mu.Unlock()
		b.logger.Warn(
			"permission request timed out",
			"requestId", request.RequestID,
		)
		return PermissionResponse{
			Type:      "permission_response",
			RequestID: request.RequestID,
			Behavior:  "deny",
			Message:   "Approval timed out",
		}
	}
}

/**
 * Respond processes a frontend permission decision and unblocks the waiting agent goroutine.
 * Params:
 * - requestID: the permission request identifier.
 * - response: the frontend's allow/deny decision.
 */
func (b *PermissionBroker) Respond(requestID string, response PermissionResponse) {
	b.mu.Lock()
	ch, ok := b.pending[requestID]
	b.mu.Unlock()

	if !ok {
		b.logger.Warn("response for unknown request", "requestId", requestID)
		return
	}

	response.Type = "permission_response"
	response.RequestID = requestID
	ch <- response
}

/**
 * Subscribe registers a new SSE client for a session and returns a channel
 * that receives permission requests as they arrive.
 * Params:
 * - sessionID: the session to subscribe to.
 * Returns:
 * - a receive-only channel of SSERequest and a cancel function to unsubscribe.
 */
func (b *PermissionBroker) Subscribe(sessionID string) (<-chan SSERequest, func()) {
	ch := make(chan SSERequest, 16)

	b.mu.Lock()
	if b.sseClients[sessionID] == nil {
		b.sseClients[sessionID] = make(map[chan SSERequest]struct{})
	}
	b.sseClients[sessionID][ch] = struct{}{}
	b.mu.Unlock()

	cancel := func() {
		b.mu.Lock()
		delete(b.sseClients[sessionID], ch)
		if len(b.sseClients[sessionID]) == 0 {
			delete(b.sseClients, sessionID)
		}
		b.mu.Unlock()
		close(ch)
	}

	return ch, cancel
}

func (b *PermissionBroker) pushToSSEClients(sessionID string, request SSERequest) {
	b.mu.Lock()
	clients := b.sseClients[sessionID]
	b.mu.Unlock()

	for ch := range clients {
		select {
		case ch <- request:
		default:
			b.logger.Warn("sse client channel full, dropping request", "sessionId", sessionID)
		}
	}
}
