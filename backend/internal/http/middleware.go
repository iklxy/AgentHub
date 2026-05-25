// Date: 2026-05-25
// Author: XinYang Li

package httpx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const requestIDContextKey contextKey = "requestID"

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

/**
 * WriteHeader captures the response status code before forwarding it.
 * Params:
 * - statusCode: the HTTP status code written by the downstream handler.
 */
func (recorder *statusRecorder) WriteHeader(statusCode int) {
	recorder.statusCode = statusCode
	recorder.ResponseWriter.WriteHeader(statusCode)
}

/**
 * RequestLogging wraps the router with structured request timing logs.
 * Params:
 * - logger: the shared backend logger used for request diagnostics.
 * - next: the downstream HTTP handler being wrapped.
 */
func RequestLogging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestID := generateRequestID()
		startedAt := time.Now()
		recorder := &statusRecorder{
			ResponseWriter: writer,
			statusCode:     http.StatusOK,
		}
		recorder.Header().Set("X-Request-Id", requestID)

		requestWithContext := request.WithContext(context.WithValue(request.Context(), requestIDContextKey, requestID))
		next.ServeHTTP(recorder, requestWithContext)

		logger.Info(
			"http request completed",
			"requestId", requestID,
			"method", request.Method,
			"path", request.URL.Path,
			"statusCode", recorder.statusCode,
			"durationMs", time.Since(startedAt).Milliseconds(),
			"origin", request.Header.Get("Origin"),
			"userAgent", request.UserAgent(),
			"remoteAddr", request.RemoteAddr,
		)
	})
}

/**
 * CORS allows the front-end development origin to call the Go API directly.
 * Params:
 * - allowedOrigin: the exact front-end origin allowed to call the API.
 * - next: the downstream HTTP handler wrapped by the CORS layer.
 */
func CORS(allowedOrigin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")

		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(writer, request)
	})
}

func generateRequestID() string {
	buffer := make([]byte, 8)
	_, _ = rand.Read(buffer)
	return hex.EncodeToString(buffer)
}

func getRequestID(request *http.Request) string {
	value := request.Context().Value(requestIDContextKey)
	requestID, ok := value.(string)
	if !ok {
		return ""
	}
	return requestID
}

/**
 * getAuthenticatedUserID extracts the bearer token and treats it as the current user identifier.
 * Params:
 * - request: the incoming HTTP request that may contain the Authorization header.
 */
func getAuthenticatedUserID(request *http.Request) (string, error) {
	authorization := strings.TrimSpace(request.Header.Get("Authorization"))
	if authorization == "" {
		return "", errors.New("missing authorization header")
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(authorization, prefix) {
		return "", errors.New("invalid authorization header")
	}

	userID := strings.TrimSpace(strings.TrimPrefix(authorization, prefix))
	if userID == "" {
		return "", errors.New("missing bearer token")
	}

	return userID, nil
}
