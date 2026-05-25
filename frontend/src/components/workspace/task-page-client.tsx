// Date: 2026-05-25
// Author: XinYang Li

"use client";

import { useEffect, useState, useTransition } from "react";
import { useRouter } from "next/navigation";

import { AgentColumn } from "@/components/workspace/agent-column";
import { ChatColumn } from "@/components/workspace/chat-column";
import { SessionCreateModal } from "@/components/workspace/session-create-modal";
import { TaskSessionSidebar } from "@/components/workspace/task-session-sidebar";
import {
  createMessage,
  createSession,
  getCurrentUser,
  getMessages,
  getSessionAgents,
  getSessions,
  getTask,
  getTasks,
  getWorkspace,
} from "@/lib/api";
import { getStoredToken } from "@/lib/auth";
import type { AgentOption, Message, Session, Task, User, Workspace } from "@/types/domain";

/**
 * Renders the task workspace page backed by real API calls.
 * @param props.taskId The task identifier from the route segment.
 * @returns The interactive task page.
 */
export function TaskPageClient({ taskId }: { taskId: string }): JSX.Element {
  const router = useRouter();
  const [isPending, startTransition] = useTransition();
  const [errorMessage, setErrorMessage] = useState("");
  const [token, setToken] = useState<string | null>(null);
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [workspace, setWorkspace] = useState<Workspace | null>(null);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [task, setTask] = useState<Task | null>(null);
  const [sessions, setSessions] = useState<Session[]>([]);
  const [sessionAgents, setSessionAgents] = useState<AgentOption[]>([]);
  const [activeSessionId, setActiveSessionId] = useState("");
  const [messages, setMessages] = useState<Message[]>([]);
  const [isCreateSessionOpen, setIsCreateSessionOpen] = useState(false);
  const [sessionDraft, setSessionDraft] = useState({
    title: "",
    chatMode: "single" as const,
    primaryAgentId: "",
  });

  useEffect(() => {
    const storedToken = getStoredToken();
    if (!storedToken) {
      router.replace("/login");
      return;
    }

    setToken(storedToken);

    startTransition(async () => {
      try {
        const [user, loadedWorkspace, loadedTasks, loadedTask, loadedSessions, loadedAgents] = await Promise.all([
          getCurrentUser(storedToken),
          getWorkspace(storedToken),
          getTasks(storedToken),
          getTask(storedToken, taskId),
          getSessions(storedToken, taskId),
          getSessionAgents(storedToken, taskId),
        ]);

        setCurrentUser(user);
        setWorkspace(loadedWorkspace);
        setTasks(loadedTasks);
        setTask(loadedTask);
        setSessions(loadedSessions);
        setSessionAgents(loadedAgents);

        const initialSessionId = loadedTask.currentSessionId || loadedSessions[0]?.id || "";
        setActiveSessionId(initialSessionId);
      } catch (error) {
        setErrorMessage(error instanceof Error ? error.message : "加载任务失败");
      }
    });
  }, [router, taskId]);

  useEffect(() => {
    if (!token || !activeSessionId) {
      return;
    }

    setMessages([]);

    startTransition(async () => {
      try {
        const loadedMessages = await getMessages(token, taskId, activeSessionId);
        setMessages(loadedMessages);
      } catch (error) {
        setErrorMessage(error instanceof Error ? error.message : "加载消息失败");
      }
    });
  }, [activeSessionId, taskId, token]);

  const activeSession = sessions.find((session) => session.id === activeSessionId) ?? sessions[0] ?? null;

  if (!token || !currentUser || !workspace || !task || !activeSession) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-mist text-ink">
        <div className="rounded-[28px] border border-line bg-paper px-6 py-5 shadow-panel">
          {errorMessage || "正在加载任务工作台..."}
        </div>
      </main>
    );
  }

  return (
    <main className="h-screen overflow-hidden bg-mist p-6 text-ink">
      <div className="grid h-full gap-6 xl:grid-cols-[320px_280px_1fr]">
        <TaskSessionSidebar
          activeSessionId={activeSession.id}
          onCreateSession={() => setIsCreateSessionOpen(true)}
          onSelectSession={(sessionId) => {
            setErrorMessage("");
            setMessages([]);
            setActiveSessionId(sessionId);
          }}
          sessions={sessions}
          task={task}
          user={currentUser}
        />

        <AgentColumn session={activeSession} />

        <ChatColumn
          errorMessage={errorMessage}
          isSending={isPending}
          messages={messages}
          onSendMessage={(content) => {
            setErrorMessage("");

            startTransition(async () => {
              try {
                const response = await createMessage(token, task.id, {
                  sessionId: activeSession.id,
                  content,
                });

                setMessages((current) => [...current, response.userMessage, response.assistantMessage]);
                setTask((current) =>
                  current
                    ? {
                        ...current,
                        status: "running",
                        currentSessionId: activeSession.id,
                        updatedAtLabel: "刚刚",
                      }
                    : current,
                );
                setTasks((current) =>
                  current.map((item) =>
                    item.id === task.id
                      ? {
                          ...item,
                          status: "running",
                          currentSessionId: activeSession.id,
                          updatedAtLabel: "刚刚",
                        }
                      : item,
                  ),
                );
                setSessions((current) =>
                  current.map((item) =>
                    item.id === activeSession.id
                      ? {
                          ...item,
                          startedAt: item.startedAt || new Date().toISOString(),
                          lastActiveAtLabel: "刚刚",
                          lastMessagePreview: response.assistantMessage.content,
                        }
                      : item,
                  ),
                );
              } catch (error) {
                setErrorMessage(error instanceof Error ? error.message : "发送消息失败");
              }
            });
          }}
          session={activeSession}
          task={task}
        />
      </div>

      <SessionCreateModal
        agents={sessionAgents}
        chatMode={sessionDraft.chatMode}
        isOpen={isCreateSessionOpen}
        isSubmitting={isPending}
        onAgentChange={(value) => setSessionDraft((current) => ({ ...current, primaryAgentId: value }))}
        onChatModeChange={(value) => setSessionDraft((current) => ({ ...current, chatMode: value }))}
        onClose={() => setIsCreateSessionOpen(false)}
        onSubmit={() => {
          setErrorMessage("");

          startTransition(async () => {
            try {
              const createdSession = await createSession(token, {
                taskId: task.id,
                title: sessionDraft.title,
                chatMode: sessionDraft.chatMode,
                primaryAgentId: sessionDraft.primaryAgentId,
              });

              setSessions((current) => [...current, createdSession].sort((left, right) => {
                if (left.sessionKind === "primary" && right.sessionKind !== "primary") {
                  return -1;
                }
                if (left.sessionKind !== "primary" && right.sessionKind === "primary") {
                  return 1;
                }
                return 0;
              }));
              setSessionDraft({
                title: "",
                chatMode: "single",
                primaryAgentId: "",
              });
              setIsCreateSessionOpen(false);
              setActiveSessionId(createdSession.id);
              setMessages([]);
            } catch (error) {
              setErrorMessage(error instanceof Error ? error.message : "创建 Session 失败");
            }
          });
        }}
        onTitleChange={(value) => setSessionDraft((current) => ({ ...current, title: value }))}
        primaryAgentId={sessionDraft.primaryAgentId}
        title={sessionDraft.title}
      />
    </main>
  );
}
