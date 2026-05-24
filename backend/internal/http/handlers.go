// Date: 2026-05-25
// Author: XinYang Li

package httpx

import (
	"encoding/json"
	"errors"
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
 * GetConversations returns the default main agent conversation for a task.
 * Params:
 * - writer: the HTTP response writer.
 * - request: the incoming HTTP request that contains the task conversation path.
 */
func (h *Handlers) GetConversations(writer http.ResponseWriter, request *http.Request) {
	userID, authErr := getAuthenticatedUserID(request)
	if authErr != nil {
		h.logFailure(request, "conversations_get_auth_user", authErr)
		WriteError(writer, http.StatusUnauthorized, authErr.Error())
		return
	}

	taskID, action, err := parseTaskSubroute(request.URL.Path)
	if err != nil || action != "conversations" {
		if err != nil {
			h.logFailure(request, "conversations_get_parse_route", err, "userId", userID)
		}
		WriteError(writer, http.StatusBadRequest, "invalid conversations route")
		return
	}

	conversation, err := h.Store.GetConversation(userID, taskID)
	if err != nil {
		h.logFailure(request, "conversations_get", err, "userId", userID, "taskId", taskID)
		WriteError(writer, http.StatusNotFound, err.Error())
		return
	}

	WriteJSON(writer, http.StatusOK, []domain.Conversation{conversation})
}

/**
 * GetMessages returns the current transcript for a task.
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

	messages, err := h.Store.ListMessages(userID, taskID)
	if err != nil {
		h.logFailure(request, "messages_list", err, "userId", userID, "taskId", taskID)
		WriteError(writer, http.StatusNotFound, err.Error())
		return
	}

	WriteJSON(writer, http.StatusOK, messages)
}

/**
 * CreateMessage executes the main agent and stores the resulting message pair.
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

	history, err := h.Store.ListMessages(userID, taskID)
	if err != nil {
		h.logFailure(request, "message_create_list_history", err, "userId", userID, "taskId", taskID)
		WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

	assistantContent, err := h.AgentService.RunMainAgent(task, history, input.Content)
	if err != nil {
		h.logFailure(request, "message_create_run_agent", err, "userId", userID, "taskId", taskID)
		WriteError(writer, http.StatusBadGateway, "agent execution failed")
		return
	}

	userMessage, assistantMessage, err := h.Store.CreateMessagePair(userID, taskID, input.Content, assistantContent)
	if err != nil {
		h.logFailure(request, "message_create_store_pair", err, "userId", userID, "taskId", taskID)
		WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

	h.Logger.Info("message round stored", "taskId", taskID)
	WriteJSON(writer, http.StatusCreated, map[string]domain.Message{
		"userMessage":      userMessage,
		"assistantMessage": assistantMessage,
	})
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

/**
 * Error returns the route error text.
 * Params:
 * - none: the error carries its message inside the receiver.
 */
func (e *routeError) Error() string {
	return e.message
}

func getAuthenticatedUserID(request *http.Request) (string, error) {
	authorization := strings.TrimSpace(request.Header.Get("Authorization"))
	if authorization == "" {
		return "", errors.New("missing authorization token")
	}

	token := strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
	if token == "" || token == authorization {
		return "", errors.New("invalid authorization token")
	}

	return token, nil
}
