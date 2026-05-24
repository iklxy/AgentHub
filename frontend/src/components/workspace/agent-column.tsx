// Date: 2026-05-25
// Author: XinYang Li

import { Sparkles } from "lucide-react";

import { Avatar } from "@/components/ui/avatar";
import { Panel } from "@/components/ui/panel";
import type { Conversation } from "@/types/domain";

/**
 * Renders the middle agent list column for the task workspace.
 * @param props.conversation The active conversation that maps to the main agent.
 * @returns The middle column panel.
 */
export function AgentColumn({ conversation }: { conversation: Conversation }): JSX.Element {
  return (
    <Panel className="h-full min-h-[calc(100vh-3rem)] p-5">
      <div className="mb-5 flex items-center justify-between">
        <div>
          <p className="text-xs uppercase tracking-[0.22em] text-ink/42">参与协作</p>
          <h2 className="mt-2 font-display text-2xl text-ink">主 Agent</h2>
        </div>
        <span className="inline-flex items-center gap-2 rounded-full border border-pine/20 bg-pine/10 px-3 py-1 text-xs text-pine">
          <span className="h-2 w-2 rounded-full bg-pine animate-pulseLine" />
          可用
        </span>
      </div>

      <article className="rounded-[28px] border border-pine/20 bg-white p-4 shadow-panel">
        <div className="flex items-start gap-3">
          <Avatar className="h-12 w-12" name={conversation.agentName} />
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <h3 className="text-base font-semibold text-ink">{conversation.agentName}</h3>
              <span className="rounded-full bg-mist px-2 py-1 text-[11px] uppercase tracking-[0.14em] text-ink/58">
                Main
              </span>
            </div>
            <p className="text-sm leading-7 text-ink/62">{conversation.summary}</p>
          </div>
        </div>
        <div className="mt-4 rounded-2xl border border-line bg-mist px-3 py-3 text-xs leading-6 text-ink/60">
          <div className="mb-1 flex items-center gap-2 font-semibold text-ink/70">
            <Sparkles className="h-4 w-4 text-pine" />
            当前能力
          </div>
          围绕当前 task 理解上下文、回答问题，并把后续多 Agent 扩展的结构预留在同一个工作台中。
        </div>
      </article>
    </Panel>
  );
}
