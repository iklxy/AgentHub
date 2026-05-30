// Date: 2026-05-25
// Author: XinYang Li

package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

// AgentSession manages a running Python agent process with bidirectional JSONL communication.
type AgentSession struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Scanner
	logger *slog.Logger
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
 * StartAgent launches the Python agent process with bidirectional pipes for permission approval.
 * Params:
 * - task: the task metadata used to build the execution context.
 * - session: the active task session bound to the current chat round.
 * - userInput: the latest user message content.
 * Returns:
 * - an AgentSession for reading messages and writing responses.
 */
func (s *Service) StartAgent(task domain.Task, session domain.Session, userInput string) (*AgentSession, error) {
	if !supportsClaudeCodeRuntime(session.RuntimeProvider) {
		return nil, fmt.Errorf("%w for %s", ErrRuntimeNotImplemented, session.RuntimeProvider)
	}

	resumeSessionID := strings.TrimSpace(session.RuntimeSessionKey)
	runtimeSessionID := resumeSessionID
	if runtimeSessionID == "" {
		runtimeSessionID = session.ID
	}

	agentWorkDir, ruleFile, err := s.ensureHarness(session, task, runtimeSessionID)
	if err != nil {
		return nil, err
	}

	args := []string{
		"-m", "agenthub_agent.main",
		"--agent-name", session.PrimaryAgentName,
		"--session-id", runtimeSessionID,
		"--agent-workdir", agentWorkDir,
		"--agent-rule-file", ruleFile,
		"--task-title", task.Title,
		"--task-description", task.Description,
		"--session-title", session.Title,
		"--user-input", userInput,
	}

	if resumeSessionID != "" {
		args = append(args, "--resume-session-id", resumeSessionID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	cmd := exec.CommandContext(ctx, s.pythonBin, args...)
	cmd.Dir = s.codeWorkDir
	cmd.Env = append(os.Environ(), fmt.Sprintf("PYTHONPATH=%s", s.pythonPath))

	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start agent process: %w", err)
	}

	// Cancel context on process exit so pipes clean up.
	go func() {
		_ = cmd.Wait()
		cancel()
	}()

	return &AgentSession{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewScanner(stdout),
		logger: s.logger,
	}, nil
}

/**
 * ReadMessage reads the next JSONL message from the agent's stdout.
 * Returns nil and an error when the stream ends or a read error occurs.
 */
func (a *AgentSession) ReadMessage() (*AgentMessage, error) {
	if !a.stdout.Scan() {
		if err := a.stdout.Err(); err != nil {
			return nil, fmt.Errorf("stdout read error: %w", err)
		}
		return nil, io.EOF
	}

	line := a.stdout.Bytes()
	var msg AgentMessage
	if err := json.Unmarshal(line, &msg); err != nil {
		a.logger.Error("failed to parse agent message", "line", string(line), "error", err)
		return nil, fmt.Errorf("failed to parse agent message: %w", err)
	}

	return &msg, nil
}

/**
 * WriteResponse writes a permission response as JSONL to the agent's stdin.
 */
func (a *AgentSession) WriteResponse(response PermissionResponse) error {
	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	data = append(data, '\n')
	if _, err := a.stdin.Write(data); err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}

	return nil
}

/**
 * Close terminates the agent process and cleans up resources.
 */
func (a *AgentSession) Close() error {
	if a.stdin != nil {
		_ = a.stdin.Close()
	}
	if a.cmd != nil && a.cmd.Process != nil {
		return a.cmd.Process.Kill()
	}
	return nil
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
