// Date: 2026-05-25
// Author: XinYang Li

"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { login, register } from "@/lib/api";
import { setStoredToken } from "@/lib/auth";

type Mode = "login" | "register";

/**
 * Renders the interactive authentication form used by login and register pages.
 * @param props.mode Determines whether the component submits login or register requests.
 * @returns The interactive auth form.
 */
export function AuthForm({ mode }: { mode: Mode }): JSX.Element {
  const router = useRouter();
  const [isPending, startTransition] = useTransition();
  const [errorMessage, setErrorMessage] = useState("");
  const [formState, setFormState] = useState({
    username: "",
    email: "",
    password: "",
    confirmPassword: "",
  });

  const isRegister = mode === "register";

  return (
    <form
      className="space-y-4"
      onSubmit={(event) => {
        event.preventDefault();
        setErrorMessage("");

        if (isRegister && formState.password !== formState.confirmPassword) {
          setErrorMessage("两次输入的密码不一致。");
          return;
        }

        startTransition(async () => {
          try {
            const response = isRegister
              ? await register({
                  username: formState.username,
                  email: formState.email,
                  password: formState.password,
                })
              : await login({
                  email: formState.email,
                  password: formState.password,
                });

            setStoredToken(response.token);
            router.push("/workspace");
          } catch (error) {
            setErrorMessage(error instanceof Error ? error.message : "提交失败");
          }
        });
      }}
    >
      {errorMessage ? <div className="rounded-2xl border border-ember/20 bg-ember/10 px-4 py-3 text-sm text-ember">{errorMessage}</div> : null}

      {isRegister ? (
        <div className="space-y-2">
          <label className="text-sm font-medium text-ink/72">用户名</label>
          <Input
            onChange={(event) => setFormState((current) => ({ ...current, username: event.target.value }))}
            placeholder="输入你的名称"
            type="text"
            value={formState.username}
          />
        </div>
      ) : null}

      <div className="space-y-2">
        <label className="text-sm font-medium text-ink/72">邮箱</label>
        <Input
          onChange={(event) => setFormState((current) => ({ ...current, email: event.target.value }))}
          placeholder="xinyang@example.com"
          type="email"
          value={formState.email}
        />
      </div>

      {isRegister ? (
        <div className="grid gap-4 md:grid-cols-2">
          <div className="space-y-2">
            <label className="text-sm font-medium text-ink/72">密码</label>
            <Input
              onChange={(event) => setFormState((current) => ({ ...current, password: event.target.value }))}
              placeholder="请输入密码"
              type="password"
              value={formState.password}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium text-ink/72">确认密码</label>
            <Input
              onChange={(event) => setFormState((current) => ({ ...current, confirmPassword: event.target.value }))}
              placeholder="再次输入密码"
              type="password"
              value={formState.confirmPassword}
            />
          </div>
        </div>
      ) : (
        <div className="space-y-2">
          <label className="text-sm font-medium text-ink/72">密码</label>
          <Input
            onChange={(event) => setFormState((current) => ({ ...current, password: event.target.value }))}
            placeholder="请输入密码"
            type="password"
            value={formState.password}
          />
        </div>
      )}

      <Button className="mt-3 w-full" disabled={isPending}>
        {isPending ? "提交中..." : isRegister ? "注册" : "登录"}
      </Button>
    </form>
  );
}
