// Date: 2026-05-30
// Author: XinYang Li

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Panel } from "@/components/ui/panel";
import type { PermissionRequest } from "@/lib/api";

const TOOL_LABELS: Record<string, string> = {
  Bash: "Bash 命令审批",
  Write: "文件写入审批",
  Edit: "文件编辑审批",
  MultiEdit: "批量编辑审批",
};

/**
 * Renders a permission approval card for one agent tool request.
 * @param props.permission The permission request from the backend.
 * @param props.onAllow Called when the user approves, with optionally modified input.
 * @param props.onDeny Called when the user denies.
 * @param props.isSubmitting Whether a response is being sent.
 * @returns The approval card component.
 */
export function RuntimeApprovalCard({
  permission,
  onAllow,
  onDeny,
  isSubmitting,
}: {
  permission: PermissionRequest;
  onAllow: (updatedInput: Record<string, unknown>) => void;
  onDeny: (message: string) => void;
  isSubmitting: boolean;
}): JSX.Element {
  const [denyReason, setDenyReason] = useState("");
  const [showDenyInput, setShowDenyInput] = useState(false);

  const toolLabel = TOOL_LABELS[permission.toolName] ?? `${permission.toolName} 工具审批`;
  const command = permission.input?.command as string | undefined;

  return (
    <Panel className="mx-7 mb-5 rounded-2xl border-ember/20 bg-ember/5 p-5 shadow-sm">
      <div className="space-y-4">
        <div className="flex items-start justify-between gap-3">
          <div>
            <h3 className="text-sm font-semibold text-ink">{toolLabel}</h3>
            {command && (
              <pre className="mt-2 overflow-x-auto rounded-lg bg-black/8 px-3 py-2 text-xs text-ink/78">
                <code>{command}</code>
              </pre>
            )}
            {permission.toolName === "Write" && typeof permission.input?.file_path === "string" && (
              <p className="mt-1 text-xs text-ink/58">
                文件路径：{String(permission.input.file_path)}
              </p>
            )}
            {permission.toolName === "Edit" && typeof permission.input?.file_path === "string" && (
              <p className="mt-1 text-xs text-ink/58">
                编辑文件：{String(permission.input.file_path)}
              </p>
            )}
          </div>
        </div>

        {showDenyInput ? (
          <div className="space-y-3">
            <textarea
              className="w-full rounded-lg border border-line bg-white px-3 py-2 text-sm text-ink placeholder:text-ink/38 focus:outline-none focus:ring-2 focus:ring-pine/30"
              onChange={(event) => setDenyReason(event.target.value)}
              placeholder="输入拒绝原因（可选）"
              rows={2}
              value={denyReason}
            />
            <div className="flex justify-end gap-2">
              <Button
                disabled={isSubmitting}
                onClick={() => setShowDenyInput(false)}
                size="sm"
                variant="secondary"
              >
                取消
              </Button>
              <Button
                disabled={isSubmitting}
                onClick={() => onDeny(denyReason)}
                size="sm"
              >
                {isSubmitting ? "提交中..." : "确认拒绝"}
              </Button>
            </div>
          </div>
        ) : (
          <div className="flex justify-end gap-2">
            <Button
              disabled={isSubmitting}
              onClick={() => setShowDenyInput(true)}
              size="sm"
              variant="secondary"
            >
              拒绝
            </Button>
            <Button
              disabled={isSubmitting}
              onClick={() => onAllow(command ? { command } : {})}
              size="sm"
            >
              {isSubmitting ? "提交中..." : "允许"}
            </Button>
          </div>
        )}
      </div>
    </Panel>
  );
}
