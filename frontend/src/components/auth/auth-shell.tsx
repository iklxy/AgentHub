// Date: 2026-05-25
// Author: XinYang Li

import Link from "next/link";
import type { ReactNode } from "react";

/**
 * Renders the shared authentication page frame.
 * @param props.title The main heading shown inside the form card.
 * @param props.subtitle The short helper copy below the heading.
 * @param props.children The form body content for the auth page.
 * @param props.footerLabel The static footer label before the link.
 * @param props.footerHref The destination used for the footer link.
 * @param props.footerCta The interactive footer link label.
 * @returns The composed authentication layout.
 */
export function AuthShell({
  title,
  subtitle,
  children,
  footerLabel,
  footerHref,
  footerCta,
}: {
  title: string;
  subtitle: string;
  children: ReactNode;
  footerLabel: string;
  footerHref: string;
  footerCta: string;
}): JSX.Element {
  return (
    <main className="relative min-h-screen overflow-hidden bg-mist px-6 py-10 text-ink">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top_left,_rgba(29,87,69,0.12),_transparent_35%),radial-gradient(circle_at_bottom_right,_rgba(184,93,61,0.10),_transparent_32%)]" />
      <div className="relative mx-auto grid min-h-[calc(100vh-5rem)] max-w-6xl items-center gap-8 lg:grid-cols-[1.1fr_0.9fr]">
        <section className="animate-rise space-y-6">
          <p className="text-xs uppercase tracking-[0.4em] text-pine/70">AgentHub</p>
          <div className="space-y-4">
            <h1 className="max-w-xl font-display text-5xl leading-tight md:text-6xl">多 Agent 协作，不从空白页开始。</h1>
            <p className="max-w-lg text-base leading-8 text-ink/70">
              v0.1 先聚焦任务工作台，让用户围绕一个 task 与主 Agent 持续协作，而不是进入一个泛化聊天框。
            </p>
          </div>
          <div className="grid max-w-xl gap-3 text-sm text-ink/68">
            <div className="rounded-2xl border border-white/70 bg-white/70 px-4 py-3 backdrop-blur-sm">任务是上下文单位，不是消息列表的附属品。</div>
            <div className="rounded-2xl border border-white/70 bg-white/70 px-4 py-3 backdrop-blur-sm">主 Agent 银河是第一位参与者，后续结构再向多 Agent 扩展。</div>
          </div>
        </section>

        <section className="animate-rise rounded-[32px] border border-white/80 bg-white/85 p-8 shadow-panel backdrop-blur-sm [animation-delay:120ms]">
          <div className="mb-8 space-y-2">
            <h2 className="font-display text-3xl text-ink">{title}</h2>
            <p className="text-sm leading-7 text-ink/65">{subtitle}</p>
          </div>
          {children}
          <div className="mt-6 text-sm text-ink/55">
            {footerLabel}{" "}
            <Link className="font-semibold text-pine transition hover:text-[#164939]" href={footerHref}>
              {footerCta}
            </Link>
          </div>
        </section>
      </div>
    </main>
  );
}
