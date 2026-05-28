// Date: 2026-05-27
// Author: XinYang Li

"use client";

import { Check, Copy, CornerDownLeft, Pin, Quote, RotateCcw } from "lucide-react";
import { useState } from "react";

import { AttachmentStrip } from "@/components/workspace/attachment-strip";
import { Button } from "@/components/ui/button";
import { MarkdownContent } from "@/components/workspace/markdown-content";
import { cn } from "@/lib/utils";
import type { Message } from "@/types/domain";

type MessageBlock =
  | {
      type: "text";
      value: string;
    }
  | {
      type: "code";
      value: string;
      language: string;
    };

/**
 * Splits message content into prose and fenced code blocks.
 * Params:
 * - content: the raw message content that may contain fenced code blocks.
 * Returns:
 * - the ordered message block list used by the bubble renderer.
 */
function splitMessageContent(content: string): MessageBlock[] {
  const parts = content.split("```");

  return parts.map((value, index) => {
    if (index % 2 === 0) {
      return {
        type: "text",
        value,
      };
    }

    const [firstLine, ...restLines] = value.split("\n");
    const language = firstLine.trim();
    const body = restLines.join("\n").trim() || value.trim();

    return {
      type: "code",
      value: body,
      language,
    };
  });
}

/**
 * Writes one code block into the system clipboard.
 * Params:
 * - value: the code block content that should be copied.
 * Returns:
 * - a promise that resolves after the clipboard write finishes.
 */
async function copyCodeBlock(value: string): Promise<void> {
  if (typeof navigator !== "undefined" && navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(value);
      return;
    } catch {
      // Fall through to the legacy copy path when clipboard permissions fail.
    }
  }

  const textarea = document.createElement("textarea");
  textarea.value = value;
  textarea.setAttribute("readonly", "true");
  textarea.style.position = "fixed";
  textarea.style.opacity = "0";
  textarea.style.pointerEvents = "none";
  document.body.appendChild(textarea);
  textarea.select();
  textarea.setSelectionRange(0, textarea.value.length);

  const copied = document.execCommand("copy");
  document.body.removeChild(textarea);

  if (!copied) {
    throw new Error("copy failed");
  }
}

/**
 * Renders one chat bubble with role-sensitive layout, markdown support, and copyable code blocks.
 * Params:
 * - props.message: the message entity rendered in the transcript.
 * - props.onQuote: optional callback that starts one quote action from the current message.
 * - props.onReply: optional callback that starts one reply action from the current message.
 * - props.onRegenerate: optional callback that reruns the selected assistant message.
 * - props.onTogglePin: optional callback that toggles the pinned state for the current message.
 * - props.isRegenerating: whether the regenerate action is currently blocked by an in-flight request.
 * - props.isPinUpdating: whether the pin action is currently blocked by an in-flight request.
 * Returns:
 * - the message bubble element.
 */
export function MessageBubble({
  message,
  isRegenerating = false,
  isPinUpdating = false,
  onQuote,
  onReply,
  onRegenerate,
  onTogglePin,
}: {
  message: Message;
  isRegenerating?: boolean;
  isPinUpdating?: boolean;
  onQuote?: (message: Message) => void;
  onReply?: (message: Message) => void;
  onRegenerate?: (message: Message) => void;
  onTogglePin?: (message: Message, nextPinned: boolean) => void;
}): JSX.Element {
  const [copiedBlockKey, setCopiedBlockKey] = useState<string | null>(null);
  const isUser = message.role === "user";
  const isSystem = message.role === "system";
  const blocks = splitMessageContent(message.content);

  if (isSystem) {
    return (
      <div className="mx-auto max-w-xl rounded-full border border-line bg-white/80 px-4 py-2 text-center text-xs text-ink/58">
        {message.content}
      </div>
    );
  }

  return (
    <div className={cn("flex animate-rise", isUser ? "justify-end" : "justify-start")}>
      <article
        className={cn(
          "max-w-2xl rounded-[28px] px-5 py-4 shadow-panel transition",
          isUser ? "bg-pine text-paper" : "border border-line bg-white text-ink",
        )}
      >
        <div className="space-y-3">
          <AttachmentStrip attachments={message.attachments ?? []} compact />
          {blocks.map((block, index) => {
            const blockKey = `${message.id}-${index}`;

            if (block.type === "code") {
              return (
                <div className="relative" key={blockKey}>
                  {block.language ? (
                    <p
                      className={cn(
                        "absolute left-4 top-3 z-10 text-[11px] uppercase tracking-[0.16em]",
                        isUser ? "text-paper/62" : "text-ink/38",
                      )}
                    >
                      {block.language}
                    </p>
                  ) : null}
                  <Button
                    aria-label="复制代码块"
                    className={cn(
                      "absolute right-2 top-2 z-10 h-9 w-9 rounded-full p-0 shadow-sm",
                      isUser ? "bg-white/12 text-paper hover:bg-white/18" : "bg-white text-ink hover:bg-paper",
                    )}
                    onClick={() => {
                      copyCodeBlock(block.value)
                        .then(() => {
                          setCopiedBlockKey(blockKey);
                          window.setTimeout(() => {
                            setCopiedBlockKey((current) => (current === blockKey ? null : current));
                          }, 1200);
                        })
                        .catch(() => {
                          setCopiedBlockKey(null);
                        });
                    }}
                    size="icon"
                    type="button"
                    variant="ghost"
                  >
                    {copiedBlockKey === blockKey ? <Check className="h-3.5 w-3.5" /> : <Copy className="h-3.5 w-3.5" />}
                  </Button>
                  <pre
                    className={cn(
                      "overflow-x-auto rounded-2xl px-4 pb-3 pt-12 pr-14 text-xs leading-6",
                      isUser ? "bg-black/15 text-paper" : "bg-mist text-ink",
                    )}
                  >
                    <code>{block.value}</code>
                  </pre>
                </div>
              );
            }

            return (
              <MarkdownContent
                content={block.value}
                inlineCodeClassName={isUser ? "rounded-md bg-white/12 px-1.5 py-0.5 font-mono text-[13px] text-paper" : undefined}
                key={blockKey}
                paragraphClassName="whitespace-pre-wrap text-sm leading-7"
                tableClassName={isUser ? "overflow-x-auto rounded-2xl border border-white/16 bg-white/8" : undefined}
              />
            );
          })}
        </div>
        <div className="mt-4 flex items-center justify-between gap-3">
          <div className="flex items-center gap-2">
            {onQuote ? (
              <Button
                className={cn(isUser ? "text-paper/72 hover:bg-white/12 hover:text-paper" : "text-ink/48 hover:text-ink")}
                onClick={() => onQuote(message)}
                size="sm"
                type="button"
                variant="ghost"
              >
                <Quote className="h-3.5 w-3.5" />
              </Button>
            ) : null}
            {onReply ? (
              <Button
                className={cn(isUser ? "text-paper/72 hover:bg-white/12 hover:text-paper" : "text-ink/48 hover:text-ink")}
                onClick={() => onReply(message)}
                size="sm"
                type="button"
                variant="ghost"
              >
                <CornerDownLeft className="h-3.5 w-3.5" />
              </Button>
            ) : null}
            {onRegenerate && message.role === "assistant" ? (
              <Button
                className={cn(isUser ? "text-paper/72 hover:bg-white/12 hover:text-paper" : "text-ink/48 hover:text-ink")}
                disabled={isRegenerating}
                onClick={() => onRegenerate(message)}
                size="sm"
                type="button"
                variant="ghost"
              >
                <RotateCcw className="h-3.5 w-3.5" />
              </Button>
            ) : null}
            {onTogglePin ? (
              <Button
                className={cn(
                  message.isPinned
                    ? isUser
                      ? "bg-white/14 text-paper hover:bg-white/20 hover:text-paper"
                      : "bg-pine/10 text-pine hover:bg-pine/14"
                    : isUser
                      ? "text-paper/72 hover:bg-white/12 hover:text-paper"
                      : "text-ink/48 hover:text-ink",
                )}
                disabled={isPinUpdating}
                onClick={() => onTogglePin(message, !message.isPinned)}
                size="sm"
                type="button"
                variant="ghost"
              >
                <Pin className="h-3.5 w-3.5" />
              </Button>
            ) : null}
          </div>
          <p className={cn("text-[11px] uppercase tracking-[0.16em]", isUser ? "text-paper/60" : "text-ink/35")}>
            {message.timeLabel}
          </p>
        </div>
      </article>
    </div>
  );
}
