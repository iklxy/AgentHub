// Date: 2026-05-25
// Author: XinYang Li

package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/lixinyang/agenthub/backend/internal/agent"
	"github.com/lixinyang/agenthub/backend/internal/domain"
	"github.com/lixinyang/agenthub/backend/internal/store"
)

// Handlers groups the dependencies used by the HTTP layer.
type Handlers struct {
	Logger       *slog.Logger
	Store        store.Repository
	AgentService *agent.Service
}

func (h *Handlers) logFailure(request *http.Request, operation string, err error, attrs ...any) {
	fields := []any{
		"requestId", getRequestID(request),
		"operation", operation,
		"method", request.Method,
		"path", request.URL.Path,
		"error", err,
	}
	fields = append(fields, attrs...)
	h.Logger.Error("http handler failed", fields...)
}

/**
 * Health serves a lightweight readiness check.
 * Params:
 * - writer: the HTTP response writer.
 * - request: the incoming HTTP request metadata.
 */
func (h *Handlers) Health(writer http.ResponseWriter, request *http.Request) {
	WriteJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
}

/**
 * Register creates a user and returns a token for front-end storage.
 * Params:
 * - writer: the HTTP response writer.
 * - request: the incoming HTTP request that carries the register payload.
 */
func (h *Handlers) Register(writer http.ResponseWriter, request *http.Request) {
	var input domain.RegisterRequest

	if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
		h.logFailure(request, "register_decode", err)
		WriteError(writer, http.StatusBadRequest, "invalid register payload")
		return
	}

	user, err := h.Store.CreateUser(input)
	if err != nil {
		h.logFailure(request, "register_create_user", err, "email", input.Email)
		WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

	h.Logger.Info("register request accepted", "email", input.Email, "userId", user.ID)
	WriteJSON(writer, http.StatusCreated, map[string]any{
		"token": user.ID,
		"user":  user,
	})
}

/**
 * Login verifies credentials and returns a user token.
 * Params:
 * - writer: the HTTP response writer.
 * - request: the incoming HTTP request that carries the login payload.
 */
func (h *Handlers) Login(writer http.ResponseWriter, request *http.Request) {
	var input domain.LoginRequest

	if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
		h.logFailure(request, "login_decode", err)
		WriteError(writer, http.StatusBadRequest, "invalid login payload")
		return
	}

	user, err := h.Store.AuthenticateUser(input)
	if err != nil {
		h.logFailure(request, "login_authenticate", err, "email", input.Email)
		WriteError(writer, http.StatusUnauthorized, err.Error())
		return
	}

	h.Logger.Info("login request accepted", "email", input.Email, "userId", user.ID)
	WriteJSON(writer, http.StatusOK, map[string]any{
		"token": user.ID,
		"user":  user,
	})
}

/**
 * Me returns the authenticated user snapshot resolved from the bearer token.
 * Params:
 * - writer: the HTTP response writer.
 * - request: the incoming HTTP request metadata.
 */
func (h *Handlers) Me(writer http.ResponseWriter, request *http.Request) {
	userID, err := getAuthenticatedUserID(request)
	if err != nil {
		h.logFailure(request, "me_get_auth_user", err)
		WriteError(writer, http.StatusUnauthorized, err.Error())
		return
	}

	user, err := h.Store.GetUserByID(userID)
	if err != nil {
		h.logFailure(request, "me_get_user", err, "userId", userID)
		WriteError(writer, http.StatusUnauthorized, err.Error())
		return
	}

	WriteJSON(writer, http.StatusOK, user)
}

/**
 * GetWorkspace returns the current workspace snapshot for the authenticated user.
 * Params:
 * - writer: the HTTP response writer.
 * - request: the incoming HTTP request metadata.
 */
func (h *Handlers) GetWorkspace(writer http.ResponseWriter, request *http.Request) {
	userID, err := getAuthenticatedUserID(request)
	if err != nil {
		h.logFailure(request, "workspace_get_auth_user", err)
		WriteError(writer, http.StatusUnauthorized, err.Error())
		return
	}

	workspace, err := h.Store.GetWorkspaceByUserID(userID)
	if err != nil {
		h.logFailure(request, "workspace_get", err, "userId", userID)
		WriteError(writer, http.StatusNotFound, err.Error())
		return
	}

	WriteJSON(writer, http.StatusOK, workspace)
}

/**
 * UpdateWorkspace updates workspace metadata for the authenticated user.
 * Params:
 * - writer: the HTTP response writer.
 * - request: the incoming HTTP request that carries the workspace payload.
 */
func (h *Handlers) UpdateWorkspace(writer http.ResponseWriter, request *http.Request) {
	userID, err := getAuthenticatedUserID(request)
	if err != nil {
		h.logFailure(request, "workspace_update_auth_user", err)
		WriteError(writer, http.StatusUnauthorized, err.Error())
		return
	}

	var input domain.UpdateWorkspaceRequest
	if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
		h.logFailure(request, "workspace_update_decode", err, "userId", userID)
		WriteError(writer, http.StatusBadRequest, "invalid workspace payload")
		return
	}

	workspace, err := h.Store.UpdateWorkspace(userID, input)
	if err != nil {
		h.logFailure(request, "workspace_update", err, "userId", userID)
		WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

	h.Logger.Info("workspace updated", "workspaceId", workspace.ID)
	WriteJSON(writer, http.StatusOK, workspace)
}

/**
 * ListTasks returns all current workspace tasks for the authenticated user.
 * Params:
 * - writer: the HTTP response writer.
 * - request: the incoming HTTP request metadata.
 */
func (h *Handlers) ListTasks(writer http.ResponseWriter, request *http.Request) {
	userID, err := getAuthenticatedUserID(request)
	if err != nil {
		h.logFailure(request, "tasks_list_auth_user", err)
		WriteError(writer, http.StatusUnauthorized, err.Error())
		return
	}

	tasks, err := h.Store.ListTasks(userID)
	if err != nil {
		h.logFailure(request, "tasks_list", err, "userId", userID)
		WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(writer, http.StatusOK, tasks)
}

/**
 * CreateTask creates a new task for the authenticated user.
 * Params:
 * - writer: the HTTP response writer.
 * - request: the incoming HTTP request that carries the task payload.
 */
func (h *Handlers) CreateTask(writer http.ResponseWriter, request *http.Request) {
	userID, err := getAuthenticatedUserID(request)
	if err != nil {
		h.logFailure(request, "task_create_auth_user", err)
		WriteError(writer, http.StatusUnauthorized, err.Error())
		return
	}

	var input domain.CreateTaskRequest
	if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
		h.logFailure(request, "task_create_decode", err, "userId", userID)
		WriteError(writer, http.StatusBadRequest, "invalid task payload")
		return
	}

	task, err := h.Store.CreateTask(userID, input)
	if err != nil {
		h.logFailure(request, "task_create", err, "userId", userID, "title", input.Title)
		WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

	h.Logger.Info("task created", "taskId", task.ID, "title", task.Title)
	WriteJSON(writer, http.StatusCreated, task)
}

/**
 * GetTask routes task detail queries using the task identifier from the request path.
 * Params:
 * - writer: the HTTP response writer.
 * - request: the incoming HTTP request that contains the task detail path.
 */
func (h *Handlers) GetTask(writer http.ResponseWriter, request *http.Request) {
	userID, authErr := getAuthenticatedUserID(request)
	if authErr != nil {
		h.logFailure(request, "task_get_auth_user", authErr)
		WriteError(writer, http.StatusUnauthorized, authErr.Error())
		return
	}

	taskID, _, err := parseTaskSubroute(request.URL.Path)
	if err != nil {
		h.logFailure(request, "task_get_parse_route", err, "userId", userID)
		WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

	task, err := h.Store.GetTask(userID, taskID)
	if err != nil {
		h.logFailure(request, "task_get", err, "userId", userID, "taskId", taskID)
		WriteError(writer, http.StatusNotFound, err.Error())
		return
	}

	WriteJSON(writer, http.StatusOK, task)
}

/**
 * ListSessions returns the task session list for the current user.
 * Params:
 * - writer: the HTTP response writer.
 * - request: the incoming HTTP request that contains the task session path.
 */
func (h *Handlers) ListSessions(writer http.ResponseWriter, request *http.Request) {
	userID, authErr := getAuthenticatedUserID(request)
	if authErr != nil {
		h.logFailure(request, "sessions_list_auth_user", authErr)
		WriteError(writer, http.StatusUnauthorized, authErr.Error())
		return
	}

	taskID, action, err := parseTaskSubroute(request.URL.Path)
	if err != nil || action != "sessions" {
		if err != nil {
			h.logFailure(request, "sessions_list_parse_route", err, "userId", userID)
		}
		WriteError(writer, http.StatusBadRequest, "invalid sessions route")
		return
	}

	sessions, err := h.Store.ListSessions(userID, taskID)
	if err != nil {
		h.logFailure(request, "sessions_list", err, "userId", userID, "taskId", taskID)
		WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(writer, http.StatusOK, sessions)
}

/**
 * ListTaskSessionAgents returns available branch-session agent choices for the task.
 * Params:
 * - writer: the HTTP response writer.
 * - request: the incoming HTTP request that contains the task agent-selection path.
 */
func (h *Handlers) ListTaskSessionAgents(writer http.ResponseWriter, request *http.Request) {
	userID, authErr := getAuthenticatedUserID(request)
	if authErr != nil {
		h.logFailure(request, "session_agents_list_auth_user", authErr)
		WriteError(writer, http.StatusUnauthorized, authErr.Error())
		return
	}

	taskID, action, err := parseTaskSubroute(request.URL.Path)
	if err != nil || action != "session-agents" {
		if err != nil {
			h.logFailure(request, "session_agents_list_parse_route", err, "userId", userID)
		}
		WriteError(writer, http.StatusBadRequest, "invalid session agents route")
		return
	}

	agents, err := h.Store.ListSessionAgents(userID, taskID)
	if err != nil {
		h.logFailure(request, "session_agents_list", err, "userId", userID, "taskId", taskID)
		WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(writer, http.StatusOK, agents)
}

/**
 * ListAgents returns enabled agents for the current task context selected by query string.
 * Params:
 * - writer: the HTTP response writer.
 * - request: the incoming HTTP request that may contain the taskId query parameter.
 */
func (h *Handlers) ListAgents(writer http.ResponseWriter, request *http.Request) {
	userID, authErr := getAuthenticatedUserID(request)
	if authErr != nil {
		h.logFailure(request, "agents_list_auth_user", authErr)
		WriteError(writer, http.StatusUnauthorized, authErr.Error())
		return
	}

	taskID := strings.TrimSpace(request.URL.Query().Get("taskId"))
	if taskID == "" {
		WriteError(writer, http.StatusBadRequest, "taskId is required")
		return
	}

	agents, err := h.Store.ListSessionAgents(userID, taskID)
	if err != nil {
		h.logFailure(request, "agents_list", err, "userId", userID, "taskId", taskID)
		WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(writer, http.StatusOK, agents)
}

/**
 * CreateSession creates one new task session for the authenticated user.
 * Params:
 * - writer: the HTTP response writer.
 * - request: the incoming HTTP request that carries the session payload.
 */
func (h *Handlers) CreateSession(writer http.ResponseWriter, request *http.Request) {
	userID, authErr := getAuthenticatedUserID(request)
	if authErr != nil {
		h.logFailure(request, "session_create_auth_user", authErr)
		WriteError(writer, http.StatusUnauthorized, authErr.Error())
		return
	}

	var input domain.CreateSessionRequest
	if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
		h.logFailure(request, "session_create_decode", err, "userId", userID)
		WriteError(writer, http.StatusBadRequest, "invalid session payload")
		return
	}

	session, err := h.Store.CreateSession(userID, input)
	if err != nil {
		h.logFailure(request, "session_create", err, "userId", userID, "taskId", input.TaskID, "primaryAgentId", input.PrimaryAgentID)
		WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

	h.Logger.Info("session created", "sessionId", session.ID, "taskId", session.TaskID, "agentId", session.PrimaryAgentID)
	WriteJSON(writer, http.StatusCreated, session)
}

/**
 * GetSession returns one session detail for the authenticated user.
 * Params:
 * - writer: the HTTP response writer.
 * - request: the incoming HTTP request that contains the session detail path.
 */
func (h *Handlers) GetSession(writer http.ResponseWriter, request *http.Request) {
	userID, authErr := getAuthenticatedUserID(request)
	if authErr != nil {
		h.logFailure(request, "session_get_auth_user", authErr)
		WriteError(writer, http.StatusUnauthorized, authErr.Error())
		return
	}

	sessionID, err := parseSessionRoute(request.URL.Path)
	if err != nil {
		h.logFailure(request, "session_get_parse_route", err, "userId", userID)
		WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

	session, err := h.Store.GetSession(userID, sessionID)
	if err != nil {
		h.logFailure(request, "session_get", err, "userId", userID, "sessionId", sessionID)
		WriteError(writer, http.StatusNotFound, err.Error())
		return
	}

	WriteJSON(writer, http.StatusOK, session)
}

/**
 * UpdateSession updates editable metadata for one session.
 * Params:
 * - writer: the HTTP response writer.
 * - request: the incoming HTTP request that carries the session update payload.
 */
func (h *Handlers) UpdateSession(writer http.ResponseWriter, request *http.Request) {
	userID, authErr := getAuthenticatedUserID(request)
	if authErr != nil {
		h.logFailure(request, "session_update_auth_user", authErr)
		WriteError(writer, http.StatusUnauthorized, authErr.Error())
		return
	}

	sessionID, err := parseSessionRoute(request.URL.Path)
	if err != nil {
		h.logFailure(request, "session_update_parse_route", err, "userId", userID)
		WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

	var input domain.UpdateSessionRequest
	if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
		h.logFailure(request, "session_update_decode", err, "userId", userID, "sessionId", sessionID)
		WriteError(writer, http.StatusBadRequest, "invalid session payload")
		return
	}

	session, err := h.Store.UpdateSession(userID, sessionID, input)
	if err != nil {
		h.logFailure(request, "session_update", err, "userId", userID, "sessionId", sessionID)
		WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(writer, http.StatusOK, session)
}

/**
 * GetMessages returns the current transcript for a task session.
 * Params:
 * - writer: the HTTP response writer.
 * - request: the incoming HTTP request that contains the task messages path.
 */
func (h *Handlers) GetMessages(writer http.ResponseWriter, request *http.Request) {
	userID, authErr := getAuthenticatedUserID(request)
	if authErr != nil {
		h.logFailure(request, "messages_list_auth_user", authErr)
		WriteError(writer, http.StatusUnauthorized, authErr.Error())
		return
	}

	taskID, action, err := parseTaskSubroute(request.URL.Path)
	if err != nil || action != "messages" {
		if err != nil {
			h.logFailure(request, "messages_list_parse_route", err, "userId", userID)
		}
		WriteError(writer, http.StatusBadRequest, "invalid messages route")
		return
	}

	sessionID := strings.TrimSpace(request.URL.Query().Get("sessionId"))
	if sessionID == "" {
		WriteError(writer, http.StatusBadRequest, "sessionId is required")
		return
	}

	messages, err := h.Store.ListMessages(userID, taskID, sessionID)
	if err != nil {
		h.logFailure(request, "messages_list", err, "userId", userID, "taskId", taskID, "sessionId", sessionID)
		WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(writer, http.StatusOK, messages)
}

/**
 * CreateMessage executes the active session agent and stores the resulting message pair.
 * Params:
 * - writer: the HTTP response writer.
 * - request: the incoming HTTP request that carries the message payload.
 */
func (h *Handlers) CreateMessage(writer http.ResponseWriter, request *http.Request) {
	userID, authErr := getAuthenticatedUserID(request)
	if authErr != nil {
		h.logFailure(request, "message_create_auth_user", authErr)
		WriteError(writer, http.StatusUnauthorized, authErr.Error())
		return
	}

	taskID, action, err := parseTaskSubroute(request.URL.Path)
	if err != nil || action != "messages" {
		if err != nil {
			h.logFailure(request, "message_create_parse_route", err, "userId", userID)
		}
		WriteError(writer, http.StatusBadRequest, "invalid messages route")
		return
	}

	var input domain.CreateMessageRequest
	if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
		h.logFailure(request, "message_create_decode", err, "userId", userID, "taskId", taskID)
		WriteError(writer, http.StatusBadRequest, "invalid message payload")
		return
	}

	task, err := h.Store.GetTask(userID, taskID)
	if err != nil {
		h.logFailure(request, "message_create_get_task", err, "userId", userID, "taskId", taskID)
		WriteError(writer, http.StatusNotFound, err.Error())
		return
	}

	session, err := h.Store.GetSession(userID, input.SessionID)
	if err != nil {
		h.logFailure(request, "message_create_get_session", err, "userId", userID, "taskId", taskID, "sessionId", input.SessionID)
		WriteError(writer, http.StatusNotFound, err.Error())
		return
	}
	if session.TaskID != taskID {
		h.logFailure(request, "message_create_session_mismatch", errors.New("session does not belong to task"), "userId", userID, "taskId", taskID, "sessionId", input.SessionID)
		WriteError(writer, http.StatusBadRequest, "session does not belong to the task")
		return
	}

	assistantContent, err := h.AgentService.RunSessionAgent(task, session, input.Content)
	if err != nil {
		h.logFailure(
			request,
			"message_create_run_agent",
			err,
			"userId", userID,
			"taskId", taskID,
			"sessionId", session.ID,
			"agentId", session.PrimaryAgentID,
			"agentName", session.PrimaryAgentName,
		)
		if errors.Is(err, agent.ErrRuntimeNotImplemented) {
			WriteError(writer, http.StatusNotImplemented, err.Error())
			return
		}
		WriteError(writer, http.StatusBadGateway, "agent execution failed")
		return
	}

	userMessage, assistantMessage, err := h.Store.CreateMessagePair(userID, taskID, session.ID, input.Content, assistantContent, nil)
	if err != nil {
		h.logFailure(request, "message_create_store_pair", err, "userId", userID, "taskId", taskID, "sessionId", session.ID)
		WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

	h.Logger.Info("message round stored", "taskId", taskID, "sessionId", session.ID, "agentId", session.PrimaryAgentID)
	WriteJSON(writer, http.StatusCreated, map[string]domain.Message{
		"userMessage":      userMessage,
		"assistantMessage": assistantMessage,
	})
}

/**
 * CreateQuotedMessage executes one derived quote message round against multiple source messages.
 * Params:
 * - writer: the HTTP response writer.
 * - request: the incoming HTTP request that carries the quote payload.
 */
func (h *Handlers) CreateQuotedMessage(writer http.ResponseWriter, request *http.Request) {
	userID, authErr := getAuthenticatedUserID(request)
	if authErr != nil {
		h.logFailure(request, "message_quote_auth_user", authErr)
		WriteError(writer, http.StatusUnauthorized, authErr.Error())
		return
	}

	var input domain.CreateDerivedMessageRequest
	if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
		h.logFailure(request, "message_quote_decode", err, "userId", userID)
		WriteError(writer, http.StatusBadRequest, "invalid quote payload")
		return
	}

	trimmedUserContent := strings.TrimSpace(input.Content)
	if trimmedUserContent == "" {
		WriteError(writer, http.StatusBadRequest, "content is required")
		return
	}

	if len(input.MessageIDs) == 0 {
		WriteError(writer, http.StatusBadRequest, "messageIds are required")
		return
	}

	sourceMessages, err := h.Store.GetMessagesByIDs(userID, input.MessageIDs)
	if err != nil {
		h.logFailure(request, "message_quote_get_sources", err, "userId", userID)
		WriteError(writer, http.StatusNotFound, err.Error())
		return
	}

	baseMessage := sourceMessages[0]
	task, err := h.Store.GetTask(userID, baseMessage.TaskID)
	if err != nil {
		h.logFailure(request, "message_quote_get_task", err, "userId", userID, "taskId", baseMessage.TaskID)
		WriteError(writer, http.StatusNotFound, err.Error())
		return
	}

	session, err := h.Store.GetSession(userID, baseMessage.SessionID)
	if err != nil {
		h.logFailure(request, "message_quote_get_session", err, "userId", userID, "sessionId", baseMessage.SessionID)
		WriteError(writer, http.StatusNotFound, err.Error())
		return
	}

	for _, sourceMessage := range sourceMessages[1:] {
		if sourceMessage.TaskID != task.ID || sourceMessage.SessionID != session.ID {
			WriteError(writer, http.StatusBadRequest, "quoted messages must belong to the same session")
			return
		}
	}

	composedPrompt := buildQuotedPrompt(sourceMessages, trimmedUserContent)
	assistantContent, err := h.AgentService.RunSessionAgent(task, session, composedPrompt)
	if err != nil {
		h.logFailure(
			request,
			"message_quote_run_agent",
			err,
			"userId", userID,
			"taskId", task.ID,
			"sessionId", session.ID,
		)
		if errors.Is(err, agent.ErrRuntimeNotImplemented) {
			WriteError(writer, http.StatusNotImplemented, err.Error())
			return
		}
		WriteError(writer, http.StatusBadGateway, "agent execution failed")
		return
	}

	userMessage, assistantMessage, err := h.Store.CreateMessagePair(userID, task.ID, session.ID, trimmedUserContent, assistantContent, nil)
	if err != nil {
		h.logFailure(request, "message_quote_store_pair", err, "userId", userID, "taskId", task.ID, "sessionId", session.ID)
		WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(writer, http.StatusCreated, map[string]domain.Message{
		"userMessage":      userMessage,
		"assistantMessage": assistantMessage,
	})
}

/**
 * CreateMessageFromAction executes reply message creation against one source message.
 * Params:
 * - writer: the HTTP response writer.
 * - request: the incoming HTTP request that carries the reply payload and message route.
 */
func (h *Handlers) CreateMessageFromAction(writer http.ResponseWriter, request *http.Request) {
	userID, authErr := getAuthenticatedUserID(request)
	if authErr != nil {
		h.logFailure(request, "message_action_auth_user", authErr)
		WriteError(writer, http.StatusUnauthorized, authErr.Error())
		return
	}

	messageID, action, err := parseMessageActionRoute(request.URL.Path)
	if err != nil {
		h.logFailure(request, "message_action_parse_route", err, "userId", userID)
		WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

	if action != "reply" && action != "regenerate" {
		WriteError(writer, http.StatusBadRequest, "invalid message action")
		return
	}

	var input domain.CreateDerivedMessageRequest
	if action == "reply" {
		if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
			h.logFailure(request, "message_reply_decode", err, "userId", userID, "messageId", messageID)
			WriteError(writer, http.StatusBadRequest, "invalid reply payload")
			return
		}
	}

	sourceMessage, err := h.Store.GetMessageByID(userID, messageID)
	if err != nil {
		h.logFailure(request, "message_action_get_source", err, "userId", userID, "messageId", messageID, "action", action)
		WriteError(writer, http.StatusNotFound, err.Error())
		return
	}

	task, err := h.Store.GetTask(userID, sourceMessage.TaskID)
	if err != nil {
		h.logFailure(request, "message_action_get_task", err, "userId", userID, "taskId", sourceMessage.TaskID, "messageId", messageID, "action", action)
		WriteError(writer, http.StatusNotFound, err.Error())
		return
	}

	session, err := h.Store.GetSession(userID, sourceMessage.SessionID)
	if err != nil {
		h.logFailure(request, "message_action_get_session", err, "userId", userID, "sessionId", sourceMessage.SessionID, "messageId", messageID, "action", action)
		WriteError(writer, http.StatusNotFound, err.Error())
		return
	}

	if action == "regenerate" {
		h.handleRegenerateMessage(writer, request, userID, task, session, sourceMessage)
		return
	}

	trimmedUserContent := strings.TrimSpace(input.Content)
	if trimmedUserContent == "" {
		WriteError(writer, http.StatusBadRequest, "content is required")
		return
	}

	composedPrompt, replyToMessageID := buildReplyPrompt(sourceMessage, trimmedUserContent)
	assistantContent, err := h.AgentService.RunSessionAgent(task, session, composedPrompt)
	if err != nil {
		h.logFailure(
			request,
			"message_reply_run_agent",
			err,
			"userId", userID,
			"taskId", task.ID,
			"sessionId", session.ID,
			"sourceMessageId", sourceMessage.ID,
		)
		if errors.Is(err, agent.ErrRuntimeNotImplemented) {
			WriteError(writer, http.StatusNotImplemented, err.Error())
			return
		}
		WriteError(writer, http.StatusBadGateway, "agent execution failed")
		return
	}

	userMessage, assistantMessage, err := h.Store.CreateMessagePair(userID, task.ID, session.ID, trimmedUserContent, assistantContent, replyToMessageID)
	if err != nil {
		h.logFailure(request, "message_reply_store_pair", err, "userId", userID, "taskId", task.ID, "sessionId", session.ID, "sourceMessageId", sourceMessage.ID)
		WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(writer, http.StatusCreated, map[string]domain.Message{
		"userMessage":      userMessage,
		"assistantMessage": assistantMessage,
	})
}

/**
 * handleRegenerateMessage reruns the active agent for one assistant message and appends the regenerated answer.
 * Params:
 * - writer: the HTTP response writer.
 * - request: the incoming HTTP request that triggered the regenerate action.
 * - userID: the authenticated user identifier from the bearer token.
 * - task: the task that owns the selected source message.
 * - session: the session that owns the selected source message.
 * - sourceMessage: the assistant message selected for regeneration.
 */
func (h *Handlers) handleRegenerateMessage(
	writer http.ResponseWriter,
	request *http.Request,
	userID string,
	task domain.Task,
	session domain.Session,
	sourceMessage domain.Message,
) {
	if sourceMessage.Role != "assistant" {
		WriteError(writer, http.StatusBadRequest, "only assistant messages can be regenerated")
		return
	}

	composedPrompt := buildRegeneratePrompt(sourceMessage)
	assistantContent, err := h.AgentService.RunSessionAgent(task, session, composedPrompt)
	if err != nil {
		h.logFailure(
			request,
			"message_regenerate_run_agent",
			err,
			"userId", userID,
			"taskId", task.ID,
			"sessionId", session.ID,
			"sourceMessageId", sourceMessage.ID,
		)
		if errors.Is(err, agent.ErrRuntimeNotImplemented) {
			WriteError(writer, http.StatusNotImplemented, err.Error())
			return
		}
		WriteError(writer, http.StatusBadGateway, "agent execution failed")
		return
	}

	assistantMessage, err := h.Store.CreateAssistantMessage(userID, task.ID, session.ID, assistantContent)
	if err != nil {
		h.logFailure(request, "message_regenerate_store", err, "userId", userID, "taskId", task.ID, "sessionId", session.ID, "sourceMessageId", sourceMessage.ID)
		WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(writer, http.StatusCreated, map[string]domain.Message{
		"assistantMessage": assistantMessage,
	})
}

/**
 * buildQuotedPrompt creates the agent-facing prompt for one multi-message quote action.
 * Params:
 * - sourceMessages: the quoted messages selected by the user in display order.
 * - userContent: the raw user input typed in the composer.
 * Returns:
 * - the composed prompt sent to the active agent.
 */
func buildQuotedPrompt(sourceMessages []domain.Message, userContent string) string {
	segments := make([]string, 0, len(sourceMessages))
	for index, sourceMessage := range sourceMessages {
		segments = append(segments, fmt.Sprintf("引用消息%d：%s", index+1, sourceMessage.Content))
	}

	return fmt.Sprintf("以下是前文中的多个上下文片段：\n%s\n\n请结合这些上下文回答用户的新问题：\n%s", strings.Join(segments, "\n"), userContent)
}

/**
 * buildReplyPrompt creates the agent-facing prompt for one reply action.
 * Params:
 * - sourceMessage: the source message selected by the user.
 * - userContent: the raw user input typed in the composer.
 * Returns:
 * - the composed prompt sent to the active agent and the reply target identifier.
 */
func buildReplyPrompt(sourceMessage domain.Message, userContent string) (string, *string) {
	replyToID := sourceMessage.ID
	return fmt.Sprintf("请基于以下消息继续回答：\n%s\n\n用户回复内容：\n%s", sourceMessage.Content, userContent), &replyToID
}

/**
 * buildRegeneratePrompt creates the agent-facing prompt for one assistant-message regeneration action.
 * Params:
 * - sourceMessage: the assistant message selected by the user for regeneration.
 * Returns:
 * - the composed prompt sent to the active agent.
 */
func buildRegeneratePrompt(sourceMessage domain.Message) string {
	return fmt.Sprintf("针对该条消息%s，重新生成答案。", sourceMessage.Content)
}

/**
 * parseTaskSubroute extracts the task identifier and trailing action from a task route.
 * Params:
 * - path: the request path that starts with /api/tasks/{taskId}.
 */
func parseTaskSubroute(path string) (string, string, error) {
	trimmed := strings.TrimPrefix(path, "/api/tasks/")
	segments := strings.Split(strings.Trim(trimmed, "/"), "/")

	if len(segments) == 0 || segments[0] == "" {
		return "", "", errorsNew("task id is required")
	}

	if len(segments) == 1 {
		return segments[0], "", nil
	}

	return segments[0], segments[1], nil
}

/**
 * parseSessionRoute extracts the session identifier from /api/sessions/{sessionId}.
 * Params:
 * - path: the request path that starts with /api/sessions/.
 */
func parseSessionRoute(path string) (string, error) {
	trimmed := strings.TrimPrefix(path, "/api/sessions/")
	sessionID := strings.Trim(trimmed, "/")
	if sessionID == "" {
		return "", errorsNew("session id is required")
	}
	return sessionID, nil
}

/**
 * parseMessageActionRoute extracts the message identifier and action from /api/messages/{messageId}/{action}.
 * Params:
 * - path: the request path that starts with /api/messages/.
 */
func parseMessageActionRoute(path string) (string, string, error) {
	trimmed := strings.TrimPrefix(path, "/api/messages/")
	segments := strings.Split(strings.Trim(trimmed, "/"), "/")

	if len(segments) < 2 || segments[0] == "" || segments[1] == "" {
		return "", "", errorsNew("message id and action are required")
	}

	return segments[0], segments[1], nil
}

/**
 * errorsNew creates a lightweight error without importing the errors package twice in many handlers.
 * Params:
 * - message: the error text returned to the caller.
 */
func errorsNew(message string) error {
	return &routeError{message: message}
}

type routeError struct {
	message string
}

func (r *routeError) Error() string {
	return r.message
}
