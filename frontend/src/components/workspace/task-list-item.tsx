// Date: 2026-05-25
// Author: XinYang Li

import Link from "next/link";

import { StatusBadge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import type { Task } from "@/types/domain";

/**
 * Renders one task row for the workspace and task navigation lists.
 * @param props.task The task data rendered inside the list item.
 * @param props.isActive Whether the current route points to this task.
 * @param props.href The target route for the task detail page.
 * @returns The linked task list item.
 */
export function TaskListItem({ task, isActive, href }: { task: Task; isActive: boolean; href: string }): JSX.Element {
  return (
    <Link
      href={href}
      className={cn(
        "group block rounded-[24px] border px-4 py-4 transition duration-200 hover:-translate-y-0.5 hover:border-pine/30 hover:bg-white",
        isActive ? "border-pine/30 bg-white shadow-panel" : "border-transparent bg-transparent",
      )}
    >
      <div className="mb-3 flex items-start justify-between gap-4">
        <div className="space-y-1">
          <h3 className="text-sm font-semibold text-ink">{task.title}</h3>
          <p className="line-clamp-2 text-xs leading-6 text-ink/58">{task.description}</p>
        </div>
        <StatusBadge className="shrink-0" status={task.status} />
      </div>
      <p className="text-[11px] uppercase tracking-[0.18em] text-ink/34">{task.updatedAtLabel}</p>
    </Link>
  );
}
