// Date: 2026-05-25
// Author: XinYang Li

package httpx

import (
	"log/slog"
	"net/http"
)

/**
 * NewRouter wires the v0.1 backend routes to their handlers.
 * Params:
 * - logger: the shared backend logger used for request logging middleware.
 * - handlers: the HTTP handlers with store and logger dependencies attached.
 */
func NewRouter(logger *slog.Logger, handlers *Handlers) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", handlers.Health)
	mux.HandleFunc("POST /api/register", handlers.Register)
	mux.HandleFunc("POST /api/login", handlers.Login)
	mux.HandleFunc("GET /api/me", handlers.Me)
	mux.HandleFunc("GET /api/workspace", handlers.GetWorkspace)
	mux.HandleFunc("PATCH /api/workspace", handlers.UpdateWorkspace)
	mux.HandleFunc("GET /api/tasks", handlers.ListTasks)
	mux.HandleFunc("POST /api/tasks", handlers.CreateTask)
	mux.HandleFunc("GET /api/tasks/", routeTaskSubresources(handlers))
	mux.HandleFunc("POST /api/tasks/", routeTaskSubresources(handlers))

	return RequestLogging(logger, mux)
}

/**
 * routeTaskSubresources dispatches /api/tasks/{taskId}/... requests by method and suffix.
 * Params:
 * - handlers: the shared handler collection used for the subresource routes.
 */
func routeTaskSubresources(handlers *Handlers) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/api/tasks/":
			WriteError(writer, http.StatusBadRequest, "task id is required")
		case request.Method == http.MethodGet && hasSuffix(request.URL.Path, "/conversations"):
			handlers.GetConversations(writer, request)
		case request.Method == http.MethodGet && hasSuffix(request.URL.Path, "/messages"):
			handlers.GetMessages(writer, request)
		case request.Method == http.MethodGet:
			handlers.GetTask(writer, request)
		case request.Method == http.MethodPost && hasSuffix(request.URL.Path, "/messages"):
			handlers.CreateMessage(writer, request)
		default:
			WriteError(writer, http.StatusMethodNotAllowed, "method not allowed")
		}
	}
}

/**
 * hasSuffix checks whether a route path ends with the expected suffix.
 * Params:
 * - path: the incoming request path.
 * - suffix: the expected trailing route suffix.
 */
func hasSuffix(path string, suffix string) bool {
	return len(path) >= len(suffix) && path[len(path)-len(suffix):] == suffix
}
