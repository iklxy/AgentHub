// Date: 2026-05-25
// Author: XinYang Li

"use client";

import { X } from "lucide-react";
import { useEffect, type ReactNode } from "react";

import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

/**
 * Renders a shared fullscreen overlay used by workspace dialogs and side panels.
 * @param props.children The modal body content rendered inside the shell.
 * @param props.className Optional layout overrides for the inner surface.
 * @param props.isOpen Controls whether the shell should be mounted.
 * @param props.onClose Closes the shell when the overlay or close button is triggered.
 * @param props.title The visible title shown in the shell header.
 * @returns The overlay shell or null when closed.
 */
export function WorkspaceModalShell({
  children,
  className,
  isOpen,
  onClose,
  title,
}: {
  children: ReactNode;
  className?: string;
  isOpen: boolean;
  onClose: () => void;
  title: string;
}): JSX.Element | null {
  useEffect(() => {
    if (!isOpen) {
      return;
    }

    /**
     * Handles the Escape key to close the active shell.
     * @param event The keyboard event emitted by the browser.
     */
    function handleEscape(event: KeyboardEvent): void {
      if (event.key === "Escape") {
        onClose();
      }
    }

    window.addEventListener("keydown", handleEscape);
    return () => window.removeEventListener("keydown", handleEscape);
  }, [isOpen, onClose]);

  if (!isOpen) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-ink/30 px-4 py-8 backdrop-blur-sm">
      <button
        aria-label="关闭弹窗"
        className="absolute inset-0 cursor-default"
        onClick={onClose}
        type="button"
      />
      <section
        className={cn("relative z-10 w-full max-w-2xl rounded-[32px] border border-line bg-paper shadow-panel", className)}
      >
        <header className="flex items-start justify-between gap-4 border-b border-line px-6 py-5">
          <div>
            <p className="text-xs uppercase tracking-[0.22em] text-pine/64">Workspace</p>
            <h2 className="mt-2 font-display text-3xl text-ink">{title}</h2>
          </div>
          <Button aria-label="关闭弹窗" onClick={onClose} size="icon" type="button" variant="ghost">
            <X className="h-5 w-5" />
          </Button>
        </header>
        {children}
      </section>
    </div>
  );
}
