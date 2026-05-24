// Date: 2026-05-25
// Author: XinYang Li

package domain

// User represents a user account for AgentHub v0.1.
type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatarUrl"`
}

// Workspace represents one personal workspace tied to the current user.
type Workspace struct {
	ID          string `json:"id"`
	UserID      string `json:"userId"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Task represents one task entity inside the workspace.
type Task struct {
	ID             string `json:"id"`
	WorkspaceID    string `json:"workspaceId"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Status         string `json:"status"`
	UpdatedAtLabel string `json:"updatedAtLabel"`
}

// Conversation represents the main agent conversation bound to a task.
type Conversation struct {
	ID        string `json:"id"`
	TaskID    string `json:"taskId"`
	AgentName string `json:"agentName"`
	AgentType string `json:"agentType"`
	Summary   string `json:"summary"`
}

// Message represents one chat message row for a task transcript.
type Message struct {
	ID             string `json:"id"`
	TaskID         string `json:"taskId"`
	ConversationID string `json:"conversationId"`
	Role           string `json:"role"`
	Content        string `json:"content"`
	Status         string `json:"status"`
	TimeLabel      string `json:"timeLabel"`
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

// CreateMessageRequest carries user input for the main agent.
type CreateMessageRequest struct {
	Content string `json:"content"`
}
