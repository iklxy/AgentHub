// Date: 2026-05-25
// Author: XinYang Li

"use client";

import { Lock, Mail, UserCircle2, X } from "lucide-react";

import { Avatar } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { User } from "@/types/domain";

/**
 * Renders the user settings side panel opened from the sidebar footer gear action.
 * @param props.isOpen Controls whether the side panel is visible.
 * @param props.onClose Closes the side panel.
 * @param props.onLogout Logs the current user out of the workspace.
 * @param props.user The current user snapshot used to populate the fields.
 * @returns The user settings panel or null when closed.
 */
export function UserSettingsPanel({
  isOpen,
  onClose,
  onLogout,
  user,
}: {
  isOpen: boolean;
  onClose: () => void;
  onLogout: () => void;
  user: User;
}): JSX.Element | null {
  if (!isOpen) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-50 flex justify-end bg-ink/28 backdrop-blur-sm">
      <button
        aria-label="关闭用户设置"
        className="absolute inset-0 cursor-default"
        onClick={onClose}
        type="button"
      />
      <aside className="relative z-10 flex h-full w-full max-w-[420px] flex-col border-l border-line bg-paper shadow-panel">
        <header className="flex items-start justify-between gap-4 border-b border-line px-6 py-5">
          <div>
            <p className="text-xs uppercase tracking-[0.22em] text-pine/64">Profile</p>
            <h2 className="mt-2 font-display text-3xl text-ink">用户信息</h2>
          </div>
          <Button aria-label="关闭用户设置" onClick={onClose} size="icon" type="button" variant="ghost">
            <X className="h-5 w-5" />
          </Button>
        </header>

        <div className="flex-1 space-y-6 overflow-y-auto px-6 py-6">
          <div className="flex items-center gap-4 rounded-[28px] border border-line bg-white px-5 py-5">
            <Avatar className="h-14 w-14 text-lg" name={user.username} />
            <div>
              <p className="text-lg font-semibold text-ink">{user.username}</p>
              <p className="text-sm text-ink/56">{user.email}</p>
            </div>
          </div>

          <div className="space-y-4">
            <div className="space-y-2">
              <p className="flex items-center gap-2 text-sm font-semibold text-ink">
                <UserCircle2 className="h-4 w-4 text-pine" />
                用户名称
              </p>
              <Input defaultValue={user.username} placeholder="用户名称" />
            </div>
            <div className="space-y-2">
              <p className="flex items-center gap-2 text-sm font-semibold text-ink">
                <Mail className="h-4 w-4 text-pine" />
                登录邮箱
              </p>
              <Input defaultValue={user.email} placeholder="邮箱地址" type="email" />
            </div>
            <div className="space-y-2">
              <p className="flex items-center gap-2 text-sm font-semibold text-ink">
                <Lock className="h-4 w-4 text-pine" />
                当前状态
              </p>
              <div className="rounded-[22px] border border-line bg-white px-4 py-3 text-sm leading-7 text-ink/62">
                v0.1 仅提供用户信息编辑区 UI，账号资料保存接口将在后续版本接入。
              </div>
            </div>
          </div>
        </div>

        <footer className="border-t border-line px-6 py-5">
          <Button className="w-full justify-center" onClick={onLogout} type="button" variant="secondary">
            退出登录
          </Button>
        </footer>
      </aside>
    </div>
  );
}
