// Date: 2026-05-25
// Author: XinYang Li

import { cn } from "@/lib/utils";

/**
 * Renders a lightweight avatar chip with initials fallback.
 * @param props.name The label used to derive the initials fallback.
 * @param props.className Optional layout overrides for size or placement.
 * @returns The avatar element.
 */
export function Avatar({ name, className }: { name: string; className?: string }): JSX.Element {
  return (
    <div
      className={cn(
        "flex h-11 w-11 items-center justify-center rounded-full border border-pine/20 bg-gradient-to-br from-mist to-moss text-sm font-semibold text-pine",
        className,
      )}
    >
      {name.slice(0, 1)}
    </div>
  );
}
