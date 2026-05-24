// Date: 2026-05-25
// Author: XinYang Li

package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/lixinyang/agenthub/backend/internal/app"
)

/**
 * Starts the AgentHub backend server process.
 * Params:
 * - none: the server reads configuration from environment variables.
 */
func main() {
	serverApp, err := app.New()
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}
	defer func() {
		if closeErr := serverApp.Store.Close(); closeErr != nil {
			serverApp.Logger.Error("backend store close failed", "error", closeErr)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := serverApp.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverApp.Logger.Error("backend server stopped unexpectedly", "error", err)
		}
	}()

	serverApp.Logger.Info("backend server started", "address", serverApp.Server.Addr)
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := serverApp.Server.Shutdown(shutdownCtx); err != nil {
		serverApp.Logger.Error("backend server shutdown failed", "error", err)
		return
	}

	serverApp.Logger.Info("backend server stopped cleanly")
}
