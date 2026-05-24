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
  updatedAtLabel: string;
};

export type Conversation = {
  id: string;
  taskId: string;
  agentName: string;
  agentType: "main";
  summary: string;
};

export type MessageRole = "user" | "assistant" | "system";

export type Message = {
  id: string;
  role: MessageRole;
  content: string;
  timeLabel: string;
};
