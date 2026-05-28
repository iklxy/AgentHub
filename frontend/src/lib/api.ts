// Date: 2026-05-25
// Author: XinYang Li

import type { AgentOption, Message, Session, Task, User, Workspace } from "@/types/domain";
import { uiLog } from "@/lib/logger";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://192.168.139.155:8080";

type RegisterPayload = {
  username: string;
  email: string;
  password: string;
};

type LoginPayload = {
  email: string;
  password: string;
};

type UpdateWorkspacePayload = {
  name: string;
  description: string;
};

type CreateTaskPayload = {
  title: string;
  description: string;
};

type CreateSessionPayload = {
  taskId: string;
  title: string;
  chatMode: "single";
  primaryAgentId: string;
};

type UpdateSessionPayload = {
  title?: string;
  isPinned?: boolean;
  isArchived?: boolean;
};

type CreateMessagePayload = {
  sessionId: string;
  content: string;
};

type CreateMessageActionPayload = {
  content: string;
  messageIds?: string[];
};

type AuthResponse = {
  token: string;
  user: User;
};

type MessageRoundResponse = {
  userMessage: Message;
  assistantMessage: Message;
};

type RegenerateMessageResponse = {
  assistantMessage: Message;
};

/**
 * Executes an HTTP request against the AgentHub backend.
 * @param path The API path appended to the configured backend base URL.
 * @param init The request options, including method, body, and token.
 * @returns The parsed JSON response body.
 */
async function request<T>(
  path: string,
  init?: RequestInit & {
    token?: string | null;
  },
): Promise<T> {
  const headers = new Headers(init?.headers);
  headers.set("Content-Type", "application/json");
  const method = init?.method ?? "GET";

  if (init?.token) {
    headers.set("Authorization", `Bearer ${init.token}`);
  }

  let response: Response;
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      ...init,
      headers,
    });
  } catch (error) {
    uiLog("error", "api network request failed", {
      apiBaseUrl: API_BASE_URL,
      path,
      method,
      error: error instanceof Error ? error.message : String(error),
    });
    throw new Error(`Network request failed: ${method} ${API_BASE_URL}${path}`);
  }

  if (!response.ok) {
    const errorPayload = (await response.json().catch(() => null)) as { error?: string } | null;
    uiLog("error", "api request returned non-ok status", {
      apiBaseUrl: API_BASE_URL,
      path,
      method,
      status: response.status,
      statusText: response.statusText,
      error: errorPayload?.error ?? null,
    });
    throw new Error(errorPayload?.error ?? `Request failed: ${response.status}`);
  }

  return (await response.json()) as T;
}

/**
 * Registers a new user account through the backend API.
 * @param payload The registration form payload.
 * @returns The auth response containing token and user snapshot.
 */
export function register(payload: RegisterPayload): Promise<AuthResponse> {
  return request<AuthResponse>("/api/register", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

/**
 * Logs a user into the backend API.
 * @param payload The login form payload.
 * @returns The auth response containing token and user snapshot.
 */
export function login(payload: LoginPayload): Promise<AuthResponse> {
  return request<AuthResponse>("/api/login", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

/**
 * Loads the current authenticated user snapshot.
 * @param token The bearer token stored in the browser.
 * @returns The current user entity.
 */
export function getCurrentUser(token: string): Promise<User> {
  return request<User>("/api/me", { method: "GET", token });
}

/**
 * Loads the authenticated workspace snapshot.
 * @param token The bearer token stored in the browser.
 * @returns The workspace entity for the current user.
 */
export function getWorkspace(token: string): Promise<Workspace> {
  return request<Workspace>("/api/workspace", { method: "GET", token });
}

/**
 * Updates workspace metadata for the current user.
 * @param token The bearer token stored in the browser.
 * @param payload The new workspace name and description values.
 * @returns The updated workspace entity.
 */
export function updateWorkspace(token: string, payload: UpdateWorkspacePayload): Promise<Workspace> {
  return request<Workspace>("/api/workspace", {
    method: "PATCH",
    token,
    body: JSON.stringify(payload),
  });
}

/**
 * Loads the current user's task list.
 * @param token The bearer token stored in the browser.
 * @returns The ordered task list.
 */
export function getTasks(token: string): Promise<Task[]> {
  return request<Task[]>("/api/tasks", { method: "GET", token });
}

/**
 * Creates a task for the current user.
 * @param token The bearer token stored in the browser.
 * @param payload The task title and description payload.
 * @returns The created task entity.
 */
export function createTask(token: string, payload: CreateTaskPayload): Promise<Task> {
  return request<Task>("/api/tasks", {
    method: "POST",
    token,
    body: JSON.stringify(payload),
  });
}

/**
 * Loads one task by identifier.
 * @param token The bearer token stored in the browser.
 * @param taskId The task identifier from the route.
 * @returns The task entity.
 */
export function getTask(token: string, taskId: string): Promise<Task> {
  return request<Task>(`/api/tasks/${taskId}`, { method: "GET", token });
}

/**
 * Loads the session list for one task.
 * @param token The bearer token stored in the browser.
 * @param taskId The task identifier from the route.
 * @returns The ordered session list.
 */
export function getSessions(token: string, taskId: string): Promise<Session[]> {
  return request<Session[]>(`/api/tasks/${taskId}/sessions`, { method: "GET", token });
}

/**
 * Loads the selectable branch-session agents for one task.
 * @param token The bearer token stored in the browser.
 * @param taskId The task identifier from the route.
 * @returns The available agent option list.
 */
export function getSessionAgents(token: string, taskId: string): Promise<AgentOption[]> {
  return request<AgentOption[]>(`/api/tasks/${taskId}/session-agents`, { method: "GET", token });
}

/**
 * Creates one new task session for the current task.
 * @param token The bearer token stored in the browser.
 * @param payload The session creation payload.
 * @returns The created session entity.
 */
export function createSession(token: string, payload: CreateSessionPayload): Promise<Session> {
  return request<Session>("/api/sessions", {
    method: "POST",
    token,
    body: JSON.stringify(payload),
  });
}

/**
 * Updates one existing task session.
 * @param token The bearer token stored in the browser.
 * @param sessionId The session identifier to update.
 * @param payload The editable session payload.
 * @returns The updated session entity.
 */
export function updateSession(token: string, sessionId: string, payload: UpdateSessionPayload): Promise<Session> {
  return request<Session>(`/api/sessions/${sessionId}`, {
    method: "PATCH",
    token,
    body: JSON.stringify(payload),
  });
}

/**
 * Loads messages for one task session transcript.
 * @param token The bearer token stored in the browser.
 * @param taskId The task identifier from the route.
 * @param sessionId The active session identifier.
 * @returns The ordered message list.
 */
export function getMessages(token: string, taskId: string, sessionId: string): Promise<Message[]> {
  const params = new URLSearchParams({ sessionId });
  return request<Message[]>(`/api/tasks/${taskId}/messages?${params.toString()}`, { method: "GET", token });
}

/**
 * Sends a new user message for one task session transcript.
 * @param token The bearer token stored in the browser.
 * @param taskId The task identifier from the route.
 * @param payload The user message content payload with its target session identifier.
 * @returns The created user and assistant message pair.
 */
export function createMessage(token: string, taskId: string, payload: CreateMessagePayload): Promise<MessageRoundResponse> {
  return request<MessageRoundResponse>(`/api/tasks/${taskId}/messages`, {
    method: "POST",
    token,
    body: JSON.stringify(payload),
  });
}

/**
 * Creates a quote-based message round from one source message.
 * @param token The bearer token stored in the browser.
 * @param payload The user input plus source message identifiers selected in the transcript.
 * @returns The created user and assistant message pair.
 */
export function quoteMessage(token: string, payload: CreateMessageActionPayload): Promise<MessageRoundResponse> {
  return request<MessageRoundResponse>("/api/messages/quote", {
    method: "POST",
    token,
    body: JSON.stringify(payload),
  });
}

/**
 * Creates a reply-based message round from one source message.
 * @param token The bearer token stored in the browser.
 * @param messageId The source message identifier selected in the transcript.
 * @param payload The user input that should reply to the selected message.
 * @returns The created user and assistant message pair.
 */
export function replyMessage(token: string, messageId: string, payload: CreateMessageActionPayload): Promise<MessageRoundResponse> {
  return request<MessageRoundResponse>(`/api/messages/${messageId}/reply`, {
    method: "POST",
    token,
    body: JSON.stringify(payload),
  });
}

/**
 * Regenerates one assistant message and appends the new assistant answer to the current session.
 * @param token The bearer token stored in the browser.
 * @param messageId The assistant message identifier selected in the transcript.
 * @returns The regenerated assistant message created by the backend.
 */
export function regenerateMessage(token: string, messageId: string): Promise<RegenerateMessageResponse> {
  return request<RegenerateMessageResponse>(`/api/messages/${messageId}/regenerate`, {
    method: "POST",
    token,
    body: JSON.stringify({}),
  });
}
