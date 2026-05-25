// Date: 2026-05-25
// Author: XinYang Li

"use client";

import { ArrowRight, Clock3, FolderKanban, Sparkles } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState, useTransition } from "react";

import { Button } from "@/components/ui/button";
import { Panel } from "@/components/ui/panel";
import { TaskCreateModal } from "@/components/workspace/task-create-modal";
import { UserSettingsPanel } from "@/components/workspace/user-settings-panel";
import { WorkspaceEditModal } from "@/components/workspace/workspace-edit-modal";
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
  const [isWorkspaceModalOpen, setIsWorkspaceModalOpen] = useState(false);
  const [isTaskModalOpen, setIsTaskModalOpen] = useState(false);
  const [isSettingsPanelOpen, setIsSettingsPanelOpen] = useState(false);

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
  const runningTaskCount = tasks.filter((task) => task.status === "running").length;

  return (
    <main className="min-h-screen bg-mist p-6 text-ink">
      <div className="grid gap-6 xl:grid-cols-[320px_1fr]">
        <WorkspaceSidebar
          onCreateTask={() => setIsTaskModalOpen(true)}
          onEditWorkspace={() => setIsWorkspaceModalOpen(true)}
          onOpenSettings={() => setIsSettingsPanelOpen(true)}
          tasks={tasks}
          user={currentUser}
          workspace={workspace}
        />

        <section className="grid">
          {errorMessage ? <div className="rounded-[24px] border border-ember/20 bg-ember/10 px-5 py-4 text-sm text-ember">{errorMessage}</div> : null}

          <Panel className="h-full min-h-[calc(100vh-3rem)] p-8">
            <div className="grid h-full gap-8 lg:grid-cols-[1.1fr_0.9fr]">
              <div className="space-y-5">
                <h1 className="max-w-2xl font-display text-5xl leading-tight text-ink">让任务进入工作台，再让协作自然沉淀成记录与结果物。</h1>
                {firstTask ? (
                  <Link
                    className="inline-flex h-11 items-center justify-center gap-2 rounded-full border border-line bg-paper px-5 py-3 text-sm font-semibold text-ink transition duration-200 hover:-translate-y-0.5 hover:border-pine/40 hover:bg-mist"
                    href={`/workspace/tasks/${firstTask.id}`}
                  >
                    进入最近任务
                    <ArrowRight className="h-4 w-4" />
                  </Link>
                ) : (
                  <Button onClick={() => setIsTaskModalOpen(true)} type="button">
                    创建第一个任务
                  </Button>
                )}
              </div>

              <div className="grid content-start gap-4 sm:grid-cols-2">
                {[
                  {
                    icon: FolderKanban,
                    label: "任务总数",
                    note: "",
                    value: `${tasks.length}`,
                  },
                  {
                    icon: Sparkles,
                    label: "运行中",
                    note: "",
                    value: `${runningTaskCount}`,
                  },
                  {
                    icon: Clock3,
                    label: "最近更新",
                    note: firstTask ? firstTask.title : "等待新任务进入",
                    value: firstTask?.updatedAtLabel ?? "暂无",
                  },
                  {
                    icon: Sparkles,
                    label: "当前成员",
                    note: "",
                    value: currentUser.username,
                  },
                ].map((item) => (
                  <div className="rounded-[26px] border border-line bg-white p-5" key={item.label}>
                    <div className="flex items-center gap-2 text-sm font-semibold text-pine">
                      <item.icon className="h-4 w-4" />
                      {item.label}
                    </div>
                    <p className="mt-5 font-display text-4xl text-ink">{item.value}</p>
                    <p className="mt-2 text-sm leading-7 text-ink/58">{item.note}</p>
                  </div>
                ))}
              </div>
            </div>
          </Panel>
        </section>
      </div>

      <WorkspaceEditModal
        description={workspaceDraft.description}
        isOpen={isWorkspaceModalOpen}
        isSubmitting={isPending}
        name={workspaceDraft.name}
        onClose={() => setIsWorkspaceModalOpen(false)}
        onDescriptionChange={(value) => setWorkspaceDraft((current) => ({ ...current, description: value }))}
        onNameChange={(value) => setWorkspaceDraft((current) => ({ ...current, name: value }))}
        onSubmit={() => {
          setErrorMessage("");

          startTransition(async () => {
            try {
              const updatedWorkspace = await updateWorkspace(token, workspaceDraft);
              setWorkspaceState(updatedWorkspace);
              setIsWorkspaceModalOpen(false);
            } catch (error) {
              setErrorMessage(error instanceof Error ? error.message : "更新工作区失败");
            }
          });
        }}
      />

      <TaskCreateModal
        description={taskDraft.description}
        isOpen={isTaskModalOpen}
        isSubmitting={isPending}
        onClose={() => setIsTaskModalOpen(false)}
        onDescriptionChange={(value) => setTaskDraft((current) => ({ ...current, description: value }))}
        onSubmit={() => {
          setErrorMessage("");

          startTransition(async () => {
            try {
              const createdTask = await createTask(token, taskDraft);
              setTasks((current) => [createdTask, ...current]);
              setTaskDraft({ title: "", description: "" });
              setIsTaskModalOpen(false);
              router.push(`/workspace/tasks/${createdTask.id}`);
            } catch (error) {
              setErrorMessage(error instanceof Error ? error.message : "新建任务失败");
            }
          });
        }}
        onTitleChange={(value) => setTaskDraft((current) => ({ ...current, title: value }))}
        title={taskDraft.title}
      />

      <UserSettingsPanel
        isOpen={isSettingsPanelOpen}
        onClose={() => setIsSettingsPanelOpen(false)}
        onLogout={() => {
          clearStoredToken();
          router.push("/login");
        }}
        user={currentUser}
      />
    </main>
  );
}
