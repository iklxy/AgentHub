// Date: 2026-05-25
// Author: XinYang Li

"use client";

import { ArrowUp } from "lucide-react";
import { useEffect, useRef, useState } from "react";

import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";

/**
 * Renders the fixed chat composer for the task workspace.
 * @param props.placeholder The prompt hint shown inside the textarea.
 * @param props.isSending Whether the current request is already in flight.
 * @param props.onSend The callback executed after the user submits a new message.
 * @returns The interactive chat input block.
 */
export function ChatInput({
  placeholder,
  isSending,
  onSend,
}: {
  placeholder: string;
  isSending: boolean;
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
