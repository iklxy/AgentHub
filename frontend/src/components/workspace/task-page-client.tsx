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
import { WorkspaceModalShell } from "@/components/workspace/workspace-modal-shell";
import { Button } from "@/components/ui/button";
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
  quoteMessage,
  replyMessage,
  updateSession,
} from "@/lib/api";
import { clearStoredToken, getStoredToken } from "@/lib/auth";
import type { AgentOption, Message, Session, Task, User, Workspace } from "@/types/domain";

/**
 * Sorts sessions for the left sidebar with pinned sessions first and recent activity second.
 * @param items The session list that should be reordered for display.
 * @returns The sorted session list copy.
 */
function sortSessions(items: Session[]): Session[] {
  return [...items].sort((left, right) => {
    if (left.isPinned !== right.isPinned) {
      return left.isPinned ? -1 : 1;
    }

    const leftTime = new Date(left.lastActiveAt).getTime();
    const rightTime = new Date(right.lastActiveAt).getTime();
    return rightTime - leftTime;
  });
}

/**
 * Builds one local optimistic user message so the transcript can render immediately before the backend round finishes.
 * @param taskId The current task identifier.
 * @param sessionId The active session identifier that owns the message.
 * @param content The trimmed user message content.
 * @returns The local optimistic message entity rendered before the server response arrives.
 */
function buildOptimisticUserMessage(taskId: string, sessionId: string, content: string): Message {
  return {
    id: `local-user-${Date.now()}`,
    taskId,
    sessionId,
    role: "user",
    content,
    timeLabel: "刚刚",
    replyToMessageId: "",
  };
}

/**
 * Normalizes a message excerpt for compact quote and reply previews.
 * @param content The raw message content.
 * @returns The shortened single-line preview text.
 */
function buildMessagePreview(content: string): string {
  return content.replace(/\s+/g, " ").trim().slice(0, 120);
}

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
  const [pendingArchiveSession, setPendingArchiveSession] = useState<Session | null>(null);
  const [messageOperation, setMessageOperation] = useState<{
    mode: "quote" | "reply";
    messages: Message[];
  } | null>(null);
  const [searchKeyword, setSearchKeyword] = useState("");
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
        setSessions(sortSessions(loadedSessions));
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

  const filteredSessions = sessions.filter((session) => {
    const keyword = searchKeyword.trim().toLowerCase();
    if (!keyword) {
      return true;
    }

    return (
      session.title.toLowerCase().includes(keyword) ||
      session.primaryAgentName.toLowerCase().includes(keyword)
    );
  });
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
          isUpdatingSession={isPending}
          onArchiveSession={(sessionId) => {
            const targetSession = sessions.find((item) => item.id === sessionId) ?? null;
            setPendingArchiveSession(targetSession);
          }}
          onCreateSession={() => setIsCreateSessionOpen(true)}
          onOpenSettings={() => setIsSettingsPanelOpen(true)}
          onSearchChange={setSearchKeyword}
          onSelectSession={(sessionId) => {
            setErrorMessage("");
            setMessages([]);
            setMessageOperation(null);
            setActiveSessionId(sessionId);
          }}
          onTogglePin={(sessionId, nextPinned) => {
            setErrorMessage("");

            startTransition(async () => {
              try {
                const updatedSession = await updateSession(token, sessionId, { isPinned: nextPinned });
                setSessions((current) =>
                  sortSessions(current.map((item) => (item.id === sessionId ? updatedSession : item))),
                );
              } catch (error) {
                setErrorMessage(error instanceof Error ? error.message : "更新置顶状态失败");
              }
            });
          }}
          searchKeyword={searchKeyword}
          sessions={filteredSessions}
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
              const currentMessageOperation = messageOperation;
              const optimisticMessage = buildOptimisticUserMessage(task.id, activeSession.id, content);

              setMessages((current) => [...current, optimisticMessage]);
              setMessageOperation(null);

              startTransition(async () => {
                try {
                  const response =
                    currentMessageOperation?.mode === "quote"
                      ? await quoteMessage(token, {
                          content,
                          messageIds: currentMessageOperation.messages.map((message) => message.id),
                        })
                      : currentMessageOperation?.mode === "reply"
                        ? await replyMessage(token, currentMessageOperation.messages[0].id, { content })
                        : await createMessage(token, task.id, {
                            sessionId: activeSession.id,
                            content,
                          });

                  setMessages((current) => [
                    ...current.filter((item) => item.id !== optimisticMessage.id),
                    response.userMessage,
                    response.assistantMessage,
                  ]);
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
                    sortSessions(
                      current.map((item) =>
                        item.id === activeSession.id
                          ? {
                              ...item,
                              startedAt: item.startedAt || new Date().toISOString(),
                              lastActiveAt: new Date().toISOString(),
                              lastActiveAtLabel: "刚刚",
                              lastMessagePreview: response.assistantMessage.content,
                            }
                          : item,
                      ),
                    ),
                  );
                } catch (error) {
                  setMessages((current) => current.filter((item) => item.id !== optimisticMessage.id));
                  setErrorMessage(error instanceof Error ? error.message : "发送消息失败");
                }
              });
            }}
            messageOperation={messageOperation}
            onClearMessageOperation={() => setMessageOperation(null)}
            onQuoteMessage={(message) => {
              setMessageOperation((current) => {
                const normalizedMessage = {
                  ...message,
                  content: buildMessagePreview(message.content),
                };

                if (!current || current.mode !== "quote") {
                  return {
                    mode: "quote",
                    messages: [normalizedMessage],
                  };
                }

                if (current.messages.some((item) => item.id === message.id)) {
                  return current;
                }

                return {
                  mode: "quote",
                  messages: [...current.messages, normalizedMessage],
                };
              });
            }}
            onReplyMessage={(message) => {
              setMessageOperation({
                mode: "reply",
                messages: [{
                  ...message,
                  content: buildMessagePreview(message.content),
                }],
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

              setSessions((current) => sortSessions([createdSession, ...current]));
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

      <WorkspaceModalShell
        isOpen={pendingArchiveSession !== null}
        onClose={() => setPendingArchiveSession(null)}
        title="确认删除 Session"
      >
        <div className="space-y-5 px-6 py-6">
          <p className="text-sm leading-7 text-ink/62">
            {pendingArchiveSession
              ? `删除后，这个会话会从当前列表移除。请确认是否删除「${pendingArchiveSession.title}」。`
              : ""}
          </p>

          <div className="flex justify-end gap-3">
            <Button onClick={() => setPendingArchiveSession(null)} type="button" variant="secondary">
              取消
            </Button>
            <Button
              disabled={isPending || pendingArchiveSession === null}
              onClick={() => {
                if (!pendingArchiveSession) {
                  return;
                }

                setErrorMessage("");

                startTransition(async () => {
                  try {
                    await updateSession(token, pendingArchiveSession.id, { isArchived: true });

                    let nextSessions: Session[] = [];
                    setSessions((current) => {
                      nextSessions = current.filter((item) => item.id !== pendingArchiveSession.id);
                      return nextSessions;
                    });

                    if (activeSessionId === pendingArchiveSession.id) {
                      const nextSessionId = nextSessions[0]?.id ?? "";
                      setActiveSessionId(nextSessionId);
                      setMessages([]);
                    }

                    setPendingArchiveSession(null);
                  } catch (error) {
                    setErrorMessage(error instanceof Error ? error.message : "删除会话失败");
                  }
                });
              }}
              type="button"
            >
              {isPending ? "删除中..." : "确认删除"}
            </Button>
          </div>
        </div>
      </WorkspaceModalShell>
    </main>
  );
}
