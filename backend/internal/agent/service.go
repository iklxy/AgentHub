// Date: 2026-05-25
// Author: XinYang Li

package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/lixinyang/agenthub/backend/internal/domain"
)

// Service wraps the Python agent entrypoint for session-aware execution.
type Service struct {
	logger      *slog.Logger
	codeWorkDir string
	harnessRoot string
	pythonPath  string
	pythonBin   string
	timeout     time.Duration
}

type harnessMeta struct {
	TaskID           string `json:"taskId"`
	TaskTitle        string `json:"taskTitle"`
	SessionID        string `json:"sessionId"`
	SessionTitle     string `json:"sessionTitle"`
	AgentID          string `json:"agentId"`
	AgentName        string `json:"agentName"`
	RuntimeProvider  string `json:"runtimeProvider"`
	RuntimeSessionID string `json:"runtimeSessionId"`
	UpdatedAt        string `json:"updatedAt"`
}

// ErrRuntimeNotImplemented marks an agent runtime target that is not yet supported by the current backend.
var ErrRuntimeNotImplemented = errors.New("agent runtime not implemented")

/**
 * NewService creates the Python agent execution wrapper.
 * Params:
 * - logger: the shared backend logger used for execution diagnostics.
 * - codeWorkDir: the Python project directory used as the subprocess working directory.
 * - harnessRoot: the local .AgentHub runtime root that stores per-agent harness files.
 * - pythonPath: the module search path passed through PYTHONPATH.
 * - pythonBin: the Python executable used to start the agent entrypoint.
 */
func NewService(logger *slog.Logger, codeWorkDir string, harnessRoot string, pythonPath string, pythonBin string) *Service {
	return &Service{
		logger:      logger.With("service", "agent-bridge"),
		codeWorkDir: codeWorkDir,
		harnessRoot: harnessRoot,
		pythonPath:  pythonPath,
		pythonBin:   pythonBin,
		timeout:     90 * time.Second,
	}
}

/**
 * RunSessionAgent executes the Python agent entrypoint and returns the assistant output.
 * Params:
 * - task: the task metadata used to build the execution context.
 * - session: the active task session bound to the current chat round.
 * - userInput: the latest user message content.
 */
func (s *Service) RunSessionAgent(task domain.Task, session domain.Session, userInput string) (string, error) {
	if !supportsClaudeCodeRuntime(session.RuntimeProvider) {
		return "", fmt.Errorf("%w for %s", ErrRuntimeNotImplemented, session.RuntimeProvider)
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	commandMode := "resume"
	if strings.TrimSpace(session.StartedAt) == "" {
		commandMode = "start"
	}

	runtimeSessionID := strings.TrimSpace(session.RuntimeSessionKey)
	if runtimeSessionID == "" {
		runtimeSessionID = session.ID
	}

	agentWorkDir, ruleFile, err := s.ensureHarness(session, task, runtimeSessionID)
	if err != nil {
		return "", err
	}

	cmd := exec.CommandContext(
		ctx,
		s.pythonBin,
		"-m",
		"agenthub_agent.main",
		"--agent-name",
		session.PrimaryAgentName,
		"--session-id",
		runtimeSessionID,
		"--agent-workdir",
		agentWorkDir,
		"--agent-rule-file",
		ruleFile,
		"--task-title",
		task.Title,
		"--task-description",
		task.Description,
		"--session-title",
		session.Title,
		"--command-mode",
		commandMode,
		"--user-input",
		userInput,
	)
	cmd.Dir = s.codeWorkDir
	cmd.Env = append(os.Environ(), fmt.Sprintf("PYTHONPATH=%s", s.pythonPath))

	startedAt := time.Now()
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		s.logger.Error(
			"agent process failed",
			"taskId", task.ID,
			"sessionId", session.ID,
			"agentId", session.PrimaryAgentID,
			"agentName", session.PrimaryAgentName,
			"runtimeProvider", session.RuntimeProvider,
			"runtimeMode", commandMode,
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
		"sessionId", session.ID,
		"agentId", session.PrimaryAgentID,
		"agentName", session.PrimaryAgentName,
		"runtimeProvider", session.RuntimeProvider,
		"runtimeMode", commandMode,
		"durationMs", time.Since(startedAt).Milliseconds(),
	)
	return result, nil
}

/**
 * supportsClaudeCodeRuntime checks whether one session runtime provider should be executed by the Claude Code CLI bridge.
 * Params:
 * - runtimeProvider: the provider_type value stored on the bound agent and copied onto the task session.
 */
func supportsClaudeCodeRuntime(runtimeProvider string) bool {
	provider := strings.TrimSpace(strings.ToLower(runtimeProvider))
	return provider == "claude_code" || provider == "claude-code"
}

/**
 * ResolveAgentPaths builds the Python code directory, PYTHONPATH, and local harness root.
 * Params:
 * - rootDir: the root path of the Python agent folder inside the repository.
 */
func ResolveAgentPaths(rootDir string) (string, string, string) {
	repoRoot := filepath.Dir(rootDir)
	return rootDir, filepath.Join(rootDir, "src"), filepath.Join(repoRoot, ".AgentHub")
}

/**
 * EnsureSessionAssetDir resolves and creates the per-session attachment directory for one agent.
 * Params:
 * - session: the active task session that owns the uploaded assets.
 * - sourceType: the logical attachment kind, currently image or file.
 * Returns:
 * - the absolute directory path used to persist uploaded assets.
 */
func (s *Service) EnsureSessionAssetDir(session domain.Session, sourceType string) (string, error) {
	_, sessionDir, _, _, err := s.ensureHarnessPaths(session)
	if err != nil {
		return "", err
	}

	directoryName := "file"
	if strings.EqualFold(strings.TrimSpace(sourceType), "image") {
		directoryName = "photo"
	}

	assetDir := filepath.Join(sessionDir, directoryName)
	if err := os.MkdirAll(assetDir, 0o755); err != nil {
		return "", err
	}

	return assetDir, nil
}

func (s *Service) ensureHarness(session domain.Session, task domain.Task, runtimeSessionID string) (string, string, error) {
	agentDir, sessionsDir, _, workDir, err := s.ensureHarnessPaths(session)
	if err != nil {
		return "", "", err
	}
	ruleFile := filepath.Join(agentDir, "Agent.md")

	if _, err := os.Stat(ruleFile); os.IsNotExist(err) {
		defaultRule := fmt.Sprintf("# %s Agent Rules\n\n- Stay focused on the current task.\n- Reply clearly and directly.\n", session.PrimaryAgentName)
		if writeErr := os.WriteFile(ruleFile, []byte(defaultRule), 0o644); writeErr != nil {
			return "", "", writeErr
		}
	}

	meta := harnessMeta{
		TaskID:           task.ID,
		TaskTitle:        task.Title,
		SessionID:        session.ID,
		SessionTitle:     session.Title,
		AgentID:          session.PrimaryAgentID,
		AgentName:        session.PrimaryAgentName,
		RuntimeProvider:  session.RuntimeProvider,
		RuntimeSessionID: runtimeSessionID,
		UpdatedAt:        time.Now().UTC().Format(time.RFC3339),
	}

	metaBytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return "", "", err
	}

	if err := os.WriteFile(filepath.Join(sessionsDir, "meta.json"), metaBytes, 0o644); err != nil {
		return "", "", err
	}

	return workDir, ruleFile, nil
}

/**
 * ensureHarnessPaths resolves and creates the standard agent/session directory tree.
 * Params:
 * - session: the active task session used to locate the agent harness subtree.
 * Returns:
 * - the agent root, session root under workspace/session/{session_id}, logs directory, and shared workspace directory.
 */
func (s *Service) ensureHarnessPaths(session domain.Session) (string, string, string, string, error) {
	agentDir := filepath.Join(s.harnessRoot, fmt.Sprintf(".%s", session.PrimaryAgentName))
	workDir := filepath.Join(agentDir, "workspace")
	sessionsDir := filepath.Join(workDir, "session", session.ID)
	logsDir := filepath.Join(sessionsDir, "logs")

	for _, path := range []string{s.harnessRoot, agentDir, sessionsDir, logsDir, workDir} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return "", "", "", "", err
		}
	}

	return agentDir, sessionsDir, logsDir, workDir, nil
}
