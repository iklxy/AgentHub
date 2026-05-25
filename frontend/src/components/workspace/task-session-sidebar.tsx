// Date: 2026-05-25
// Author: XinYang Li

"use client";

import { ArrowLeft, Plus, Settings2 } from "lucide-react";
import Link from "next/link";

import { Avatar } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { StatusBadge } from "@/components/ui/badge";
import { Panel } from "@/components/ui/panel";
import { cn } from "@/lib/utils";
import type { Session, Task, User } from "@/types/domain";

/**
 * Renders the left task workspace sidebar with session navigation.
 * @param props.activeSessionId The currently selected session identifier.
 * @param props.onCreateSession Opens the create-session dialog.
 * @param props.onOpenSettings Opens the user settings side panel from the task workspace footer.
 * @param props.onSelectSession Switches the active session in the task workspace.
 * @param props.sessions The task session list displayed in order.
 * @param props.task The current task metadata shown at the top of the sidebar.
 * @param props.user The current user snapshot shown in the footer area.
 * @returns The session-aware task sidebar.
 */
export function TaskSessionSidebar({
  activeSessionId,
  onCreateSession,
  onOpenSettings,
  onSelectSession,
  sessions,
  task,
  user,
}: {
  activeSessionId: string;
  onCreateSession: () => void;
  onOpenSettings: () => void;
  onSelectSession: (sessionId: string) => void;
  sessions: Session[];
  task: Task;
  user: User;
}): JSX.Element {
  return (
    <Panel className="flex h-full min-h-0 flex-col p-5">
      <div className="border-b border-line pb-5">
        <Link className="inline-flex items-center gap-2 text-sm text-ink/56 transition hover:text-pine" href="/workspace">
          <ArrowLeft className="h-4 w-4" />
          返回工作区
        </Link>

        <div className="mt-5 space-y-3">
          <div className="flex items-center justify-between gap-3">
            <StatusBadge status={task.status} />
          </div>
          <h1 className="font-display text-3xl leading-tight text-ink">{task.title}</h1>
          <p className="text-sm leading-7 text-ink/60">{task.description || "当前任务还没有补充描述。"}</p>
        </div>
      </div>

      <div className="mt-5 flex items-center justify-between">
        <div>
          <p className="mt-1 text-sm text-ink/56">{sessions.length} 个会话</p>
        </div>
        <Button aria-label="新建 session" onClick={onCreateSession} size="icon" type="button">
          <Plus className="h-4 w-4" />
        </Button>
      </div>

      <div className="mt-5 flex-1 space-y-3 overflow-y-auto pr-1">
        {sessions.map((session) => (
          <button
            className={cn(
              "w-full rounded-[24px] border px-4 py-4 text-left transition duration-200 hover:-translate-y-0.5 hover:border-pine/30 hover:bg-white",
              session.id === activeSessionId ? "border-pine/30 bg-white shadow-panel" : "border-transparent bg-transparent",
            )}
            key={session.id}
            onClick={() => onSelectSession(session.id)}
            type="button"
          >
            <div className="flex items-start justify-between gap-4">
              <div className="min-w-0">
                <p className="truncate text-sm font-semibold text-ink">{session.title}</p>
                <div className="mt-2 flex flex-wrap gap-2">
                  <span className="rounded-full bg-white px-2 py-1 text-[11px] uppercase tracking-[0.14em] text-pine">
                    {session.primaryAgentName}
                  </span>
                </div>
              </div>
              <p className="shrink-0 text-[11px] uppercase tracking-[0.16em] text-ink/34">{session.lastActiveAtLabel}</p>
            </div>
            <p className="mt-3 line-clamp-2 text-xs leading-6 text-ink/56">
              {session.lastMessagePreview || "当前 session 还没有消息，发送第一条内容后会开始建立 Claude session。"}
            </p>
          </button>
        ))}
        {sessions.length === 0 ? (
          <div className="rounded-[24px] border border-dashed border-line bg-white/70 px-4 py-5 text-sm leading-7 text-ink/56">
            当前任务还没有会话，点击右上角 `+` 选择 Galaxy 或 Aries 后开始协作。
          </div>
        ) : null}
      </div>

      <div className="mt-5 flex items-center gap-3 border-t border-line pt-5">
        <Avatar name={user.username} />
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-semibold text-ink">{user.username}</p>
          <p className="truncate text-xs text-ink/52">{user.email}</p>
        </div>
        <Button aria-label="打开用户设置" onClick={onOpenSettings} size="icon" type="button" variant="ghost">
          <Settings2 className="h-4 w-4" />
        </Button>
      </div>
    </Panel>
  );
}
