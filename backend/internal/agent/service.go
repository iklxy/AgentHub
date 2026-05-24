// Date: 2026-05-25
// Author: XinYang Li

package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/lixinyang/agenthub/backend/internal/domain"
)

// Service wraps the Python agent entrypoint for main-agent execution.
type Service struct {
	logger     *slog.Logger
	workDir    string
	pythonPath string
	pythonBin  string
	timeout    time.Duration
}

/**
 * NewService creates the Python agent execution wrapper.
 * Params:
 * - logger: the shared backend logger used for execution diagnostics.
 * - workDir: the agent project directory used as the subprocess working directory.
 * - pythonPath: the module search path passed through PYTHONPATH.
 * - pythonBin: the Python executable used to start the agent entrypoint.
 */
func NewService(logger *slog.Logger, workDir string, pythonPath string, pythonBin string) *Service {
	return &Service{
		logger:     logger.With("service", "agent-bridge"),
		workDir:    workDir,
		pythonPath: pythonPath,
		pythonBin:  pythonBin,
		timeout:    90 * time.Second,
	}
}

/**
 * RunMainAgent executes the Python agent entrypoint and returns the assistant output.
 * Params:
 * - task: the task metadata used to build the execution context.
 * - history: the ordered task message history passed into the prompt builder.
 * - userInput: the latest user message content.
 */
func (s *Service) RunMainAgent(task domain.Task, history []domain.Message, userInput string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	historyPayload := []map[string]string{}
	for _, message := range history {
		historyPayload = append(historyPayload, map[string]string{
			"role":    message.Role,
			"content": message.Content,
		})
	}

	rawHistory, err := json.Marshal(historyPayload)
	if err != nil {
		return "", err
	}

	cmd := exec.CommandContext(
		ctx,
		s.pythonBin,
		"-m",
		"agenthub_agent.main",
		"--task-title",
		task.Title,
		"--task-description",
		task.Description,
		"--user-input",
		userInput,
		"--history-json",
		string(rawHistory),
	)
	cmd.Dir = s.workDir
	cmd.Env = append(os.Environ(), fmt.Sprintf("PYTHONPATH=%s", s.pythonPath))

	startedAt := time.Now()
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		s.logger.Error(
			"agent process failed",
			"taskId", task.ID,
			"durationMs", time.Since(startedAt).Milliseconds(),
			"error", err,
			"stderr", stderr.String(),
		)
		return "", err
	}

	result := strings.TrimSpace(string(output))
	s.logger.Info(
		"agent process completed",
		"taskId", task.ID,
		"durationMs", time.Since(startedAt).Milliseconds(),
	)
	return result, nil
}

/**
 * ResolveAgentPaths builds the agent work directory and PYTHONPATH from the configured root.
 * Params:
 * - rootDir: the root path of the agent folder inside the repository.
 */
func ResolveAgentPaths(rootDir string) (string, string) {
	return rootDir, filepath.Join(rootDir, "src")
}
