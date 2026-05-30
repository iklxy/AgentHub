// Date: 2026-05-30
// Author: XinYang Li

package agent

// AgentMessage is a JSONL message received from the Python agent process stdout.
type AgentMessage struct {
	Type      string         `json:"type"`
	RequestID string         `json:"requestId"`
	ToolName  string         `json:"toolName"`
	Input     map[string]any `json:"input"`
	Result    string         `json:"result"`
	SessionID string         `json:"session_id"`
}

// PermissionResponse is the JSON message written to the Python agent process stdin.
type PermissionResponse struct {
	Type         string         `json:"type"`
	RequestID    string         `json:"requestId"`
	Behavior     string         `json:"behavior"`
	UpdatedInput map[string]any `json:"updatedInput,omitempty"`
	Message      string         `json:"message,omitempty"`
}

// SSERequest is the permission request pushed to frontend via SSE.
type SSERequest struct {
	RequestID string         `json:"requestId"`
	SessionID string         `json:"sessionId"`
	ToolName  string         `json:"toolName"`
	Input     map[string]any `json:"input"`
	CreatedAt string         `json:"createdAt"`
}

// FrontendResponse is the permission decision sent by the frontend via REST.
type FrontendResponse struct {
	Behavior     string         `json:"behavior"`
	UpdatedInput map[string]any `json:"updatedInput,omitempty"`
	Message      string         `json:"message,omitempty"`
}
