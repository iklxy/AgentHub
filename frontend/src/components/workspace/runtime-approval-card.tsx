// Date: 2026-05-30
// Author: XinYang Li

"use client";

import { useState } from "react";
import type { PermissionRequest } from "@/lib/api";

const TOOL_LABELS: Record<string, string> = {
  Bash: "Bash",
  Write: "写入文件",
  Edit: "编辑文件",
  MultiEdit: "批量编辑",
};

function CheckIcon() {
  return (
    <svg fill="none" height="16" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" viewBox="0 0 16 16" width="16">
      <path d="M3 8.5l3.5 3.5L13 4" />
    </svg>
  );
}

function XIcon() {
  return (
    <svg fill="none" height="16" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" viewBox="0 0 16 16" width="16">
      <path d="M3 3l10 10M13 3l-10 10" />
    </svg>
  );
}

/**
 * Renders a compact permission approval card for one agent tool request.
 * Uses icon-only approve/deny buttons per the project's minimalist design language.
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

  const toolLabel = TOOL_LABELS[permission.toolName] ?? permission.toolName;
  const command = permission.input?.command as string | undefined;
  const filePath = permission.input?.file_path as string | undefined;

  return (
    <div className="mx-7 mb-5 overflow-hidden rounded-2xl border border-ember/15 bg-amber-50/40">
      <div className="flex items-start gap-4 px-5 py-4">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="rounded-full border border-ember/25 bg-ember/8 px-2.5 py-0.5 text-[11px] font-semibold uppercase tracking-[0.12em] text-ember">
              {toolLabel}
            </span>
            {filePath && (
              <span className="truncate text-[11px] tracking-[0.06em] text-ink/42">{filePath}</span>
            )}
          </div>
          {command && (
            <pre className="mt-2.5 overflow-x-auto rounded-lg bg-black/6 px-3 py-2 text-xs text-ink/72">
              <code>{command}</code>
            </pre>
          )}
        </div>

        {showDenyInput ? (
          <div className="flex shrink-0 flex-col items-end gap-2">
            <textarea
              autoFocus
              className="w-48 resize-none rounded-xl border border-line bg-white px-3 py-2 text-xs text-ink placeholder:text-ink/32 focus:border-ember/40 focus:outline-none"
              onChange={(event) => setDenyReason(event.target.value)}
              placeholder="拒绝原因（可选）"
              rows={2}
              value={denyReason}
            />
            <div className="flex gap-2">
              <button
                className="flex h-9 items-center gap-1.5 rounded-full border border-line bg-white px-3.5 text-xs font-semibold text-ink/62 transition hover:border-pine/30 hover:text-ink"
                disabled={isSubmitting}
                onClick={() => setShowDenyInput(false)}
                type="button"
              >
                返回
              </button>
              <button
                className="flex h-9 items-center gap-1.5 rounded-full border border-ember/40 bg-ember px-3.5 text-xs font-semibold text-white transition hover:bg-ember/85 disabled:opacity-50"
                disabled={isSubmitting}
                onClick={() => onDeny(denyReason)}
                type="button"
              >
                <XIcon />
                {isSubmitting ? "提交中" : "确认"}
              </button>
            </div>
          </div>
        ) : (
          <div className="flex shrink-0 gap-2">
            <button
              className="flex h-10 w-10 items-center justify-center rounded-full border border-ember/25 bg-white text-ember transition hover:border-ember/50 hover:bg-ember/5 disabled:opacity-40"
              disabled={isSubmitting}
              onClick={() => setShowDenyInput(true)}
              title="拒绝"
              type="button"
            >
              <XIcon />
            </button>
            <button
              className="flex h-10 w-10 items-center justify-center rounded-full border border-pine/30 bg-pine text-paper shadow-sm transition hover:-translate-y-0.5 hover:bg-[#164939] hover:shadow-md disabled:opacity-40"
              disabled={isSubmitting}
              onClick={() => onAllow(command ? { command } : {})}
              title="允许"
              type="button"
            >
              {isSubmitting ? (
                <span className="h-4 w-4 animate-spin rounded-full border-2 border-paper/40 border-t-paper" />
              ) : (
                <CheckIcon />
              )}
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
