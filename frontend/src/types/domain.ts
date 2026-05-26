// Date: 2026-05-25
// Author: XinYang Li

export type User = {
  id: string;
  username: string;
  email: string;
  avatarUrl: string | null;
};

export type Workspace = {
  id: string;
  name: string;
  description: string;
};

export type TaskStatus = "idle" | "running" | "failed" | "completed";

export type Task = {
  id: string;
  title: string;
  description: string;
  status: TaskStatus;
  currentSessionId: string;
  updatedAtLabel: string;
};

export type SessionChatMode = "single" | "group";
export type SessionKind = "primary" | "branch";

export type Session = {
  id: string;
  taskId: string;
  title: string;
  chatMode: SessionChatMode;
  sessionKind: SessionKind;
  primaryAgentId: string;
  primaryAgentName: string;
  runtimeProvider: string;
  runtimeSessionKey: string;
  createdFromSessionId: string;
  startedAt: string;
  createdAt: string;
  createdAtLabel: string;
  lastActiveAt: string;
  lastActiveAtLabel: string;
  lastMessagePreview: string;
  isPinned: boolean;
};

export type AgentOption = {
  id: string;
  name: string;
  kind: string;
  providerType: string;
};

export type MessageRole = "user" | "assistant" | "system";

export type Message = {
  id: string;
  taskId: string;
  sessionId: string;
  role: MessageRole;
  content: string;
  timeLabel: string;
  replyToMessageId: string;
};
