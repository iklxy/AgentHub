// Date: 2026-05-25
// Author: XinYang Li

import { cn } from "@/lib/utils";
import type { Message } from "@/types/domain";

/**
 * Splits message content into prose and code fences for simple preview rendering.
 * @param content The raw message content that may contain fenced code blocks.
 * @returns Ordered content blocks tagged as text or code.
 */
function splitMessageContent(content: string): Array<{ type: "text" | "code"; value: string }> {
  const parts = content.split("```");

  return parts.map((value, index) => ({
    type: index % 2 === 0 ? "text" : "code",
    value: value.trim(),
  }));
}

/**
 * Renders one chat bubble with role-sensitive layout and simple code formatting.
 * @param props.message The message entity rendered in the transcript.
 * @returns The message bubble element.
 */
export function MessageBubble({ message }: { message: Message }): JSX.Element {
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
          {blocks.map((block, index) =>
            block.type === "code" ? (
              <pre
                className={cn(
                  "overflow-x-auto rounded-2xl px-4 py-3 text-xs leading-6",
                  isUser ? "bg-black/15 text-paper" : "bg-mist text-ink",
                )}
                key={`${message.id}-${index}`}
              >
                <code>{block.value}</code>
              </pre>
            ) : (
              <p className="whitespace-pre-wrap text-sm leading-7" key={`${message.id}-${index}`}>
                {block.value}
              </p>
            ),
          )}
        </div>
        <p className={cn("mt-3 text-[11px] uppercase tracking-[0.16em]", isUser ? "text-paper/60" : "text-ink/35")}>
          {message.timeLabel}
        </p>
      </article>
    </div>
  );
}
