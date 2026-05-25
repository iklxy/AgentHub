// Date: 2026-05-25
// Author: XinYang Li

"use client";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { WorkspaceModalShell } from "@/components/workspace/workspace-modal-shell";

/**
 * Renders the workspace editing dialog triggered from the sidebar.
 * @param props.description The workspace description draft value.
 * @param props.isOpen Controls whether the dialog is visible.
 * @param props.isSubmitting Indicates whether the save request is in flight.
 * @param props.name The workspace name draft value.
 * @param props.onClose Closes the dialog without saving.
 * @param props.onDescriptionChange Updates the description draft.
 * @param props.onNameChange Updates the name draft.
 * @param props.onSubmit Persists the workspace metadata changes.
 * @returns The workspace edit dialog.
 */
export function WorkspaceEditModal({
  description,
  isOpen,
  isSubmitting,
  name,
  onClose,
  onDescriptionChange,
  onNameChange,
  onSubmit,
}: {
  description: string;
  isOpen: boolean;
  isSubmitting: boolean;
  name: string;
  onClose: () => void;
  onDescriptionChange: (value: string) => void;
  onNameChange: (value: string) => void;
  onSubmit: () => void;
}): JSX.Element | null {
  return (
    <WorkspaceModalShell isOpen={isOpen} onClose={onClose} title="编辑工作区">
      <form
        className="space-y-5 px-6 py-6"
        onSubmit={(event) => {
          event.preventDefault();
          onSubmit();
        }}
      >
        <div className="space-y-2">
          <p className="text-sm font-semibold text-ink">工作区名称</p>
          <Input onChange={(event) => onNameChange(event.target.value)} placeholder="例如：产品协作中台" value={name} />
        </div>
        <div className="space-y-2">
          <p className="text-sm font-semibold text-ink">工作区描述</p>
          <Textarea
            className="min-h-36"
            onChange={(event) => onDescriptionChange(event.target.value)}
            placeholder="说明这个工作区当前服务的团队、任务范围与协作方式"
            value={description}
          />
        </div>
        <div className="flex justify-end gap-3 pt-2">
          <Button onClick={onClose} type="button" variant="secondary">
            取消
          </Button>
          <Button disabled={isSubmitting} type="submit">
            {isSubmitting ? "保存中..." : "保存工作区"}
          </Button>
        </div>
      </form>
    </WorkspaceModalShell>
  );
}
