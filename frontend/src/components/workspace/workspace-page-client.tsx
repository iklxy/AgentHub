// Date: 2026-05-25
// Author: XinYang Li

"use client";

import { ArrowRight, CheckCircle2, PenSquare, Sparkles } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState, useTransition } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Panel } from "@/components/ui/panel";
import { Textarea } from "@/components/ui/textarea";
import { WorkspaceSidebar } from "@/components/workspace/workspace-sidebar";
import { createTask, getCurrentUser, getTasks, getWorkspace, updateWorkspace } from "@/lib/api";
import { clearStoredToken, getStoredToken } from "@/lib/auth";
import type { Task, User, Workspace } from "@/types/domain";

/**
 * Renders the client-side workspace page backed by real API calls.
 * @returns The interactive workspace page.
 */
export function WorkspacePageClient(): JSX.Element {
  const router = useRouter();
  const [isPending, startTransition] = useTransition();
  const [errorMessage, setErrorMessage] = useState("");
  const [token, setToken] = useState<string | null>(null);
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [workspace, setWorkspaceState] = useState<Workspace | null>(null);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [workspaceDraft, setWorkspaceDraft] = useState({ name: "", description: "" });
  const [taskDraft, setTaskDraft] = useState({ title: "", description: "" });

  useEffect(() => {
    const storedToken = getStoredToken();
    if (!storedToken) {
      router.replace("/login");
      return;
    }

    setToken(storedToken);

    startTransition(async () => {
      try {
        const [user, loadedWorkspace, loadedTasks] = await Promise.all([
          getCurrentUser(storedToken),
          getWorkspace(storedToken),
          getTasks(storedToken),
        ]);

        setCurrentUser(user);
        setWorkspaceState(loadedWorkspace);
        setTasks(loadedTasks);
        setWorkspaceDraft({
          name: loadedWorkspace.name,
          description: loadedWorkspace.description,
        });
      } catch (error) {
        clearStoredToken();
        router.replace("/login");
      }
    });
  }, [router]);

  if (!token || !currentUser || !workspace) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-mist text-ink">
        <div className="rounded-[28px] border border-line bg-paper px-6 py-5 shadow-panel">正在加载工作区...</div>
      </main>
    );
  }

  const firstTask = tasks[0];

  return (
    <main className="min-h-screen bg-mist p-6 text-ink">
      <div className="grid gap-6 xl:grid-cols-[320px_1fr]">
        <WorkspaceSidebar tasks={tasks} user={currentUser} workspace={workspace} />

        <section className="grid gap-6">
          {errorMessage ? <div className="rounded-[24px] border border-ember/20 bg-ember/10 px-5 py-4 text-sm text-ember">{errorMessage}</div> : null}

          <Panel className="overflow-hidden">
            <div className="grid gap-6 p-8 lg:grid-cols-[1.2fr_0.8fr]">
              <div className="space-y-5">
                <p className="text-xs uppercase tracking-[0.28em] text-pine/64">Workspace Hub</p>
                <h1 className="max-w-2xl font-display text-5xl leading-tight text-ink">任务是入口，聊天是推进方式，工作台才是产品本体。</h1>
                <p className="max-w-xl text-sm leading-8 text-ink/62">
                  这里已经接到真实后端接口。工作区信息、任务列表和新建任务动作都会直接写入后端服务与 PostgreSQL。
                </p>
                <div className="flex flex-wrap gap-3">
                  {firstTask ? (
                    <Link
                      className="inline-flex h-11 items-center justify-center gap-2 rounded-full border border-line bg-paper px-5 py-3 text-sm font-semibold text-ink transition duration-200 hover:-translate-y-0.5 hover:border-pine/40 hover:bg-mist"
                      href={`/workspace/tasks/${firstTask.id}`}
                    >
                      进入最近任务
                      <ArrowRight className="h-4 w-4" />
                    </Link>
                  ) : null}
                  <Button
                    onClick={() => {
                      clearStoredToken();
                      router.push("/login");
                    }}
                    variant="secondary"
                  >
                    退出登录
                  </Button>
                </div>
              </div>

              <div className="rounded-[30px] border border-line bg-white p-6">
                <div className="mb-4 flex items-center gap-2 text-sm font-semibold text-pine">
                  <Sparkles className="h-4 w-4" />
                  使用路径
                </div>
                <div className="space-y-4">
                  {[
                    "编辑工作区，让任务边界更清晰",
                    "创建一个新的 task 并自动分配主 Agent 银河",
                    "进入 task 后直接和银河对话",
                  ].map((step) => (
                    <div className="flex items-start gap-3" key={step}>
                      <CheckCircle2 className="mt-0.5 h-5 w-5 text-pine" />
                      <p className="text-sm leading-7 text-ink/65">{step}</p>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </Panel>

          <div className="grid gap-6 lg:grid-cols-2">
            <Panel className="p-6">
              <div className="mb-4 flex items-center gap-2 text-xs uppercase tracking-[0.22em] text-ink/42">
                <PenSquare className="h-4 w-4" />
                编辑工作区
              </div>
              <form
                className="space-y-4"
                onSubmit={(event) => {
                  event.preventDefault();
                  setErrorMessage("");

                  startTransition(async () => {
                    try {
                      const updatedWorkspace = await updateWorkspace(token, workspaceDraft);
                      setWorkspaceState(updatedWorkspace);
                    } catch (error) {
                      setErrorMessage(error instanceof Error ? error.message : "更新工作区失败");
                    }
                  });
                }}
              >
                <Input
                  onChange={(event) => setWorkspaceDraft((current) => ({ ...current, name: event.target.value }))}
                  placeholder="工作区名称"
                  value={workspaceDraft.name}
                />
                <Textarea
                  className="min-h-32"
                  onChange={(event) => setWorkspaceDraft((current) => ({ ...current, description: event.target.value }))}
                  placeholder="工作区描述"
                  value={workspaceDraft.description}
                />
                <Button disabled={isPending} type="submit">
                  {isPending ? "保存中..." : "保存工作区"}
                </Button>
              </form>
            </Panel>

            <Panel className="p-6">
              <p className="text-xs uppercase tracking-[0.22em] text-ink/42">新建任务</p>
              <form
                className="mt-4 space-y-4"
                onSubmit={(event) => {
                  event.preventDefault();
                  setErrorMessage("");

                  startTransition(async () => {
                    try {
                      const createdTask = await createTask(token, taskDraft);
                      setTasks((current) => [createdTask, ...current]);
                      setTaskDraft({ title: "", description: "" });
                      router.push(`/workspace/tasks/${createdTask.id}`);
                    } catch (error) {
                      setErrorMessage(error instanceof Error ? error.message : "新建任务失败");
                    }
                  });
                }}
              >
                <Input
                  onChange={(event) => setTaskDraft((current) => ({ ...current, title: event.target.value }))}
                  placeholder="任务标题"
                  value={taskDraft.title}
                />
                <Textarea
                  className="min-h-32"
                  onChange={(event) => setTaskDraft((current) => ({ ...current, description: event.target.value }))}
                  placeholder="任务描述"
                  value={taskDraft.description}
                />
                <Button disabled={isPending} type="submit">
                  {isPending ? "创建中..." : "创建任务"}
                </Button>
              </form>
            </Panel>
          </div>
        </section>
      </div>
    </main>
  );
}
