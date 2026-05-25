// Date: 2026-05-25
// Author: XinYang Li

package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/lixinyang/agenthub/backend/internal/agent"
	"github.com/lixinyang/agenthub/backend/internal/config"
	httpx "github.com/lixinyang/agenthub/backend/internal/http"
	"github.com/lixinyang/agenthub/backend/internal/logging"
	"github.com/lixinyang/agenthub/backend/internal/store/postgres"
)

// App groups the constructed runtime dependencies for the backend server.
type App struct {
	Logger *slog.Logger
	Server *http.Server
	Store  *postgres.Store
}

/**
 * New constructs the backend application with configuration, store, router, and logger.
 * Params:
 * - none: the application reads configuration internally and wires default dependencies.
 */
func New() (*App, error) {
	cfg := config.Load()
	logger := logging.New()

	store, err := postgres.NewStore(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	agentCodeDir, pythonPath, harnessRoot := agent.ResolveAgentPaths(cfg.AgentRootDir)
	agentService := agent.NewService(logger, agentCodeDir, harnessRoot, pythonPath, cfg.PythonBin)
	handlers := &httpx.Handlers{
		Logger:       logger,
		Store:        store,
		AgentService: agentService,
	}

	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.Port),
		Handler:           httpx.CORS(cfg.FrontendOrigin, httpx.NewRouter(logger, handlers)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{
		Logger: logger,
		Server: server,
		Store:  store,
	}, nil
}
