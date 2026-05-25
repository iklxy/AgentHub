// Date: 2026-05-25
// Author: XinYang Li

import { StatusBadge } from "@/components/ui/badge";
import { Panel } from "@/components/ui/panel";
import { ChatInput } from "@/components/workspace/chat-input";
import { MessageBubble } from "@/components/workspace/message-bubble";
import type { Message, Session, Task } from "@/types/domain";

/**
 * Renders the right-side chat workspace for a selected task session.
 * @param props.errorMessage Optional status text shown above the input area.
 * @param props.isSending Whether the current task session is waiting for an assistant reply.
 * @param props.messages The ordered message list rendered inside the transcript.
 * @param props.onSendMessage The callback executed when the user submits a message.
 * @param props.session The active task session displayed in this chat column.
 * @param props.task The active task that owns the current session.
 * @returns The chat column panel.
 */
export function ChatColumn({
  errorMessage,
  isSending,
  messages,
  onSendMessage,
  session,
  task,
}: {
  errorMessage?: string;
  isSending: boolean;
  messages: Message[];
  onSendMessage: (content: string) => void;
  session: Session;
  task: Task;
}): JSX.Element {
  return (
    <Panel className="flex h-full min-h-0 flex-col overflow-hidden">
      <header className="shrink-0 border-b border-line px-7 py-5">
        <div className="flex items-start justify-between gap-4">
          <div className="space-y-2">
            <h2 className="font-display text-3xl text-ink">{session.title}</h2>
            <div className="flex flex-wrap gap-2">
              <span className="rounded-full border border-pine/20 bg-pine/10 px-3 py-1 text-[11px] uppercase tracking-[0.14em] text-pine">
                {session.primaryAgentName}
              </span>
              <span className="rounded-full bg-white px-3 py-1 text-[11px] uppercase tracking-[0.14em] text-ink/44">
                {task.title}
              </span>
            </div>
          </div>
          <StatusBadge status={task.status} />
        </div>
      </header>

      <div className="min-h-0 flex-1 space-y-5 overflow-y-auto px-7 py-6">
        {messages.length === 0 ? (
          <div className="flex h-full min-h-[320px] items-center justify-center">
            <div className="max-w-md space-y-3 text-center">
              <p className="text-xs uppercase tracking-[0.3em] text-ink/40">Ready</p>
              <h2 className="font-display text-3xl text-ink">从第一条消息开始</h2>
              <p className="text-sm leading-7 text-ink/58">这个 session 还没有任何消息，发送后会建立对应的 Claude runtime session。</p>
            </div>
          </div>
        ) : (
          messages.map((message) => <MessageBubble key={message.id} message={message} />)
        )}
      </div>

      <div className="shrink-0 border-t border-line bg-white/65 px-7 py-5 backdrop-blur-sm">
        {errorMessage ? <div className="mb-3 rounded-2xl border border-ember/20 bg-ember/10 px-4 py-3 text-sm text-ember">{errorMessage}</div> : null}
        <ChatInput isSending={isSending} onSend={onSendMessage} placeholder={`继续围绕「${session.title}」发消息，例如：把这个会话拆成更细的执行步骤。`} />
      </div>
    </Panel>
  );
}
