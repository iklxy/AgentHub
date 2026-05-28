// Date: 2026-05-25
// Author: XinYang Li

package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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
 * UploadFiles stores uploaded files or images for one task session before the next message send binds them.
 * Params:
 * - writer: the HTTP response writer.
 * - request: the multipart upload request that carries files, taskId, sessionId, and sourceType.
 */
func (h *Handlers) UploadFiles(writer http.ResponseWriter, request *http.Request) {
	userID, authErr := getAuthenticatedUserID(request)
	if authErr != nil {
		h.logFailure(request, "files_upload_auth_user", authErr)
		WriteError(writer, http.StatusUnauthorized, authErr.Error())
		return
	}

	if err := request.ParseMultipartForm(128 << 20); err != nil {
		h.logFailure(request, "files_upload_parse_form", err, "userId", userID)
		WriteError(writer, http.StatusBadRequest, "invalid upload payload")
		return
	}

	taskID := strings.TrimSpace(request.FormValue("taskId"))
	sessionID := strings.TrimSpace(request.FormValue("sessionId"))
	sourceType := normalizeAttachmentSourceType(request.FormValue("sourceType"))
	if taskID == "" || sessionID == "" {
		WriteError(writer, http.StatusBadRequest, "taskId and sessionId are required")
		return
	}
	if sourceType == "" {
		WriteError(writer, http.StatusBadRequest, "sourceType must be image or file")
		return
	}

	task, err := h.Store.GetTask(userID, taskID)
	if err != nil {
		h.logFailure(request, "files_upload_get_task", err, "userId", userID, "taskId", taskID)
		WriteError(writer, http.StatusNotFound, err.Error())
		return
	}

	session, err := h.Store.GetSession(userID, sessionID)
	if err != nil {
		h.logFailure(request, "files_upload_get_session", err, "userId", userID, "sessionId", sessionID)
		WriteError(writer, http.StatusNotFound, err.Error())
		return
	}
	if session.TaskID != task.ID {
		WriteError(writer, http.StatusBadRequest, "session does not belong to the task")
		return
	}

	fileHeaders := request.MultipartForm.File["files"]
	if len(fileHeaders) == 0 {
		WriteError(writer, http.StatusBadRequest, "files are required")
		return
	}

	attachments := make([]domain.Attachment, 0, len(fileHeaders))
	for _, fileHeader := range fileHeaders {
		openedFile, err := fileHeader.Open()
		if err != nil {
			h.logFailure(request, "files_upload_open_file", err, "userId", userID, "sessionId", sessionID, "fileName", fileHeader.Filename)
			WriteError(writer, http.StatusBadRequest, "failed to read upload file")
			return
		}

		detectedType, detectErr := detectUploadedFileType(openedFile, fileHeader.Header.Get("Content-Type"))
		if detectErr != nil {
			openedFile.Close()
			h.logFailure(request, "files_upload_detect_type", detectErr, "userId", userID, "fileName", fileHeader.Filename)
			WriteError(writer, http.StatusBadRequest, "failed to detect upload file type")
			return
		}

		resolvedSourceType := sourceType
		if strings.HasPrefix(strings.ToLower(detectedType), "image/") {
			resolvedSourceType = "image"
		}

		assetDir, err := h.AgentService.EnsureSessionAssetDir(session, resolvedSourceType)
		if err != nil {
			openedFile.Close()
			h.logFailure(request, "files_upload_prepare_dir", err, "userId", userID, "sessionId", sessionID, "sourceType", resolvedSourceType)
			WriteError(writer, http.StatusInternalServerError, "failed to prepare upload directory")
			return
		}

		safeFileName := sanitizeUploadedFileName(fileHeader.Filename)
		storagePath := filepath.Join(assetDir, buildStoredFileName(safeFileName))
		targetFile, err := os.Create(storagePath)
		if err != nil {
			openedFile.Close()
			h.logFailure(request, "files_upload_create_file", err, "userId", userID, "storagePath", storagePath)
			WriteError(writer, http.StatusInternalServerError, "failed to save uploaded file")
			return
		}

		if _, err := io.Copy(targetFile, openedFile); err != nil {
			targetFile.Close()
			openedFile.Close()
			_ = os.Remove(storagePath)
			h.logFailure(request, "files_upload_copy_file", err, "userId", userID, "storagePath", storagePath)
			WriteError(writer, http.StatusInternalServerError, "failed to persist uploaded file")
			return
		}

		targetFile.Close()
		openedFile.Close()

		attachment, err := h.Store.CreateAttachment(userID, task.ID, session.ID, safeFileName, detectedType, resolvedSourceType, storagePath)
		if err != nil {
			_ = os.Remove(storagePath)
			h.logFailure(request, "files_upload_create_attachment", err, "userId", userID, "taskId", task.ID, "sessionId", session.ID, "fileName", safeFileName)
			WriteError(writer, http.StatusBadRequest, err.Error())
			return
		}

		attachments = append(attachments, attachment)
	}

	WriteJSON(writer, http.StatusCreated, attachments)
}

/**
 * GetFile streams one uploaded attachment back to the front-end for preview or download.
 * Params:
 * - writer: the HTTP response writer.
 * - request: the incoming HTTP request that contains the attachment path.
 */
func (h *Handlers) GetFile(writer http.ResponseWriter, request *http.Request) {
	userID, authErr := getAuthenticatedUserID(request)
	if authErr != nil {
		h.logFailure(request, "files_get_auth_user", authErr)
		WriteError(writer, http.StatusUnauthorized, authErr.Error())
		return
	}

	attachmentID, err := parseFileRoute(request.URL.Path)
	if err != nil {
		h.logFailure(request, "files_get_parse_route", err, "userId", userID)
		WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

	attachment, err := h.Store.GetAttachmentByID(userID, attachmentID)
	if err != nil {
		h.logFailure(request, "files_get_attachment", err, "userId", userID, "attachmentId", attachmentID)
		WriteError(writer, http.StatusNotFound, err.Error())
		return
	}

	dispositionType := "attachment"
	if attachment.SourceType == "image" {
		dispositionType = "inline"
	}
	writer.Header().Set("Content-Type", attachment.FileType)
	writer.Header().Set("Content-Disposition", fmt.Sprintf("%s; filename=%q", dispositionType, attachment.FileName))
	http.ServeFile(writer, request, attachment.StorageKey)
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

	composedPrompt, attachments, err := h.buildAgentPrompt(userID, task.ID, session, input.Content, input.AttachmentIDs)
	if err != nil {
		h.logFailure(request, "message_create_build_prompt", err, "userId", userID, "taskId", taskID, "sessionId", session.ID)
		WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

	assistantContent, err := h.AgentService.RunSessionAgent(task, session, composedPrompt)
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

	userMessage, assistantMessage, err := h.Store.CreateMessagePair(userID, taskID, session.ID, input.Content, assistantContent, nil, attachments)
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

	basePrompt := buildQuotedPrompt(sourceMessages, trimmedUserContent)
	composedPrompt, attachments, err := h.buildAgentPrompt(userID, task.ID, session, basePrompt, input.AttachmentIDs)
	if err != nil {
		h.logFailure(request, "message_quote_build_prompt", err, "userId", userID, "taskId", task.ID, "sessionId", session.ID)
		WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

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

	userMessage, assistantMessage, err := h.Store.CreateMessagePair(userID, task.ID, session.ID, trimmedUserContent, assistantContent, nil, attachments)
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

	if action != "reply" && action != "regenerate" && action != "pin" {
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

	if action == "pin" {
		h.handleMessagePin(writer, request, userID, messageID)
		return
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

	basePrompt, replyToMessageID := buildReplyPrompt(sourceMessage, trimmedUserContent)
	composedPrompt, attachments, err := h.buildAgentPrompt(userID, task.ID, session, basePrompt, input.AttachmentIDs)
	if err != nil {
		h.logFailure(request, "message_reply_build_prompt", err, "userId", userID, "taskId", task.ID, "sessionId", session.ID, "sourceMessageId", sourceMessage.ID)
		WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

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

	userMessage, assistantMessage, err := h.Store.CreateMessagePair(userID, task.ID, session.ID, trimmedUserContent, assistantContent, replyToMessageID, attachments)
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

	composedPrompt, _, err := h.buildAgentPrompt(userID, task.ID, session, buildRegeneratePrompt(sourceMessage), nil)
	if err != nil {
		h.logFailure(request, "message_regenerate_build_prompt", err, "userId", userID, "taskId", task.ID, "sessionId", session.ID, "sourceMessageId", sourceMessage.ID)
		WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

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
 * handleMessagePin toggles one message into or out of the session pin set.
 * Params:
 * - writer: the HTTP response writer.
 * - request: the incoming HTTP request that triggered the pin action.
 * - userID: the authenticated user identifier from the bearer token.
 * - messageID: the selected message identifier from the request path.
 */
func (h *Handlers) handleMessagePin(writer http.ResponseWriter, request *http.Request, userID string, messageID string) {
	isPinned := request.Method == http.MethodPost
	updatedMessage, err := h.Store.SetMessagePin(userID, messageID, isPinned)
	if err != nil {
		h.logFailure(request, "message_pin_toggle", err, "userId", userID, "messageId", messageID, "isPinned", isPinned)
		WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(writer, http.StatusOK, updatedMessage)
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
 * buildAgentPrompt injects pinned context and uploaded attachment paths ahead of the base user prompt.
 * Params:
 * - userID: the authenticated user identifier from the bearer token.
 * - taskID: the active task identifier.
 * - session: the active session that owns both runtime context and draft attachments.
 * - basePrompt: the already-constructed business prompt for normal send, quote, reply, or regenerate.
 * - attachmentIDs: the uploaded draft attachment identifiers selected for this message send.
 * Returns:
 * - the final agent prompt and the resolved attachment records that should bind to the user message.
 */
func (h *Handlers) buildAgentPrompt(
	userID string,
	taskID string,
	session domain.Session,
	basePrompt string,
	attachmentIDs []string,
) (string, []domain.Attachment, error) {
	promptSegments := make([]string, 0, 3)

	pinnedMessages, err := h.Store.ListPinnedMessages(userID, session.ID)
	if err != nil {
		return "", nil, err
	}
	if len(pinnedMessages) > 0 {
		pinnedLines := make([]string, 0, len(pinnedMessages))
		for index, pinnedMessage := range pinnedMessages {
			pinnedLines = append(pinnedLines, fmt.Sprintf("重要信息%d：%s", index+1, pinnedMessage.Content))
		}
		promptSegments = append(promptSegments, strings.Join(pinnedLines, "\n"))
	}

	attachments, err := h.Store.GetDraftAttachments(userID, taskID, session.ID, attachmentIDs)
	if err != nil {
		return "", nil, err
	}
	if len(attachments) > 0 {
		attachmentLines := make([]string, 0, len(attachments))
		for _, attachment := range attachments {
			if attachment.SourceType == "image" {
				attachmentLines = append(attachmentLines, fmt.Sprintf("请查看图片 %s，路径：%s", attachment.FileName, attachment.StorageKey))
				continue
			}
			attachmentLines = append(attachmentLines, fmt.Sprintf("请阅读文件 %s，路径：%s", attachment.FileName, attachment.StorageKey))
		}
		promptSegments = append(promptSegments, fmt.Sprintf("以下是本轮新上传的资料，请先读取这些路径对应的内容：\n%s", strings.Join(attachmentLines, "\n")))
	}

	promptSegments = append(promptSegments, basePrompt)
	return strings.Join(promptSegments, "\n\n"), attachments, nil
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
 * parseFileRoute extracts the attachment identifier from /api/files/{attachmentId}.
 * Params:
 * - path: the request path that starts with /api/files/.
 */
func parseFileRoute(path string) (string, error) {
	trimmed := strings.TrimPrefix(path, "/api/files/")
	attachmentID := strings.Trim(trimmed, "/")
	if attachmentID == "" {
		return "", errorsNew("attachment id is required")
	}
	return attachmentID, nil
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

/**
 * normalizeAttachmentSourceType validates the front-end upload kind into the storage-friendly enum.
 * Params:
 * - value: the sourceType field submitted by the multipart upload request.
 */
func normalizeAttachmentSourceType(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	switch normalized {
	case "image", "file":
		return normalized
	default:
		return ""
	}
}

/**
 * sanitizeUploadedFileName removes path segments from one uploaded file name before local storage.
 * Params:
 * - fileName: the original file name submitted by the browser.
 */
func sanitizeUploadedFileName(fileName string) string {
	baseName := strings.TrimSpace(filepath.Base(fileName))
	baseName = strings.ReplaceAll(baseName, "/", "-")
	baseName = strings.ReplaceAll(baseName, "\\", "-")
	if baseName == "" || baseName == "." {
		return "upload"
	}
	return baseName
}

/**
 * buildStoredFileName creates a unique on-disk file name while preserving the original extension.
 * Params:
 * - fileName: the sanitized display file name selected by the user.
 */
func buildStoredFileName(fileName string) string {
	extension := filepath.Ext(fileName)
	baseName := strings.TrimSuffix(fileName, extension)
	safeBaseName := strings.ReplaceAll(baseName, " ", "_")
	return fmt.Sprintf("%s-%d%s", safeBaseName, time.Now().UnixNano(), extension)
}

/**
 * detectUploadedFileType resolves the most reliable MIME type for one uploaded browser file.
 * Params:
 * - file: the opened multipart file used for content sniffing.
 * - fallbackType: the MIME type reported by the browser headers when available.
 */
func detectUploadedFileType(file multipart.File, fallbackType string) (string, error) {
	sniffBuffer := make([]byte, 512)
	readBytes, err := file.Read(sniffBuffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	detectedType := strings.TrimSpace(fallbackType)
	if detectedType == "" || detectedType == "application/octet-stream" {
		detectedType = http.DetectContentType(sniffBuffer[:readBytes])
	}
	if detectedType == "" {
		detectedType = "application/octet-stream"
	}

	return detectedType, nil
}
