// Date: 2026-05-25
// Author: XinYang Li

import type { ReactNode } from "react";

import { cn } from "@/lib/utils";

/**
 * Wraps content inside a shared elevated surface.
 * @param props.children The nested UI content rendered inside the panel.
 * @param props.className Optional class names for layout adjustments.
 * @returns The styled panel container.
 */
export function Panel({ children, className }: { children: ReactNode; className?: string }): JSX.Element {
  return <section className={cn("rounded-[28px] border border-line bg-paper shadow-panel", className)}>{children}</section>;
}
