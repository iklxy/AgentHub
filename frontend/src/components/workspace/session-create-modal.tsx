// Date: 2026-05-25
// Author: XinYang Li

"use client";

import type { AgentOption } from "@/types/domain";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
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
          <select
            className="h-12 w-full rounded-3xl border border-line bg-white px-4 text-sm text-ink outline-none transition focus:border-pine/50 focus:ring-4 focus:ring-pine/10"
            onChange={(event) => onAgentChange(event.target.value)}
            value={primaryAgentId}
          >
            <option value="">请选择 Agent</option>
            {agents.map((agent) => (
              <option key={agent.id} value={agent.id}>
                {agent.name}
              </option>
            ))}
          </select>
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
