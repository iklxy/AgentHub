// Date: 2026-05-25
// Author: XinYang Li

package domain

// User represents a user account for AgentHub.
type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatarUrl"`
}

// Workspace represents one workspace visible to the current user.
type Workspace struct {
	ID          string `json:"id"`
	UserID      string `json:"userId"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Task represents one task entity inside the workspace.
type Task struct {
	ID               string `json:"id"`
	WorkspaceID      string `json:"workspaceId"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	Status           string `json:"status"`
	CurrentSessionID string `json:"currentSessionId"`
	UpdatedAtLabel   string `json:"updatedAtLabel"`
}

// Session represents one task-level conversation container.
type Session struct {
	ID                 string `json:"id"`
	TaskID             string `json:"taskId"`
	Title              string `json:"title"`
	ChatMode           string `json:"chatMode"`
	SessionKind        string `json:"sessionKind"`
	PrimaryAgentID     string `json:"primaryAgentId"`
	PrimaryAgentName   string `json:"primaryAgentName"`
	RuntimeProvider    string `json:"runtimeProvider"`
	RuntimeSessionKey  string `json:"runtimeSessionKey"`
	CreatedFromSession string `json:"createdFromSessionId"`
	StartedAt          string `json:"startedAt"`
	CreatedAt          string `json:"createdAt"`
	CreatedAtLabel     string `json:"createdAtLabel"`
	LastActiveAt       string `json:"lastActiveAt"`
	LastActiveAtLabel  string `json:"lastActiveAtLabel"`
	LastMessagePreview string `json:"lastMessagePreview"`
	IsPinned           bool   `json:"isPinned"`
}

// AgentOption represents one available agent choice in session creation flows.
type AgentOption struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Kind         string `json:"kind"`
	ProviderType string `json:"providerType"`
}

// Message represents one chat message row for a session transcript.
type Message struct {
	ID        string `json:"id"`
	TaskID    string `json:"taskId"`
	SessionID string `json:"sessionId"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	Status    string `json:"status"`
	TimeLabel string `json:"timeLabel"`
}

// LoginRequest carries the login payload.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest carries the register payload.
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UpdateWorkspaceRequest carries workspace editing input.
type UpdateWorkspaceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateTaskRequest carries task creation input.
type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// CreateSessionRequest carries task session creation input.
type CreateSessionRequest struct {
	TaskID         string `json:"taskId"`
	Title          string `json:"title"`
	ChatMode       string `json:"chatMode"`
	PrimaryAgentID string `json:"primaryAgentId"`
}

// UpdateSessionRequest carries task session update input.
type UpdateSessionRequest struct {
	Title      *string `json:"title"`
	IsPinned   *bool   `json:"isPinned"`
	IsArchived *bool   `json:"isArchived"`
}

// CreateMessageRequest carries user input for one session round.
type CreateMessageRequest struct {
	SessionID string `json:"sessionId"`
	Content   string `json:"content"`
}
