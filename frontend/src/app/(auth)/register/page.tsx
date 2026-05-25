// Date: 2026-05-25
// Author: XinYang Li

import { AuthForm } from "@/components/auth/auth-form";
import { AuthShell } from "@/components/auth/auth-shell";

/**
 * Renders the registration page described in the v0.1 document.
 * @returns The registration page UI.
 */
export default function RegisterPage(): JSX.Element {
  return (
    <AuthShell
      footerCta="去登录"
      footerHref="/login"
      footerLabel="已经有账号？"
      subtitle=""
      title="注册"
    >
      <AuthForm mode="register" />
    </AuthShell>
  );
}
