// Date: 2026-05-25
// Author: XinYang Li

package store

import "github.com/lixinyang/agenthub/backend/internal/domain"

// Repository defines the persistence contract for AgentHub v0.1.
type Repository interface {
	CreateUser(input domain.RegisterRequest) (domain.User, error)
	AuthenticateUser(input domain.LoginRequest) (domain.User, error)
	GetUserByID(userID string) (domain.User, error)
	GetWorkspaceByUserID(userID string) (domain.Workspace, error)
	UpdateWorkspace(userID string, input domain.UpdateWorkspaceRequest) (domain.Workspace, error)
	ListTasks(userID string) ([]domain.Task, error)
	CreateTask(userID string, input domain.CreateTaskRequest) (domain.Task, error)
	GetTask(userID string, taskID string) (domain.Task, error)
	GetConversation(userID string, taskID string) (domain.Conversation, error)
	ListMessages(userID string, taskID string) ([]domain.Message, error)
	CreateMessagePair(userID string, taskID string, userContent string, assistantContent string) (domain.Message, domain.Message, error)
}
