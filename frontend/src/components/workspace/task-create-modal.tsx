// Date: 2026-05-25
// Author: XinYang Li

"use client";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { WorkspaceModalShell } from "@/components/workspace/workspace-modal-shell";

/**
 * Renders the task creation dialog triggered from the sidebar plus action.
 * @param props.description The task description draft value.
 * @param props.isOpen Controls whether the dialog is visible.
 * @param props.isSubmitting Indicates whether the create request is running.
 * @param props.onClose Closes the dialog without creating a task.
 * @param props.onDescriptionChange Updates the task description draft.
 * @param props.onSubmit Creates the task with the current draft values.
 * @param props.onTitleChange Updates the task title draft.
 * @param props.title The task title draft value.
 * @returns The task creation dialog.
 */
export function TaskCreateModal({
  description,
  isOpen,
  isSubmitting,
  onClose,
  onDescriptionChange,
  onSubmit,
  onTitleChange,
  title,
}: {
  description: string;
  isOpen: boolean;
  isSubmitting: boolean;
  onClose: () => void;
  onDescriptionChange: (value: string) => void;
  onSubmit: () => void;
  onTitleChange: (value: string) => void;
  title: string;
}): JSX.Element | null {
  return (
    <WorkspaceModalShell isOpen={isOpen} onClose={onClose} title="新建任务">
      <form
        className="space-y-5 px-6 py-6"
        onSubmit={(event) => {
          event.preventDefault();
          onSubmit();
        }}
      >
        <div className="space-y-2">
          <p className="text-sm font-semibold text-ink">任务标题</p>
          <Input onChange={(event) => onTitleChange(event.target.value)} placeholder="例如：整理 v0.1 发布方案" value={title} />
        </div>
        <div className="space-y-2">
          <p className="text-sm font-semibold text-ink">任务描述</p>
          <Textarea
            className="min-h-36"
            onChange={(event) => onDescriptionChange(event.target.value)}
            placeholder="写明目标、上下文、需要产出的结果物，以及当前推进状态"
            value={description}
          />
        </div>
        <div className="flex justify-end gap-3 pt-2">
          <Button onClick={onClose} type="button" variant="secondary">
            取消
          </Button>
          <Button disabled={isSubmitting} type="submit">
            {isSubmitting ? "创建中..." : "创建任务"}
          </Button>
        </div>
      </form>
    </WorkspaceModalShell>
  );
}
