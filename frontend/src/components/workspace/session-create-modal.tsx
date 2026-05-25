// Date: 2026-05-25
// Author: XinYang Li

"use client";

import { Bot, Sparkles } from "lucide-react";

import type { AgentOption } from "@/types/domain";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import { WorkspaceModalShell } from "@/components/workspace/workspace-modal-shell";

/**
 * Renders the task session creation dialog used by the v0.2 task workspace.
 * @param props.agents The selectable single-chat agent options.
 * @param props.chatMode The current chat mode draft value.
 * @param props.isOpen Controls whether the dialog is visible.
 * @param props.isSubmitting Indicates whether the create request is currently in flight.
 * @param props.onAgentChange Updates the selected primary agent identifier.
 * @param props.onChatModeChange Updates the selected chat mode value.
 * @param props.onClose Closes the dialog without creating a session.
 * @param props.onSubmit Creates the session with the current draft values.
 * @param props.onTitleChange Updates the session title draft.
 * @param props.primaryAgentId The currently selected primary agent identifier.
 * @param props.title The current session title draft value.
 * @returns The session creation dialog.
 */
export function SessionCreateModal({
  agents,
  chatMode,
  isOpen,
  isSubmitting,
  onAgentChange,
  onChatModeChange,
  onClose,
  onSubmit,
  onTitleChange,
  primaryAgentId,
  title,
}: {
  agents: AgentOption[];
  chatMode: "single";
  isOpen: boolean;
  isSubmitting: boolean;
  onAgentChange: (value: string) => void;
  onChatModeChange: (value: "single") => void;
  onClose: () => void;
  onSubmit: () => void;
  onTitleChange: (value: string) => void;
  primaryAgentId: string;
  title: string;
}): JSX.Element | null {
  return (
    <WorkspaceModalShell isOpen={isOpen} onClose={onClose} title="新建 Session">
      <form
        className="space-y-5 px-6 py-6"
        onSubmit={(event) => {
          event.preventDefault();
          onSubmit();
        }}
      >
        <div className="space-y-2">
          <p className="text-sm font-semibold text-ink">Session 标题</p>
          <Input onChange={(event) => onTitleChange(event.target.value)} placeholder="例如：文档整理会话" value={title} />
        </div>

        <div className="space-y-2">
          <p className="text-sm font-semibold text-ink">聊天模式</p>
          <select
            className="h-12 w-full rounded-3xl border border-line bg-white px-4 text-sm text-ink outline-none transition focus:border-pine/50 focus:ring-4 focus:ring-pine/10"
            onChange={(event) => onChatModeChange(event.target.value as "single")}
            value={chatMode}
          >
            <option value="single">single</option>
          </select>
        </div>

        <div className="space-y-2">
          <p className="text-sm font-semibold text-ink">选择 Agent</p>
          <div className="grid gap-3">
            {agents.map((agent) => (
              <button
                className={cn(
                  "rounded-[24px] border px-4 py-4 text-left transition duration-200 hover:-translate-y-0.5",
                  primaryAgentId === agent.id
                    ? "border-pine/40 bg-pine/8 shadow-panel"
                    : "border-line bg-white hover:border-pine/24 hover:bg-mist",
                )}
                key={agent.id}
                onClick={() => onAgentChange(agent.id)}
                type="button"
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <div className="flex items-center gap-2">
                      <div className="flex h-9 w-9 items-center justify-center rounded-2xl bg-pine/10 text-pine">
                        <Bot className="h-4 w-4" />
                      </div>
                      <div>
                        <p className="text-sm font-semibold text-ink">{agent.name}</p>
                        <p className="text-xs uppercase tracking-[0.16em] text-ink/42">{agent.kind}</p>
                      </div>
                    </div>
                    <p className="mt-3 text-xs leading-6 text-ink/56">
                      {agent.name === "Galaxy" ? "适合通用协作、任务推进与连续对话。" : "适合文档整理、改写、归纳与内容加工。"}
                    </p>
                  </div>
                  {primaryAgentId === agent.id ? (
                    <span className="inline-flex items-center gap-1 rounded-full border border-pine/20 bg-pine/10 px-2.5 py-1 text-[11px] uppercase tracking-[0.14em] text-pine">
                      <Sparkles className="h-3 w-3" />
                      已选中
                    </span>
                  ) : null}
                </div>
              </button>
            ))}
          </div>
          <p className="text-xs leading-6 text-ink/48">Galaxy 和 Aries 都可以作为单聊模式下的可选 Agent，由你主动创建新的会话。</p>
        </div>

        <div className="flex justify-end gap-3 pt-2">
          <Button onClick={onClose} type="button" variant="secondary">
            取消
          </Button>
          <Button disabled={isSubmitting || !title.trim() || !primaryAgentId} type="submit">
            {isSubmitting ? "创建中..." : "创建 Session"}
          </Button>
        </div>
      </form>
    </WorkspaceModalShell>
  );
}
