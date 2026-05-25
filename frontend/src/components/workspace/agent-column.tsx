// Date: 2026-05-25
// Author: XinYang Li

import { Avatar } from "@/components/ui/avatar";
import { Panel } from "@/components/ui/panel";
import type { Session } from "@/types/domain";

/**
 * Renders the middle agent list column for the active task session.
 * @param props.session The active task session that defines the primary agent.
 * @returns The middle column panel.
 */
export function AgentColumn({ session }: { session: Session }): JSX.Element {
  return (
    <Panel className="flex h-full min-h-0 flex-col p-5">
      <div className="mb-5 flex items-center justify-between">
        <div>
          <p className="text-xs uppercase tracking-[0.22em] text-ink/42">参与协作</p>
          <h2 className="mt-2 font-display text-2xl text-ink">当前主 Agent</h2>
        </div>
        <span className="inline-flex items-center gap-2 rounded-full border border-pine/20 bg-pine/10 px-3 py-1 text-xs text-pine">
          <span className="h-2 w-2 animate-pulseLine rounded-full bg-pine" />
          可用
        </span>
      </div>

      <article className="rounded-[28px] border border-pine/20 bg-white p-4 shadow-panel">
        <div className="flex items-start gap-3">
          <Avatar className="h-12 w-12" name={session.primaryAgentName} />
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <h3 className="text-base font-semibold text-ink">{session.primaryAgentName}</h3>
              <span className="rounded-full bg-mist px-2 py-1 text-[11px] uppercase tracking-[0.14em] text-ink/58">
                {session.sessionKind === "primary" ? "Primary" : "Branch"}
              </span>
            </div>
            <p className="text-sm leading-7 text-ink/60">当前会话：{session.title}</p>
          </div>
        </div>
      </article>
    </Panel>
  );
}
