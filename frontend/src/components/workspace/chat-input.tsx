// Date: 2026-05-25
// Author: XinYang Li

"use client";

import { ArrowUp } from "lucide-react";
import { useEffect, useRef, useState } from "react";

import { Button } from "@/components/ui/button";
import { MarkdownContent } from "@/components/workspace/markdown-content";
import { Textarea } from "@/components/ui/textarea";

/**
 * Renders the fixed chat composer for the task workspace.
 * @param props.placeholder The prompt hint shown inside the textarea.
 * @param props.isSending Whether the current request is already in flight.
 * @param props.operationHint Optional active quote or reply context shown above the composer.
 * @param props.onClearOperation Clears the active quote or reply context.
 * @param props.onSend The callback executed after the user submits a new message.
 * @returns The interactive chat input block.
 */
export function ChatInput({
  placeholder,
  isSending,
  operationHint,
  onClearOperation,
  onSend,
}: {
  placeholder: string;
  isSending: boolean;
  operationHint?: {
    mode: "quote" | "reply";
    sourceContents: string[];
  } | null;
  onClearOperation?: () => void;
  onSend: (value: string) => void;
}): JSX.Element {
  const [value, setValue] = useState("");
  const [isComposing, setIsComposing] = useState(false);
  const [pendingEnter, setPendingEnter] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement | null>(null);

  useEffect(() => {
    const textarea = textareaRef.current;
    if (!textarea) {
      return;
    }

    textarea.style.height = "0px";
    textarea.style.height = `${Math.min(textarea.scrollHeight, 192)}px`;
  }, [value]);

  useEffect(() => {
    if (!pendingEnter) {
      return;
    }

    /**
     * Clears the pending double-enter window when the second Enter key does not arrive in time.
     * @param none The timer does not accept external parameters.
     */
    const timer = window.setTimeout(() => {
      setPendingEnter(false);
    }, 300);

    return () => window.clearTimeout(timer);
  }, [pendingEnter]);

  return (
    <form
      className="rounded-[30px] border border-line bg-white p-4 shadow-panel"
      onSubmit={(event) => {
        event.preventDefault();

        const trimmed = value.trim();
        if (!trimmed || isSending) {
          return;
        }

        onSend(trimmed);
        setValue("");
      }}
    >
      {operationHint ? (
        <div className="mb-4 rounded-[22px] border border-line bg-mist/70 px-4 py-3">
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <p className="text-xs uppercase tracking-[0.16em] text-pine">
                {operationHint.mode === "quote" ? `引用消息 ${operationHint.sourceContents.length}` : "回复当前消息"}
              </p>
              <div className="mt-2 space-y-2">
                {operationHint.sourceContents.map((sourceContent, index) => (
                  <div className="rounded-2xl border border-line/70 bg-white/70 px-3 py-2" key={`${sourceContent}-${index}`}>
                    <p className="mb-1 text-[11px] uppercase tracking-[0.14em] text-ink/42">
                      {operationHint.mode === "quote" ? `引用消息 ${index + 1}` : "回复消息"}
                    </p>
                    <MarkdownContent
                      content={sourceContent}
                      inlineCodeClassName="rounded-md bg-ink/6 px-1.5 py-0.5 font-mono text-[13px] text-ink"
                      paragraphClassName="whitespace-pre-wrap text-sm leading-6 text-ink/62"
                      tableClassName="overflow-x-auto rounded-2xl border border-line/70 bg-white/75"
                    />
                  </div>
                ))}
              </div>
            </div>
            {onClearOperation ? (
              <Button onClick={onClearOperation} size="sm" type="button" variant="ghost">
                取消
              </Button>
            ) : null}
          </div>
        </div>
      ) : null}
      <Textarea
        className="h-11 min-h-0 max-h-48 resize-none overflow-y-auto border-none p-0 leading-7 focus:ring-0"
        onChange={(event) => setValue(event.target.value)}
        onCompositionEnd={() => setIsComposing(false)}
        onCompositionStart={() => setIsComposing(true)}
        onKeyDown={(event) => {
          if (event.key !== "Enter" || event.shiftKey || event.ctrlKey || event.metaKey) {
            return;
          }

          if (isComposing || event.nativeEvent.isComposing) {
            return;
          }

          if (!pendingEnter) {
            setPendingEnter(true);
            return;
          }

          event.preventDefault();
          setPendingEnter(false);

          const trimmed = value.trim();
          if (!trimmed || isSending) {
            return;
          }

          onSend(trimmed);
          setValue("");
        }}
        placeholder={placeholder}
        ref={textareaRef}
        rows={1}
        value={value}
      />
      <div className="mt-4 flex items-center justify-between">
        <div />
        <Button className="h-12 w-12 rounded-full p-0" disabled={isSending || value.trim().length === 0} size="icon" type="submit">
          <ArrowUp className="h-4 w-4" />
        </Button>
      </div>
    </form>
  );
}
