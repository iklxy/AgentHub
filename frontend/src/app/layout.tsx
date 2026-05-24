// Date: 2026-05-25
// Author: XinYang Li

import type { Metadata } from "next";
import type { ReactNode } from "react";

import "./globals.css";

export const metadata: Metadata = {
  title: "AgentHub",
  description: "AgentHub v0.1 workspace and task collaboration UI.",
};

/**
 * Defines the top-level document shell for the Next.js application.
 * @param props.children The route content rendered inside the body element.
 * @returns The full HTML layout.
 */
export default function RootLayout({ children }: { children: ReactNode }): JSX.Element {
  return (
    <html lang="zh-CN">
      <body>{children}</body>
    </html>
  );
}
