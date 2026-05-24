// Date: 2026-05-25
// Author: XinYang Li

import { cn } from "@/lib/utils";
import type { TaskStatus } from "@/types/domain";

const statusStyles: Record<TaskStatus, string> = {
  idle: "border-line bg-white text-ink/70",
  running: "border-pine/20 bg-pine/10 text-pine",
  failed: "border-ember/20 bg-ember/10 text-ember",
  completed: "border-moss bg-moss/60 text-ink",
};

const statusLabels: Record<TaskStatus, string> = {
  idle: "待开始",
  running: "进行中",
  failed: "失败",
  completed: "已完成",
};

/**
 * Renders a status badge for task state communication.
 * @param props.status The logical task status to translate into UI language.
 * @param props.className Optional extra class names for layout overrides.
 * @returns The styled status badge.
 */
export function StatusBadge({ status, className }: { status: TaskStatus; className?: string }): JSX.Element {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full border px-3 py-1 text-xs font-medium tracking-[0.12em]",
        statusStyles[status],
        className,
      )}
    >
      {statusLabels[status]}
    </span>
  );
}
