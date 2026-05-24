// Date: 2026-05-25
// Author: XinYang Li

import * as React from "react";

import { cn } from "@/lib/utils";

/**
 * Renders a shared textarea for task descriptions and chat input.
 * @param props Standard HTML textarea props for multiline content.
 * @returns The styled textarea element.
 */
export function Textarea({ className, ...props }: React.TextareaHTMLAttributes<HTMLTextAreaElement>): JSX.Element {
  return (
    <textarea
      className={cn(
        "min-h-28 w-full rounded-3xl border border-line bg-white px-4 py-3 text-sm text-ink outline-none transition placeholder:text-ink/35 focus:border-pine/50 focus:ring-4 focus:ring-pine/10",
        className,
      )}
      {...props}
    />
  );
}
