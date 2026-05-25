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
      subtitle="进入AgentHub，开启你的智能体协作之旅。"
      title="登录"
    >
      <AuthForm mode="login" />
    </AuthShell>
  );
}
