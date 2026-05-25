// Date: 2026-05-25
// Author: XinYang Li

"use client";

import { useEffect, useState, useTransition } from "react";
import { useRouter } from "next/navigation";

import { AgentColumn } from "@/components/workspace/agent-column";
import { ChatColumn } from "@/components/workspace/chat-column";
import { SessionCreateModal } from "@/components/workspace/session-create-modal";
import { TaskSessionSidebar } from "@/components/workspace/task-session-sidebar";
import { UserSettingsPanel } from "@/components/workspace/user-settings-panel";
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
import { clearStoredToken, getStoredToken } from "@/lib/auth";
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
  const [isSettingsPanelOpen, setIsSettingsPanelOpen] = useState(false);
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

  if (!token || !currentUser || !workspace || !task) {
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
          activeSessionId={activeSession?.id ?? ""}
          onCreateSession={() => setIsCreateSessionOpen(true)}
          onOpenSettings={() => setIsSettingsPanelOpen(true)}
          onSelectSession={(sessionId) => {
            setErrorMessage("");
            setMessages([]);
            setActiveSessionId(sessionId);
          }}
          sessions={sessions}
          task={task}
          user={currentUser}
        />

        {activeSession ? (
          <AgentColumn session={activeSession} />
        ) : (
          <div className="rounded-[32px] border border-line bg-paper p-6 shadow-panel">
            <div className="space-y-3">
              <p className="text-xs uppercase tracking-[0.22em] text-ink/42">参与协作</p>
              <h2 className="font-display text-2xl text-ink">先创建一个会话</h2>
              <p className="text-sm leading-7 text-ink/58">单聊模式下，Galaxy 和 Aries 都作为平级 Agent 供你选择。创建后，这里会显示当前会话绑定的 Agent。</p>
            </div>
          </div>
        )}

        {activeSession ? (
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
        ) : (
          <div className="flex min-h-0 flex-col overflow-hidden rounded-[32px] border border-line bg-paper shadow-panel">
            <div className="border-b border-line px-7 py-5">
              <h2 className="font-display text-3xl text-ink">还没有会话</h2>
              <p className="mt-2 text-sm leading-7 text-ink/58">点击左侧 `+` 新建会话，再选择 Galaxy 或 Aries 开始当前任务的单聊协作。</p>
            </div>
            <div className="flex flex-1 items-center justify-center px-7 py-6">
              <div className="max-w-md space-y-3 text-center">
                <p className="text-xs uppercase tracking-[0.3em] text-ink/40">Ready</p>
                <h2 className="font-display text-3xl text-ink">先选一个 Agent</h2>
                <p className="text-sm leading-7 text-ink/58">这次热修后，任务不会自动生成默认对话。会话由你主动创建，后续群聊模式再引入主 Agent 和子 Agent 的概念。</p>
              </div>
            </div>
          </div>
        )}
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

              setSessions((current) => [createdSession, ...current]);
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

      <UserSettingsPanel
        isOpen={isSettingsPanelOpen}
        onClose={() => setIsSettingsPanelOpen(false)}
        onLogout={() => {
          clearStoredToken();
          router.push("/login");
        }}
        user={currentUser}
      />
    </main>
  );
}
