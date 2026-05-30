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
	"slices"
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

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO agents (
		    id, workspace_id, name, kind, status, description, capability_tags,
		    source_kind, provider_type, tool_schema_json, created_by, created_at, updated_at
		  )
		  VALUES ($1, $2, $3, $4, $5, $6, '[]'::jsonb, $7, $8, '{}'::jsonb, $9, $10, $11)`,
		generateUUID(),
		workspaceID,
		"Aries",
		"document",
		"active",
		"文档 Agent，负责当前工作区下单聊模式的文档整理与改写协作",
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
 * CreateTask inserts one task for the authenticated user without auto-creating any session.
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

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Task{}, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	taskID := generateUUID()

	task := domain.Task{
		ID:               taskID,
		WorkspaceID:      workspace.ID,
		Title:            input.Title,
		Description:      input.Description,
		Status:           "idle",
		CurrentSessionID: "",
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
		nil,
		nil,
		userID,
		now,
		now,
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
		    ts.created_at,
		    ts.last_active_at,
		    COALESCE(ts.last_message_preview, ''),
		    COALESCE(ts.is_pinned, false)
		 FROM task_sessions ts
		 INNER JOIN workspace_members wm ON wm.workspace_id = ts.workspace_id
		 LEFT JOIN agents a ON a.id = ts.primary_agent_id
		 WHERE wm.user_id = $1 AND ts.task_id = $2 AND ts.status = 'active' AND ts.archived_at IS NULL
		 ORDER BY ts.is_pinned DESC, ts.last_active_at DESC, ts.created_at DESC`,
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
		    ts.created_at,
		    ts.last_active_at,
		    COALESCE(ts.last_message_preview, ''),
		    COALESCE(ts.is_pinned, false)
		 FROM task_sessions ts
		 INNER JOIN workspace_members wm ON wm.workspace_id = ts.workspace_id
		 LEFT JOIN agents a ON a.id = ts.primary_agent_id
		 WHERE wm.user_id = $1 AND ts.id = $2 AND ts.archived_at IS NULL
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
 * CreateSession creates one single-chat session under a task for the authenticated user.
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
		"",
		nil,
	)
	if err != nil {
		return domain.Session{}, err
	}

	return s.GetSession(userID, sessionID)
}

/**
 * UpdateSession updates editable metadata and list-management flags for one session.
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

	nextTitle := session.Title
	if input.Title != nil {
		trimmedTitle := strings.TrimSpace(*input.Title)
		if trimmedTitle != "" {
			nextTitle = trimmedTitle
		}
	}

	nextPinned := session.IsPinned
	if input.IsPinned != nil {
		nextPinned = *input.IsPinned
	}

	now := time.Now().UTC()

	if input.IsArchived != nil {
		var archivedAt any
		if *input.IsArchived {
			archivedAt = now
		}

		_, err = s.db.ExecContext(
			ctx,
			`UPDATE task_sessions
			 SET title = $2, is_pinned = $3, archived_at = $4, updated_at = $5
			 WHERE id = $1`,
			session.ID,
			nextTitle,
			nextPinned,
			archivedAt,
			now,
		)
		if err != nil {
			return domain.Session{}, err
		}

		if *input.IsArchived {
			_, err = s.db.ExecContext(
				ctx,
				`UPDATE tasks
				 SET current_session_id = CASE WHEN current_session_id = $2::uuid THEN NULL ELSE current_session_id END,
				     current_primary_agent_id = CASE WHEN current_session_id = $2::uuid THEN NULL ELSE current_primary_agent_id END,
				     updated_at = $3
				 WHERE id = $1`,
				session.TaskID,
				session.ID,
				now,
			)
			if err != nil {
				return domain.Session{}, err
			}
			return domain.Session{}, nil
		}
	} else {
		_, err = s.db.ExecContext(
			ctx,
			`UPDATE task_sessions
			 SET title = $2, is_pinned = $3, updated_at = $4
			 WHERE id = $1`,
			session.ID,
			nextTitle,
			nextPinned,
			now,
		)
		if err != nil {
			return domain.Session{}, err
		}
	}

	return s.GetSession(userID, sessionID)
}

/**
 * ListSessionAgents returns the enabled agents that can be used in the single-chat session creation flow.
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
		 WHERE workspace_id = $1 AND status = 'active'
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
		`SELECT m.id, m.task_id, m.session_id, m.role, m.content_md, m.created_at,
		        COALESCE(m.reply_to_message_id::text, ''), COALESCE(m.is_pinned, false)
		 FROM messages m
		 INNER JOIN workspace_members wm ON wm.workspace_id = m.workspace_id
		 WHERE wm.user_id = $1 AND m.task_id = $2 AND m.session_id = $3
		 ORDER BY m.created_at ASC,
		          CASE m.role
		            WHEN 'user' THEN 0
		            WHEN 'assistant' THEN 1
		            ELSE 2
		          END ASC,
		          m.id ASC`,
		userID,
		taskID,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []domain.Message{}
	messageIDs := []string{}
	for rows.Next() {
		var message domain.Message
		var createdAt time.Time
		if err := rows.Scan(&message.ID, &message.TaskID, &message.SessionID, &message.Role, &message.Content, &createdAt, &message.ReplyToID, &message.IsPinned); err != nil {
			return nil, err
		}
		message.Status = "success"
		message.TimeLabel = createdAt.Local().Format("15:04")
		messages = append(messages, message)
		messageIDs = append(messageIDs, message.ID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	attachmentsByMessageID, err := s.loadAttachmentsByMessageIDs(ctx, userID, messageIDs)
	if err != nil {
		return nil, err
	}

	for index := range messages {
		messages[index].Attachments = attachmentsByMessageID[messages[index].ID]
	}

	return messages, nil
}

/**
 * GetMessageByID loads one message the current user can access.
 * Params:
 * - userID: the authenticated user identifier from the bearer token.
 * - messageID: the message identifier selected by the front-end action.
 * Returns:
 * - the resolved message entity for quote and reply actions.
 */
func (s *Store) GetMessageByID(userID string, messageID string) (domain.Message, error) {
	ctx := context.Background()
	var message domain.Message
	var createdAt time.Time

	err := s.db.QueryRowContext(
		ctx,
		`SELECT m.id, m.task_id, m.session_id, m.role, m.content_md, m.created_at,
		        COALESCE(m.reply_to_message_id::text, ''), COALESCE(m.is_pinned, false)
		 FROM messages m
		 INNER JOIN workspace_members wm ON wm.workspace_id = m.workspace_id
		 WHERE wm.user_id = $1 AND m.id = $2
		 LIMIT 1`,
		userID,
		messageID,
	).Scan(&message.ID, &message.TaskID, &message.SessionID, &message.Role, &message.Content, &createdAt, &message.ReplyToID, &message.IsPinned)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Message{}, errors.New("message not found")
		}
		return domain.Message{}, err
	}

	message.Status = "success"
	message.TimeLabel = createdAt.Local().Format("15:04")
	attachmentsByMessageID, err := s.loadAttachmentsByMessageIDs(ctx, userID, []string{message.ID})
	if err != nil {
		return domain.Message{}, err
	}
	message.Attachments = attachmentsByMessageID[message.ID]
	return message, nil
}

/**
 * GetMessagesByIDs loads multiple messages the current user can access.
 * Params:
 * - userID: the authenticated user identifier from the bearer token.
 * - messageIDs: the source message identifiers selected by the front-end quote action.
 * Returns:
 * - the resolved message list in the same order as the incoming identifiers.
 */
func (s *Store) GetMessagesByIDs(userID string, messageIDs []string) ([]domain.Message, error) {
	ctx := context.Background()
	if len(messageIDs) == 0 {
		return nil, errors.New("message ids are required")
	}

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT m.id, m.task_id, m.session_id, m.role, m.content_md, m.created_at,
		        COALESCE(m.reply_to_message_id::text, ''), COALESCE(m.is_pinned, false)
		 FROM messages m
		 INNER JOIN workspace_members wm ON wm.workspace_id = m.workspace_id
		 WHERE wm.user_id = $1 AND m.id = ANY($2)`,
		userID,
		messageIDs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byID := make(map[string]domain.Message, len(messageIDs))
	for rows.Next() {
		var message domain.Message
		var createdAt time.Time
		if err := rows.Scan(&message.ID, &message.TaskID, &message.SessionID, &message.Role, &message.Content, &createdAt, &message.ReplyToID, &message.IsPinned); err != nil {
			return nil, err
		}
		message.Status = "success"
		message.TimeLabel = createdAt.Local().Format("15:04")
		byID[message.ID] = message
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	attachmentsByMessageID, err := s.loadAttachmentsByMessageIDs(ctx, userID, messageIDs)
	if err != nil {
		return nil, err
	}
	for messageID, message := range byID {
		message.Attachments = attachmentsByMessageID[messageID]
		byID[messageID] = message
	}

	orderedMessages := make([]domain.Message, 0, len(messageIDs))
	for _, messageID := range messageIDs {
		message, ok := byID[messageID]
		if !ok {
			return nil, errors.New("one or more messages were not found")
		}
		orderedMessages = append(orderedMessages, message)
	}

	return slices.Clone(orderedMessages), nil
}

/**
 * ListPinnedMessages loads all pinned messages for one accessible session in pin order.
 * Params:
 * - userID: the authenticated user identifier from the bearer token.
 * - sessionID: the active session identifier used for prompt injection.
 */
func (s *Store) ListPinnedMessages(userID string, sessionID string) ([]domain.Message, error) {
	ctx := context.Background()
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT m.id, m.task_id, m.session_id, m.role, m.content_md, m.created_at,
		        COALESCE(m.reply_to_message_id::text, ''), true
		 FROM session_context_pins scp
		 INNER JOIN messages m ON m.id = scp.message_id
		 INNER JOIN workspace_members wm ON wm.workspace_id = scp.workspace_id
		 WHERE wm.user_id = $1 AND scp.session_id = $2
		 ORDER BY m.pinned_at ASC NULLS LAST, m.created_at ASC, m.id ASC`,
		userID,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []domain.Message{}
	messageIDs := []string{}
	for rows.Next() {
		var message domain.Message
		var createdAt time.Time
		if err := rows.Scan(&message.ID, &message.TaskID, &message.SessionID, &message.Role, &message.Content, &createdAt, &message.ReplyToID, &message.IsPinned); err != nil {
			return nil, err
		}
		message.Status = "success"
		message.TimeLabel = createdAt.Local().Format("15:04")
		messages = append(messages, message)
		messageIDs = append(messageIDs, message.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	attachmentsByMessageID, err := s.loadAttachmentsByMessageIDs(ctx, userID, messageIDs)
	if err != nil {
		return nil, err
	}
	for index := range messages {
		messages[index].Attachments = attachmentsByMessageID[messages[index].ID]
	}

	return messages, nil
}

/**
 * SetMessagePin updates the pinned status for one accessible message and synchronizes the session_context_pins table.
 * Params:
 * - userID: the authenticated user identifier from the bearer token.
 * - messageID: the selected message identifier.
 * - isPinned: the target pinned state requested by the user.
 */
func (s *Store) SetMessagePin(userID string, messageID string, isPinned bool) (domain.Message, error) {
	ctx := context.Background()
	if _, err := s.GetMessageByID(userID, messageID); err != nil {
		return domain.Message{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Message{}, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	if isPinned {
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO session_context_pins (id, workspace_id, session_id, message_id, created_at)
			 SELECT $1, m.workspace_id, m.session_id, m.id, $3
			 FROM messages m
			 WHERE m.id = $2
			 ON CONFLICT (session_id, message_id) DO NOTHING`,
			generateUUID(),
			messageID,
			now,
		)
		if err != nil {
			return domain.Message{}, err
		}

		_, err = tx.ExecContext(
			ctx,
			`UPDATE messages
			 SET is_pinned = true, pinned_at = COALESCE(pinned_at, $2)
			 WHERE id = $1`,
			messageID,
			now,
		)
		if err != nil {
			return domain.Message{}, err
		}
	} else {
		_, err = tx.ExecContext(ctx, `DELETE FROM session_context_pins WHERE message_id = $1`, messageID)
		if err != nil {
			return domain.Message{}, err
		}
		_, err = tx.ExecContext(
			ctx,
			`UPDATE messages
			 SET is_pinned = false, pinned_at = NULL
			 WHERE id = $1`,
			messageID,
		)
		if err != nil {
			return domain.Message{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return domain.Message{}, err
	}

	return s.GetMessageByID(userID, messageID)
}

/**
 * GetDraftAttachments loads uploaded attachments that are not yet bound to another message.
 * Params:
 * - userID: the authenticated user identifier from the bearer token.
 * - taskID: the owning task identifier.
 * - sessionID: the owning session identifier.
 * - attachmentIDs: the uploaded attachment identifiers selected for one message send.
 */
func (s *Store) GetDraftAttachments(userID string, taskID string, sessionID string, attachmentIDs []string) ([]domain.Attachment, error) {
	ctx := context.Background()
	if len(attachmentIDs) == 0 {
		return []domain.Attachment{}, nil
	}

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT ta.id, ta.task_id, ta.session_id, COALESCE(ta.message_id::text, ''), ta.file_name, ta.file_type,
		        ta.source_type, COALESCE(ta.source_url, ''), ta.storage_key
		 FROM task_attachments ta
		 INNER JOIN workspace_members wm ON wm.workspace_id = ta.workspace_id
		 WHERE wm.user_id = $1
		   AND ta.task_id = $2
		   AND ta.session_id = $3
		   AND ta.created_by = $1::uuid
		   AND ta.id = ANY($4::uuid[])
		   AND ta.message_id IS NULL`,
		userID,
		taskID,
		sessionID,
		attachmentIDs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byID := make(map[string]domain.Attachment, len(attachmentIDs))
	for rows.Next() {
		var item domain.Attachment
		if err := rows.Scan(&item.ID, &item.TaskID, &item.SessionID, &item.MessageID, &item.FileName, &item.FileType, &item.SourceType, &item.SourceURL, &item.StorageKey); err != nil {
			return nil, err
		}
		byID[item.ID] = item
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	items := make([]domain.Attachment, 0, len(attachmentIDs))
	for _, attachmentID := range attachmentIDs {
		item, ok := byID[attachmentID]
		if !ok {
			return nil, errors.New("one or more attachments were not found")
		}
		items = append(items, item)
	}

	return items, nil
}

/**
 * CreateAttachment persists one uploaded file or image row for later message binding.
 * Params:
 * - userID: the authenticated user identifier from the bearer token.
 * - taskID: the owning task identifier.
 * - sessionID: the owning session identifier.
 * - fileName: the original file name shown in the UI.
 * - fileType: the detected MIME type.
 * - sourceType: the logical attachment kind, currently image or file.
 * - storageKey: the absolute local storage path written by the upload handler.
 */
func (s *Store) CreateAttachment(userID string, taskID string, sessionID string, fileName string, fileType string, sourceType string, storageKey string) (domain.Attachment, error) {
	ctx := context.Background()
	task, err := s.GetTask(userID, taskID)
	if err != nil {
		return domain.Attachment{}, err
	}

	attachmentID := generateUUID()
	sourceURL := fmt.Sprintf("/api/files/%s", attachmentID)
	now := time.Now().UTC()
	_, err = s.db.ExecContext(
		ctx,
		`INSERT INTO task_attachments (
		    id, workspace_id, task_id, session_id, message_id, file_name, file_type,
		    storage_key, source_type, source_url, meta_json, created_by, created_at
		  )
		  VALUES ($1, $2, $3, $4, NULL, $5, $6, $7, $8, $9, '{}'::jsonb, $10, $11)`,
		attachmentID,
		task.WorkspaceID,
		taskID,
		sessionID,
		fileName,
		fileType,
		storageKey,
		sourceType,
		sourceURL,
		userID,
		now,
	)
	if err != nil {
		return domain.Attachment{}, err
	}

	return domain.Attachment{
		ID:         attachmentID,
		TaskID:     taskID,
		SessionID:  sessionID,
		MessageID:  "",
		FileName:   fileName,
		FileType:   fileType,
		SourceType: sourceType,
		SourceURL:  sourceURL,
		StorageKey: storageKey,
	}, nil
}

/**
 * GetAttachmentByID loads one uploaded attachment the current user can access.
 * Params:
 * - userID: the authenticated user identifier from the bearer token.
 * - attachmentID: the attachment identifier used by render and download flows.
 */
func (s *Store) GetAttachmentByID(userID string, attachmentID string) (domain.Attachment, error) {
	ctx := context.Background()
	var item domain.Attachment
	err := s.db.QueryRowContext(
		ctx,
		`SELECT ta.id, ta.task_id, ta.session_id, COALESCE(ta.message_id::text, ''), ta.file_name, ta.file_type,
		        ta.source_type, COALESCE(ta.source_url, ''), ta.storage_key
		 FROM task_attachments ta
		 INNER JOIN workspace_members wm ON wm.workspace_id = ta.workspace_id
		 WHERE wm.user_id = $1 AND ta.id = $2
		 LIMIT 1`,
		userID,
		attachmentID,
	).Scan(&item.ID, &item.TaskID, &item.SessionID, &item.MessageID, &item.FileName, &item.FileType, &item.SourceType, &item.SourceURL, &item.StorageKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Attachment{}, errors.New("attachment not found")
		}
		return domain.Attachment{}, err
	}
	return item, nil
}

/**
 * CreateMessagePair stores the user message and assistant reply for one task session round.
 * Params:
 * - userID: the authenticated user identifier from the bearer token.
 * - taskID: the task identifier from the request path.
 * - sessionID: the session identifier bound to the active chat view.
 * - userContent: the user message content.
 * - assistantContent: the assistant reply content.
 * - replyToMessageID: the replied source message identifier when the new user message is a reply action.
 */
func (s *Store) CreateMessagePair(userID string, taskID string, sessionID string, userContent string, assistantContent string, replyToMessageID *string, attachments []domain.Attachment, newSDKSessionID string) (domain.Message, domain.Message, error) {
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
	userCreatedAt := now
	assistantCreatedAt := now.Add(time.Microsecond)
	userMessage := domain.Message{
		ID:          generateUUID(),
		TaskID:      taskID,
		SessionID:   session.ID,
		Role:        "user",
		Content:     userContent,
		Status:      "success",
		TimeLabel:   userCreatedAt.Local().Format("15:04"),
		Attachments: slices.Clone(attachments),
	}
	if replyToMessageID != nil {
		userMessage.ReplyToID = *replyToMessageID
	}

	assistantMessage := domain.Message{
		ID:        generateUUID(),
		TaskID:    taskID,
		SessionID: session.ID,
		Role:      "assistant",
		Content:   assistantContent,
		Status:    "success",
		TimeLabel: assistantCreatedAt.Local().Format("15:04"),
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO messages (
		    id, workspace_id, task_id, session_id, sender_type, sender_id, role,
		    message_type, content_md, reply_to_message_id, created_at
		  )
		  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULLIF($10, '')::uuid, $11)`,
		userMessage.ID,
		task.WorkspaceID,
		taskID,
		session.ID,
		"user",
		userID,
		"user",
		"text",
		userContent,
		userMessage.ReplyToID,
		userCreatedAt,
	)
	if err != nil {
		return domain.Message{}, domain.Message{}, err
	}

	if len(attachments) > 0 {
		attachmentIDs := make([]string, 0, len(attachments))
		for _, attachment := range attachments {
			attachmentIDs = append(attachmentIDs, attachment.ID)
		}

		_, err = tx.ExecContext(
			ctx,
			`UPDATE task_attachments
			 SET message_id = $2
			 WHERE task_id = $1
			   AND session_id = $3
			   AND created_by = $4::uuid
			   AND id = ANY($5::uuid[])
			   AND message_id IS NULL`,
			taskID,
			userMessage.ID,
			session.ID,
			userID,
			attachmentIDs,
		)
		if err != nil {
			return domain.Message{}, domain.Message{}, err
		}

		for index := range userMessage.Attachments {
			userMessage.Attachments[index].MessageID = userMessage.ID
		}
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
		assistantCreatedAt,
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

	if newSDKSessionID != "" {
		_, err = tx.ExecContext(
			ctx,
			`UPDATE task_sessions
			 SET runtime_session_key = $1
			 WHERE id = $2 AND (runtime_session_key = '' OR runtime_session_key IS NULL)`,
			newSDKSessionID,
			session.ID,
		)
		if err != nil {
			return domain.Message{}, domain.Message{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return domain.Message{}, domain.Message{}, err
	}

	return userMessage, assistantMessage, nil
}

/**
 * CreateAssistantMessage stores one assistant-only message for regenerate actions and refreshes session activity metadata.
 * Params:
 * - userID: the authenticated user identifier from the bearer token.
 * - taskID: the task identifier from the request path.
 * - sessionID: the session identifier bound to the active chat view.
 * - assistantContent: the regenerated assistant content.
 */
func (s *Store) CreateAssistantMessage(userID string, taskID string, sessionID string, assistantContent string) (domain.Message, error) {
	ctx := context.Background()
	task, err := s.GetTask(userID, taskID)
	if err != nil {
		return domain.Message{}, err
	}

	session, err := s.GetSession(userID, sessionID)
	if err != nil {
		return domain.Message{}, err
	}
	if session.TaskID != taskID {
		return domain.Message{}, errors.New("session does not belong to the task")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Message{}, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
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
		return domain.Message{}, err
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
		return domain.Message{}, err
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
		return domain.Message{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Message{}, err
	}

	return assistantMessage, nil
}

/**
 * loadAttachmentsByMessageIDs loads message-bound attachments for one accessible message batch.
 * Params:
 * - ctx: the request-independent context shared by the parent query flow.
 * - userID: the authenticated user identifier from the bearer token.
 * - messageIDs: the message identifiers whose attachments should be attached to the DTO response.
 */
func (s *Store) loadAttachmentsByMessageIDs(ctx context.Context, userID string, messageIDs []string) (map[string][]domain.Attachment, error) {
	itemsByMessageID := make(map[string][]domain.Attachment)
	if len(messageIDs) == 0 {
		return itemsByMessageID, nil
	}

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT ta.id, ta.task_id, ta.session_id, COALESCE(ta.message_id::text, ''), ta.file_name, ta.file_type,
		        ta.source_type, COALESCE(ta.source_url, ''), ta.storage_key
		 FROM task_attachments ta
		 INNER JOIN workspace_members wm ON wm.workspace_id = ta.workspace_id
		 WHERE wm.user_id = $1 AND ta.message_id = ANY($2::uuid[])
		 ORDER BY ta.created_at ASC, ta.id ASC`,
		userID,
		messageIDs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.Attachment
		if err := rows.Scan(&item.ID, &item.TaskID, &item.SessionID, &item.MessageID, &item.FileName, &item.FileType, &item.SourceType, &item.SourceURL, &item.StorageKey); err != nil {
			return nil, err
		}
		itemsByMessageID[item.MessageID] = append(itemsByMessageID[item.MessageID], item)
	}

	return itemsByMessageID, rows.Err()
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
	var createdAt time.Time
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
		&createdAt,
		&lastActiveAt,
		&session.LastMessagePreview,
		&session.IsPinned,
	)
	if err != nil {
		return domain.Session{}, err
	}

	if startedAt.Valid {
		session.StartedAt = startedAt.Time.UTC().Format(time.RFC3339)
	}
	session.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	session.CreatedAtLabel = formatRelativeTime(createdAt)
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
