// Date: 2026-05-25
// Author: XinYang Li

"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState, useTransition } from "react";

import { AgentColumn } from "@/components/workspace/agent-column";
import { ChatColumn } from "@/components/workspace/chat-column";
import { WorkspaceSidebar } from "@/components/workspace/workspace-sidebar";
import { createMessage, getConversations, getCurrentUser, getMessages, getTask, getTasks, getWorkspace } from "@/lib/api";
import { getStoredToken } from "@/lib/auth";
import type { Conversation, Message, Task, User, Workspace } from "@/types/domain";

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
  const [conversation, setConversation] = useState<Conversation | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);

  useEffect(() => {
    const storedToken = getStoredToken();
    if (!storedToken) {
      router.replace("/login");
      return;
    }

    setToken(storedToken);

    startTransition(async () => {
      try {
        const [user, loadedWorkspace, loadedTasks, loadedTask, loadedConversations, loadedMessages] = await Promise.all([
          getCurrentUser(storedToken),
          getWorkspace(storedToken),
          getTasks(storedToken),
          getTask(storedToken, taskId),
          getConversations(storedToken, taskId),
          getMessages(storedToken, taskId),
        ]);

        setCurrentUser(user);
        setWorkspace(loadedWorkspace);
        setTasks(loadedTasks);
        setTask(loadedTask);
        setConversation(loadedConversations[0] ?? null);
        setMessages(loadedMessages);
      } catch (error) {
        setErrorMessage(error instanceof Error ? error.message : "加载任务失败");
      }
    });
  }, [router, taskId]);

  if (!token || !currentUser || !workspace || !task || !conversation) {
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
        <WorkspaceSidebar activeTaskId={task.id} tasks={tasks} user={currentUser} workspace={workspace} />
        <AgentColumn conversation={conversation} />
        <ChatColumn
          errorMessage={errorMessage}
          isSending={isPending}
          messages={messages}
          onSendMessage={(content) => {
            setErrorMessage("");

            startTransition(async () => {
              try {
                const response = await createMessage(token, task.id, { content });
                setMessages((current) => [...current, response.userMessage, response.assistantMessage]);
                setTask((current) => (current ? { ...current, status: "running", updatedAtLabel: "刚刚" } : current));
                setTasks((current) =>
                  current.map((item) => (item.id === task.id ? { ...item, status: "running", updatedAtLabel: "刚刚" } : item)),
                );
              } catch (error) {
                setErrorMessage(error instanceof Error ? error.message : "发送消息失败");
              }
            });
          }}
          task={task}
        />
      </div>
    </main>
  );
}
