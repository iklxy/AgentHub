// Date: 2026-05-25
// Author: XinYang Li

package postgres

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/lixinyang/agenthub/backend/internal/domain"
)

type runtimeAgent struct {
	ID           string
	Name         string
	Kind         string
	ProviderType string
}

// Store persists AgentHub v0.2 data in PostgreSQL.
type Store struct {
	db *sql.DB
}

/**
 * NewStore opens a PostgreSQL-backed repository and performs the minimum compatibility checks.
 * Params:
 * - connectionString: the PostgreSQL connection string used to open the database handle.
 */
func NewStore(connectionString string) (*Store, error) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	store := &Store{db: db}
	if err := store.ensureCompatibility(context.Background()); err != nil {
		return nil, err
	}

	return store, nil
}

/**
 * Close releases the PostgreSQL database handle.
 * Params:
 * - none: the store closes its owned database handle.
 */
func (s *Store) Close() error {
	return s.db.Close()
}

/**
 * CreateUser inserts a new user, a default workspace, and the membership row needed by the current flow.
 * Params:
 * - input: the registration payload with username, email, and password.
 */
func (s *Store) CreateUser(input domain.RegisterRequest) (domain.User, error) {
	ctx := context.Background()
	email := strings.TrimSpace(strings.ToLower(input.Email))
	now := time.Now().UTC()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.User{}, err
	}
	defer tx.Rollback()

	userID := generateUUID()
	passwordHash := hashPassword(input.Password)

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO users (id, email, name, avatar_url, role, status, password_hash, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		userID,
		email,
		input.Username,
		nil,
		"member",
		"active",
		passwordHash,
		now,
		now,
	)
	if err != nil {
		return domain.User{}, err
	}

	workspaceID := generateUUID()
	workspaceName := fmt.Sprintf("%s 的工作区", input.Username)
	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO workspaces (id, name, description, status, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		workspaceID,
		workspaceName,
		"写一段描述，让这个工作区更容易识别",
		"active",
		userID,
		now,
		now,
	)
	if err != nil {
		return domain.User{}, err
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO workspace_members (id, workspace_id, user_id, member_role, joined_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		generateUUID(),
		workspaceID,
		userID,
		"owner",
		now,
	)
	if err != nil {
		return domain.User{}, err
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO agents (
		    id, workspace_id, name, kind, status, description, capability_tags,
		    source_kind, provider_type, tool_schema_json, created_by, created_at, updated_at
		  )
		  VALUES ($1, $2, $3, $4, $5, $6, '[]'::jsonb, $7, $8, '{}'::jsonb, $9, $10, $11)`,
		generateUUID(),
		workspaceID,
		"Galaxy",
		"main",
		"active",
		"默认主 Agent，负责当前工作区下 task 的主会话协作",
		"external_cli",
		"claude_code",
		userID,
		now,
		now,
	)
	if err != nil {
		return domain.User{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.User{}, err
	}

	return domain.User{
		ID:        userID,
		Username:  input.Username,
		Email:     email,
		AvatarURL: "",
	}, nil
}

/**
 * AuthenticateUser verifies the login payload against stored credentials.
 * Params:
 * - input: the login payload with email and password.
 */
func (s *Store) AuthenticateUser(input domain.LoginRequest) (domain.User, error) {
	ctx := context.Background()
	var user domain.User
	var passwordHash sql.NullString
	var avatarURL sql.NullString

	err := s.db.QueryRowContext(
		ctx,
		`SELECT id, name, email, avatar_url, password_hash
		 FROM users
		 WHERE email = $1`,
		strings.TrimSpace(strings.ToLower(input.Email)),
	).Scan(&user.ID, &user.Username, &user.Email, &avatarURL, &passwordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, errors.New("user not found")
		}
		return domain.User{}, err
	}

	if !passwordHash.Valid || passwordHash.String != hashPassword(input.Password) {
		return domain.User{}, errors.New("invalid password")
	}

	user.AvatarURL = avatarURL.String
	return user, nil
}

/**
 * GetUserByID returns one user by identifier.
 * Params:
 * - userID: the authenticated user identifier from the bearer token.
 */
func (s *Store) GetUserByID(userID string) (domain.User, error) {
	ctx := context.Background()
	var user domain.User
	var avatarURL sql.NullString

	err := s.db.QueryRowContext(
		ctx,
		`SELECT id, name, email, avatar_url
		 FROM users
		 WHERE id = $1`,
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &avatarURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, errors.New("user not found")
		}
		return domain.User{}, err
	}

	user.AvatarURL = avatarURL.String
	return user, nil
}

/**
 * GetWorkspaceByUserID returns the first active workspace the user is a member of.
 * Params:
 * - userID: the authenticated user identifier from the bearer token.
 */
func (s *Store) GetWorkspaceByUserID(userID string) (domain.Workspace, error) {
	ctx := context.Background()
	var workspace domain.Workspace

	err := s.db.QueryRowContext(
		ctx,
		`SELECT w.id, wm.user_id, w.name, COALESCE(w.description, '')
		 FROM workspace_members wm
		 INNER JOIN workspaces w ON w.id = wm.workspace_id
		 WHERE wm.user_id = $1 AND w.status = 'active'
		 ORDER BY w.created_at ASC
		 LIMIT 1`,
		userID,
	).Scan(&workspace.ID, &workspace.UserID, &workspace.Name, &workspace.Description)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Workspace{}, errors.New("workspace not found")
		}
		return domain.Workspace{}, err
	}

	return workspace, nil
}

/**
 * UpdateWorkspace mutates workspace metadata for the first active workspace the user can access.
 * Params:
 * - userID: the authenticated user identifier from the bearer token.
 * - input: the new workspace name and description values.
 */
func (s *Store) UpdateWorkspace(userID string, input domain.UpdateWorkspaceRequest) (domain.Workspace, error) {
	ctx := context.Background()
	workspace, err := s.GetWorkspaceByUserID(userID)
	if err != nil {
		return domain.Workspace{}, err
	}

	err = s.db.QueryRowContext(
		ctx,
		`UPDATE workspaces
		 SET name = $2, description = $3, updated_at = $4
		 WHERE id = $1
		 RETURNING id, name, COALESCE(description, '')`,
		workspace.ID,
		input.Name,
		input.Description,
		time.Now().UTC(),
	).Scan(&workspace.ID, &workspace.Name, &workspace.Description)
	if err != nil {
		return domain.Workspace{}, err
	}

	return workspace, nil
}

/**
 * ListTasks returns tasks for the authenticated user's first active workspace.
 * Params:
 * - userID: the authenticated user identifier from the bearer token.
 */
func (s *Store) ListTasks(userID string) ([]domain.Task, error) {
	ctx := context.Background()
	workspace, err := s.GetWorkspaceByUserID(userID)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, workspace_id, title, COALESCE(description, ''), status, COALESCE(current_session_id::text, ''), updated_at
		 FROM tasks
		 WHERE workspace_id = $1
		 ORDER BY updated_at DESC`,
		workspace.ID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []domain.Task{}
	for rows.Next() {
		var task domain.Task
		var rawStatus string
		var updatedAt time.Time
		if err := rows.Scan(&task.ID, &task.WorkspaceID, &task.Title, &task.Description, &rawStatus, &task.CurrentSessionID, &updatedAt); err != nil {
			return nil, err
		}
		task.Status = mapTaskStatus(rawStatus)
		task.UpdatedAtLabel = formatRelativeTime(updatedAt)
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

/**
 * CreateTask inserts a task and the default primary session for the authenticated user.
 * Params:
 * - userID: the authenticated user identifier from the bearer token.
 * - input: the task title and description values.
 */
func (s *Store) CreateTask(userID string, input domain.CreateTaskRequest) (domain.Task, error) {
	ctx := context.Background()
	workspace, err := s.GetWorkspaceByUserID(userID)
	if err != nil {
		return domain.Task{}, err
	}

	galaxyAgent, err := s.findAgentByName(ctx, workspace.ID, "Galaxy")
	if err != nil {
		return domain.Task{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Task{}, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	taskID := generateUUID()
	sessionID := generateUUID()

	task := domain.Task{
		ID:               taskID,
		WorkspaceID:      workspace.ID,
		Title:            input.Title,
		Description:      input.Description,
		Status:           "idle",
		CurrentSessionID: sessionID,
		UpdatedAtLabel:   "刚刚",
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO tasks (
		    id, workspace_id, title, description, status, current_session_id, current_primary_agent_id, created_by, created_at, updated_at
		  )
		  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		task.ID,
		task.WorkspaceID,
		task.Title,
		task.Description,
		"draft",
		sessionID,
		galaxyAgent.ID,
		userID,
		now,
		now,
	)
	if err != nil {
		return domain.Task{}, err
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO task_sessions (
		    id, workspace_id, task_id, title, chat_mode, status, is_pinned, last_active_at,
		    last_message_preview, primary_agent_id, created_by, created_at, updated_at,
		    session_kind, created_from_session_id, runtime_provider, runtime_session_key, started_at
		  )
		  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`,
		sessionID,
		task.WorkspaceID,
		task.ID,
		"主对话",
		"single",
		"active",
		false,
		now,
		"",
		galaxyAgent.ID,
		userID,
		now,
		now,
		"primary",
		nil,
		"claude_code",
		sessionID,
		nil,
	)
	if err != nil {
		return domain.Task{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Task{}, err
	}

	return task, nil
}

/**
 * GetTask returns one task inside the authenticated user's accessible workspace.
 * Params:
 * - userID: the authenticated user identifier from the bearer token.
 * - taskID: the task identifier from the request path.
 */
func (s *Store) GetTask(userID string, taskID string) (domain.Task, error) {
	ctx := context.Background()
	var task domain.Task
	var rawStatus string
	var currentSessionID sql.NullString
	var updatedAt time.Time

	err := s.db.QueryRowContext(
		ctx,
		`SELECT t.id, t.workspace_id, t.title, COALESCE(t.description, ''), t.status, t.current_session_id::text, t.updated_at
		 FROM tasks t
		 INNER JOIN workspace_members wm ON wm.workspace_id = t.workspace_id
		 WHERE wm.user_id = $1 AND t.id = $2
		 LIMIT 1`,
		userID,
		taskID,
	).Scan(&task.ID, &task.WorkspaceID, &task.Title, &task.Description, &rawStatus, &currentSessionID, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Task{}, errors.New("task not found")
		}
		return domain.Task{}, err
	}

	task.Status = mapTaskStatus(rawStatus)
	task.CurrentSessionID = currentSessionID.String
	task.UpdatedAtLabel = formatRelativeTime(updatedAt)
	return task, nil
}

/**
 * ListSessions returns the session list for one task.
 * Params:
 * - userID: the authenticated user identifier from the bearer token.
 * - taskID: the task identifier from the request path.
 */
func (s *Store) ListSessions(userID string, taskID string) ([]domain.Session, error) {
	ctx := context.Background()
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT
		    ts.id,
		    ts.task_id,
		    ts.title,
		    ts.chat_mode,
		    COALESCE(ts.session_kind, 'primary'),
		    COALESCE(ts.primary_agent_id::text, ''),
		    COALESCE(a.name, ''),
		    COALESCE(ts.runtime_provider, ''),
		    COALESCE(ts.runtime_session_key, ''),
		    COALESCE(ts.created_from_session_id::text, ''),
		    ts.started_at,
		    ts.last_active_at,
		    COALESCE(ts.last_message_preview, '')
		 FROM task_sessions ts
		 INNER JOIN workspace_members wm ON wm.workspace_id = ts.workspace_id
		 LEFT JOIN agents a ON a.id = ts.primary_agent_id
		 WHERE wm.user_id = $1 AND ts.task_id = $2 AND ts.status = 'active'
		 ORDER BY CASE WHEN COALESCE(ts.session_kind, 'primary') = 'primary' THEN 0 ELSE 1 END, ts.last_active_at DESC, ts.created_at ASC`,
		userID,
		taskID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := []domain.Session{}
	for rows.Next() {
		session, scanErr := scanSession(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

/**
 * GetSession returns one task session the current user can access.
 * Params:
 * - userID: the authenticated user identifier from the bearer token.
 * - sessionID: the session identifier from the request path or request body.
 */
func (s *Store) GetSession(userID string, sessionID string) (domain.Session, error) {
	ctx := context.Background()
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT
		    ts.id,
		    ts.task_id,
		    ts.title,
		    ts.chat_mode,
		    COALESCE(ts.session_kind, 'primary'),
		    COALESCE(ts.primary_agent_id::text, ''),
		    COALESCE(a.name, ''),
		    COALESCE(ts.runtime_provider, ''),
		    COALESCE(ts.runtime_session_key, ''),
		    COALESCE(ts.created_from_session_id::text, ''),
		    ts.started_at,
		    ts.last_active_at,
		    COALESCE(ts.last_message_preview, '')
		 FROM task_sessions ts
		 INNER JOIN workspace_members wm ON wm.workspace_id = ts.workspace_id
		 LEFT JOIN agents a ON a.id = ts.primary_agent_id
		 WHERE wm.user_id = $1 AND ts.id = $2
		 LIMIT 1`,
		userID,
		sessionID,
	)
	if err != nil {
		return domain.Session{}, err
	}
	defer rows.Close()

	if !rows.Next() {
		return domain.Session{}, errors.New("session not found")
	}

	session, err := scanSession(rows)
	if err != nil {
		return domain.Session{}, err
	}

	return session, nil
}

/**
 * CreateSession creates one branch session under a task for the authenticated user.
 * Params:
 * - userID: the authenticated user identifier from the bearer token.
 * - input: the session creation payload.
 */
func (s *Store) CreateSession(userID string, input domain.CreateSessionRequest) (domain.Session, error) {
	ctx := context.Background()
	if input.ChatMode != "single" {
		return domain.Session{}, errors.New("only single chat mode is supported in v0.2")
	}

	task, err := s.GetTask(userID, input.TaskID)
	if err != nil {
		return domain.Session{}, err
	}

	agentTarget, err := s.findAgentByID(ctx, task.WorkspaceID, input.PrimaryAgentID)
	if err != nil {
		return domain.Session{}, err
	}
	if strings.EqualFold(agentTarget.Name, "Galaxy") {
		return domain.Session{}, errors.New("Galaxy is reserved for the default primary session")
	}

	now := time.Now().UTC()
	sessionID := generateUUID()

	_, err = s.db.ExecContext(
		ctx,
		`INSERT INTO task_sessions (
		    id, workspace_id, task_id, title, chat_mode, status, is_pinned, last_active_at,
		    last_message_preview, primary_agent_id, created_by, created_at, updated_at,
		    session_kind, created_from_session_id, runtime_provider, runtime_session_key, started_at
		  )
		  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`,
		sessionID,
		task.WorkspaceID,
		task.ID,
		input.Title,
		input.ChatMode,
		"active",
		false,
		now,
		"",
		agentTarget.ID,
		userID,
		now,
		now,
		"branch",
		nil,
		agentTarget.ProviderType,
		sessionID,
		nil,
	)
	if err != nil {
		return domain.Session{}, err
	}

	return s.GetSession(userID, sessionID)
}

/**
 * UpdateSession updates the editable metadata of one session.
 * Params:
 * - userID: the authenticated user identifier from the bearer token.
 * - sessionID: the session identifier from the request path.
 * - input: the editable session update payload.
 */
func (s *Store) UpdateSession(userID string, sessionID string, input domain.UpdateSessionRequest) (domain.Session, error) {
	ctx := context.Background()
	session, err := s.GetSession(userID, sessionID)
	if err != nil {
		return domain.Session{}, err
	}

	_, err = s.db.ExecContext(
		ctx,
		`UPDATE task_sessions
		 SET title = $2, updated_at = $3
		 WHERE id = $1`,
		session.ID,
		input.Title,
		time.Now().UTC(),
	)
	if err != nil {
		return domain.Session{}, err
	}

	return s.GetSession(userID, sessionID)
}

/**
 * ListSessionAgents returns the enabled non-Galaxy agents that can be used for branch sessions.
 * Params:
 * - userID: the authenticated user identifier from the bearer token.
 * - taskID: the task identifier from the request path.
 */
func (s *Store) ListSessionAgents(userID string, taskID string) ([]domain.AgentOption, error) {
	ctx := context.Background()
	task, err := s.GetTask(userID, taskID)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, name, kind, provider_type
		 FROM agents
		 WHERE workspace_id = $1 AND status = 'active' AND LOWER(name) <> 'galaxy'
		 ORDER BY name ASC`,
		task.WorkspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	agents := []domain.AgentOption{}
	for rows.Next() {
		var item domain.AgentOption
		if err := rows.Scan(&item.ID, &item.Name, &item.Kind, &item.ProviderType); err != nil {
			return nil, err
		}
		agents = append(agents, item)
	}

	return agents, rows.Err()
}

/**
 * ListMessages returns the ordered transcript for one task session.
 * Params:
 * - userID: the authenticated user identifier from the bearer token.
 * - taskID: the task identifier from the request path.
 * - sessionID: the session identifier used by the current chat view.
 */
func (s *Store) ListMessages(userID string, taskID string, sessionID string) ([]domain.Message, error) {
	ctx := context.Background()
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT m.id, m.task_id, m.session_id, m.role, m.content_md, m.created_at
		 FROM messages m
		 INNER JOIN workspace_members wm ON wm.workspace_id = m.workspace_id
		 WHERE wm.user_id = $1 AND m.task_id = $2 AND m.session_id = $3
		 ORDER BY m.created_at ASC`,
		userID,
		taskID,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []domain.Message{}
	for rows.Next() {
		var message domain.Message
		var createdAt time.Time
		if err := rows.Scan(&message.ID, &message.TaskID, &message.SessionID, &message.Role, &message.Content, &createdAt); err != nil {
			return nil, err
		}
		message.Status = "success"
		message.TimeLabel = createdAt.Local().Format("15:04")
		messages = append(messages, message)
	}

	return messages, rows.Err()
}

/**
 * CreateMessagePair stores the user message and assistant reply for one task session round.
 * Params:
 * - userID: the authenticated user identifier from the bearer token.
 * - taskID: the task identifier from the request path.
 * - sessionID: the session identifier bound to the active chat view.
 * - userContent: the user message content.
 * - assistantContent: the assistant reply content.
 */
func (s *Store) CreateMessagePair(userID string, taskID string, sessionID string, userContent string, assistantContent string) (domain.Message, domain.Message, error) {
	ctx := context.Background()
	task, err := s.GetTask(userID, taskID)
	if err != nil {
		return domain.Message{}, domain.Message{}, err
	}

	session, err := s.GetSession(userID, sessionID)
	if err != nil {
		return domain.Message{}, domain.Message{}, err
	}
	if session.TaskID != taskID {
		return domain.Message{}, domain.Message{}, errors.New("session does not belong to the task")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Message{}, domain.Message{}, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	userMessage := domain.Message{
		ID:        generateUUID(),
		TaskID:    taskID,
		SessionID: session.ID,
		Role:      "user",
		Content:   userContent,
		Status:    "success",
		TimeLabel: now.Local().Format("15:04"),
	}

	assistantMessage := domain.Message{
		ID:        generateUUID(),
		TaskID:    taskID,
		SessionID: session.ID,
		Role:      "assistant",
		Content:   assistantContent,
		Status:    "success",
		TimeLabel: now.Local().Format("15:04"),
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO messages (
		    id, workspace_id, task_id, session_id, sender_type, sender_id, role,
		    message_type, content_md, created_at
		  )
		  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		userMessage.ID,
		task.WorkspaceID,
		taskID,
		session.ID,
		"user",
		userID,
		"user",
		"text",
		userContent,
		now,
	)
	if err != nil {
		return domain.Message{}, domain.Message{}, err
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO messages (
		    id, workspace_id, task_id, session_id, sender_type, role,
		    message_type, content_md, created_at
		  )
		  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		assistantMessage.ID,
		task.WorkspaceID,
		taskID,
		session.ID,
		"agent",
		"assistant",
		"text",
		assistantContent,
		now,
	)
	if err != nil {
		return domain.Message{}, domain.Message{}, err
	}

	_, err = tx.ExecContext(
		ctx,
		`UPDATE tasks
		 SET status = 'running', current_session_id = $2, current_primary_agent_id = NULLIF($3, '')::uuid, updated_at = $4
		 WHERE id = $1`,
		taskID,
		session.ID,
		session.PrimaryAgentID,
		now,
	)
	if err != nil {
		return domain.Message{}, domain.Message{}, err
	}

	_, err = tx.ExecContext(
		ctx,
		`UPDATE task_sessions
		 SET last_active_at = $2, last_message_at = $2, last_message_preview = $3, updated_at = $2, started_at = COALESCE(started_at, $2)
		 WHERE id = $1`,
		session.ID,
		now,
		assistantContent,
	)
	if err != nil {
		return domain.Message{}, domain.Message{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Message{}, domain.Message{}, err
	}

	return userMessage, assistantMessage, nil
}

/**
 * ensureCompatibility applies the minimal schema compatibility checks required by the current backend.
 * Params:
 * - ctx: the request-independent context used for startup SQL checks.
 */
func (s *Store) ensureCompatibility(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash TEXT`)
	return err
}

func scanSession(scanner interface {
	Scan(dest ...any) error
}) (domain.Session, error) {
	var session domain.Session
	var startedAt sql.NullTime
	var lastActiveAt time.Time

	err := scanner.Scan(
		&session.ID,
		&session.TaskID,
		&session.Title,
		&session.ChatMode,
		&session.SessionKind,
		&session.PrimaryAgentID,
		&session.PrimaryAgentName,
		&session.RuntimeProvider,
		&session.RuntimeSessionKey,
		&session.CreatedFromSession,
		&startedAt,
		&lastActiveAt,
		&session.LastMessagePreview,
	)
	if err != nil {
		return domain.Session{}, err
	}

	if startedAt.Valid {
		session.StartedAt = startedAt.Time.UTC().Format(time.RFC3339)
	}
	session.LastActiveAt = lastActiveAt.UTC().Format(time.RFC3339)
	session.LastActiveAtLabel = formatRelativeTime(lastActiveAt)
	return session, nil
}

func (s *Store) findAgentByName(ctx context.Context, workspaceID string, name string) (runtimeAgent, error) {
	var agent runtimeAgent
	err := s.db.QueryRowContext(
		ctx,
		`SELECT id, name, kind, provider_type
		 FROM agents
		 WHERE workspace_id = $1 AND status = 'active' AND name = $2
		 LIMIT 1`,
		workspaceID,
		name,
	).Scan(&agent.ID, &agent.Name, &agent.Kind, &agent.ProviderType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return runtimeAgent{}, fmt.Errorf("agent not found: %s", name)
		}
		return runtimeAgent{}, err
	}
	return agent, nil
}

func (s *Store) findAgentByID(ctx context.Context, workspaceID string, agentID string) (runtimeAgent, error) {
	var agent runtimeAgent
	err := s.db.QueryRowContext(
		ctx,
		`SELECT id, name, kind, provider_type
		 FROM agents
		 WHERE workspace_id = $1 AND status = 'active' AND id = $2
		 LIMIT 1`,
		workspaceID,
		agentID,
	).Scan(&agent.ID, &agent.Name, &agent.Kind, &agent.ProviderType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return runtimeAgent{}, errors.New("agent not found")
		}
		return runtimeAgent{}, err
	}
	return agent, nil
}

func hashPassword(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func generateUUID() string {
	buffer := make([]byte, 16)
	_, _ = rand.Read(buffer)
	buffer[6] = (buffer[6] & 0x0f) | 0x40
	buffer[8] = (buffer[8] & 0x3f) | 0x80

	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		buffer[0:4],
		buffer[4:6],
		buffer[6:8],
		buffer[8:10],
		buffer[10:16],
	)
}

func formatRelativeTime(value time.Time) string {
	diff := time.Since(value)
	switch {
	case diff < time.Minute:
		return "刚刚"
	case diff < time.Hour:
		return fmt.Sprintf("%d 分钟前", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%d 小时前", int(diff.Hours()))
	default:
		return value.Local().Format("01-02 15:04")
	}
}

func mapTaskStatus(rawStatus string) string {
	switch rawStatus {
	case "draft", "queued":
		return "idle"
	case "running":
		return "running"
	case "completed", "success":
		return "completed"
	case "failed", "error":
		return "failed"
	default:
		return "idle"
	}
}
