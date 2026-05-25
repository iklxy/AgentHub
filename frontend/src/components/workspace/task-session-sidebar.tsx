// Date: 2026-05-25
// Author: XinYang Li

"use client";

import { Archive, ArrowLeft, Pin, Plus, Search, Settings2 } from "lucide-react";
import Link from "next/link";

import { Avatar } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Panel } from "@/components/ui/panel";
import { cn } from "@/lib/utils";
import type { Session, User } from "@/types/domain";

/**
 * Renders the left task sidebar as a session management list.
 * @param props.activeSessionId The currently selected session identifier.
 * @param props.isUpdatingSession Indicates whether one session action is in flight.
 * @param props.onArchiveSession Opens the archive confirmation flow for one session.
 * @param props.onCreateSession Opens the create-session dialog.
 * @param props.onOpenSettings Opens the user settings side panel from the task workspace footer.
 * @param props.onSearchChange Updates the local session search keyword.
 * @param props.onSelectSession Switches the active session in the task workspace.
 * @param props.onTogglePin Toggles the pinned state for one session.
 * @param props.searchKeyword The current local search keyword.
 * @param props.sessions The filtered task session list displayed in order.
 * @param props.user The current user snapshot shown in the footer area.
 * @returns The session management sidebar.
 */
export function TaskSessionSidebar({
  activeSessionId,
  isUpdatingSession,
  onArchiveSession,
  onCreateSession,
  onOpenSettings,
  onSearchChange,
  onSelectSession,
  onTogglePin,
  searchKeyword,
  sessions,
  user,
}: {
  activeSessionId: string;
  isUpdatingSession: boolean;
  onArchiveSession: (sessionId: string) => void;
  onCreateSession: () => void;
  onOpenSettings: () => void;
  onSearchChange: (value: string) => void;
  onSelectSession: (sessionId: string) => void;
  onTogglePin: (sessionId: string, nextPinned: boolean) => void;
  searchKeyword: string;
  sessions: Session[];
  user: User;
}): JSX.Element {
  return (
    <Panel className="flex h-full min-h-0 flex-col p-5">
      <div className="border-b border-line pb-5">
        <Link className="inline-flex items-center gap-2 text-sm text-ink/56 transition hover:text-pine" href="/workspace">
          <ArrowLeft className="h-4 w-4" />
          返回工作区
        </Link>

        <div className="mt-5 flex items-center justify-between gap-3">
          <div>
            <p className="text-xs uppercase tracking-[0.22em] text-ink/42">会话列表</p>
            <p className="mt-1 text-sm text-ink/56">{sessions.length} 个结果</p>
          </div>
          <Button aria-label="新建 session" onClick={onCreateSession} size="icon" type="button">
            <Plus className="h-4 w-4" />
          </Button>
        </div>

        <div className="relative mt-4">
          <Search className="pointer-events-none absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-ink/38" />
          <Input
            className="pl-11"
            onChange={(event) => onSearchChange(event.target.value)}
            placeholder="搜索会话名或 Agent"
            value={searchKeyword}
          />
        </div>
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
            <div className="flex items-start justify-between gap-3">
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                  <p className="truncate text-sm font-semibold text-ink">{session.title}</p>
                  {session.isPinned ? <Pin className="h-3.5 w-3.5 text-pine" /> : null}
                </div>
                <p className="mt-2 truncate text-xs leading-6 text-ink/56">
                  {session.primaryAgentName} · {session.createdAtLabel}创建
                </p>
              </div>

              <div className="flex shrink-0 items-center gap-1">
                <Button
                  aria-label={session.isPinned ? "取消置顶会话" : "置顶会话"}
                  disabled={isUpdatingSession}
                  onClick={(event) => {
                    event.stopPropagation();
                    onTogglePin(session.id, !session.isPinned);
                  }}
                  size="icon"
                  type="button"
                  variant="ghost"
                >
                  <Pin className={cn("h-4 w-4", session.isPinned ? "text-pine" : "text-ink/42")} />
                </Button>
                <Button
                  aria-label="删除会话"
                  disabled={isUpdatingSession}
                  onClick={(event) => {
                    event.stopPropagation();
                    onArchiveSession(session.id);
                  }}
                  size="icon"
                  type="button"
                  variant="ghost"
                >
                  <Archive className="h-4 w-4 text-ink/42" />
                </Button>
              </div>
            </div>
          </button>
        ))}

        {sessions.length === 0 ? (
          <div className="rounded-[24px] border border-dashed border-line bg-white/70 px-4 py-5 text-sm leading-7 text-ink/56">
            {searchKeyword.trim()
              ? "没有匹配的会话，换一个关键词试试。"
              : "当前还没有会话，点击右上角 `+` 新建session。"}
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
