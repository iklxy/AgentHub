// Date: 2026-05-25
// Author: XinYang Li

package store

import "github.com/lixinyang/agenthub/backend/internal/domain"

// Repository defines the persistence contract for AgentHub v0.2.
type Repository interface {
	CreateUser(input domain.RegisterRequest) (domain.User, error)
	AuthenticateUser(input domain.LoginRequest) (domain.User, error)
	GetUserByID(userID string) (domain.User, error)
	GetWorkspaceByUserID(userID string) (domain.Workspace, error)
	UpdateWorkspace(userID string, input domain.UpdateWorkspaceRequest) (domain.Workspace, error)
	ListTasks(userID string) ([]domain.Task, error)
	CreateTask(userID string, input domain.CreateTaskRequest) (domain.Task, error)
	GetTask(userID string, taskID string) (domain.Task, error)
	ListSessions(userID string, taskID string) ([]domain.Session, error)
	GetSession(userID string, sessionID string) (domain.Session, error)
	CreateSession(userID string, input domain.CreateSessionRequest) (domain.Session, error)
	UpdateSession(userID string, sessionID string, input domain.UpdateSessionRequest) (domain.Session, error)
	ListSessionAgents(userID string, taskID string) ([]domain.AgentOption, error)
	ListMessages(userID string, taskID string, sessionID string) ([]domain.Message, error)
	CreateMessagePair(userID string, taskID string, sessionID string, userContent string, assistantContent string) (domain.Message, domain.Message, error)
}
