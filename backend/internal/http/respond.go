// Date: 2026-05-25
// Author: XinYang Li

package httpx

import (
	"encoding/json"
	"net/http"
)

/**
 * WriteJSON writes a JSON response body with the provided status code.
 * Params:
 * - writer: the response writer used by the current HTTP handler.
 * - statusCode: the HTTP status code applied to the response.
 * - payload: the serializable response body.
 */
func WriteJSON(writer http.ResponseWriter, statusCode int, payload any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	_ = json.NewEncoder(writer).Encode(payload)
}

/**
 * WriteError writes a stable JSON error payload.
 * Params:
 * - writer: the response writer used by the current HTTP handler.
 * - statusCode: the HTTP status code applied to the response.
 * - message: the human-readable error summary.
 */
func WriteError(writer http.ResponseWriter, statusCode int, message string) {
	WriteJSON(writer, statusCode, map[string]string{
		"error": message,
	})
}
