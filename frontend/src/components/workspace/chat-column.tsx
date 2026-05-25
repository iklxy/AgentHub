// Date: 2026-05-25
// Author: XinYang Li

import { ArrowLeft } from "lucide-react";
import Link from "next/link";

import { StatusBadge } from "@/components/ui/badge";
import { Panel } from "@/components/ui/panel";
import { ChatInput } from "@/components/workspace/chat-input";
import { MessageBubble } from "@/components/workspace/message-bubble";
import type { Message, Task } from "@/types/domain";

/**
 * Renders the right-side chat workspace for a selected task.
 * @param props.task The active task that defines the current conversation context.
 * @param props.messages The ordered message list rendered inside the transcript.
 * @param props.isSending Whether the current task is waiting for an assistant reply.
 * @param props.errorMessage Optional status text shown above the input area.
 * @param props.onSendMessage The callback executed when the user submits a message.
 * @returns The chat column panel.
 */
export function ChatColumn({
  task,
  messages,
  isSending,
  errorMessage,
  onSendMessage,
}: {
  task: Task;
  messages: Message[];
  isSending: boolean;
  errorMessage?: string;
  onSendMessage: (content: string) => void;
}): JSX.Element {
  return (
    <Panel className="flex h-full min-h-0 flex-col overflow-hidden">
      <header className="shrink-0 border-b border-line px-7 py-3">
        <div className="mb-4 flex items-center justify-between gap-4">
          <Link className="inline-flex items-center gap-2 text-sm text-ink/56 transition hover:text-pine" href="/workspace">
            <ArrowLeft className="h-4 w-4" />
            返回工作区
          </Link>
          <StatusBadge status={task.status} />
        </div>
      </header>

      <div className="min-h-0 flex-1 space-y-5 overflow-y-auto px-7 py-6">
        {messages.length === 0 ? (
          <div className="flex h-full min-h-[320px] items-center justify-center">
            <div className="max-w-md space-y-3 text-center">
              <p className="text-xs uppercase tracking-[0.3em] text-ink/40">Ready</p>
              <h2 className="font-display text-3xl text-ink">从第一条消息开始</h2>
              <p className="text-sm leading-7 text-ink/58">当前任务已经建好</p>
            </div>
          </div>
        ) : (
          messages.map((message) => <MessageBubble key={message.id} message={message} />)
        )}
      </div>

      <div className="shrink-0 border-t border-line bg-white/65 px-7 py-5 backdrop-blur-sm">
        {errorMessage ? <div className="mb-3 rounded-2xl border border-ember/20 bg-ember/10 px-4 py-3 text-sm text-ember">{errorMessage}</div> : null}
        <ChatInput isSending={isSending} onSend={onSendMessage} placeholder="围绕当前任务继续提问，例如：帮我拆成 3 个学习阶段。" />
      </div>
    </Panel>
  );
}
