// Date: 2026-05-25
// Author: XinYang Li

import * as React from "react";

import { cn } from "@/lib/utils";

/**
 * Renders a shared text input with the workspace light theme.
 * @param props Standard HTML input props for forms and dialogs.
 * @returns The styled input element.
 */
export function Input({ className, ...props }: React.InputHTMLAttributes<HTMLInputElement>): JSX.Element {
  return (
    <input
      className={cn(
        "h-12 w-full rounded-2xl border border-line bg-white px-4 text-sm text-ink outline-none transition placeholder:text-ink/35 focus:border-pine/50 focus:ring-4 focus:ring-pine/10",
        className,
      )}
      {...props}
    />
  );
}
