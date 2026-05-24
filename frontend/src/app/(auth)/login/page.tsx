// Date: 2026-05-25
// Author: XinYang Li

import { AuthForm } from "@/components/auth/auth-form";
import { AuthShell } from "@/components/auth/auth-shell";

/**
 * Renders the login page described in the v0.1 document.
 * @returns The login page UI.
 */
export default function LoginPage(): JSX.Element {
  return (
    <AuthShell
      footerCta="去注册"
      footerHref="/register"
      footerLabel="还没有账号？"
      subtitle="进入你的协作工作区，继续围绕 task 与主 Agent 银河推进任务。"
      title="登录 AgentHub"
    >
      <AuthForm mode="login" />
    </AuthShell>
  );
}
