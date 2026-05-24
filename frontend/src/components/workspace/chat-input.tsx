// Date: 2026-05-25
// Author: XinYang Li

"use client";

import { ArrowUp } from "lucide-react";
import { useState } from "react";

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
        className="min-h-28 resize-none border-none p-0 focus:ring-0"
        onChange={(event) => setValue(event.target.value)}
        onKeyDown={(event) => {
          if (event.key === "Enter" && !event.shiftKey) {
            event.preventDefault();
            const trimmed = value.trim();
            if (!trimmed || isSending) {
              return;
            }
            onSend(trimmed);
            setValue("");
          }
        }}
        placeholder={placeholder}
        value={value}
      />
      <div className="mt-4 flex items-center justify-between">
        <p className="text-xs uppercase tracking-[0.16em] text-ink/36">Enter 发送 · Shift + Enter 换行</p>
        <Button className="h-12 w-12 rounded-full p-0" disabled={isSending || value.trim().length === 0} size="icon" type="submit">
          <ArrowUp className="h-4 w-4" />
        </Button>
      </div>
    </form>
  );
}
