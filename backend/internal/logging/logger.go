// Date: 2026-05-25
// Author: XinYang Li

package logging

import (
	"log/slog"
	"os"
)

/**
 * New creates the shared structured JSON logger for the backend service.
 * Params:
 * - none: the logger writes JSON logs to stdout.
 */
func New() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	return slog.New(handler).With("service", "backend")
}
