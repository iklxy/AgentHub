// Date: 2026-05-25
// Author: XinYang Li

import { Plus, Settings2 } from "lucide-react";

import { Avatar } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Panel } from "@/components/ui/panel";
import { TaskListItem } from "@/components/workspace/task-list-item";
import type { Task, Workspace, User } from "@/types/domain";

/**
 * Renders the shared left sidebar used by the workspace and task pages.
 * @param props.workspace The workspace metadata displayed at the top of the sidebar.
 * @param props.user The current user shown in the footer area.
 * @param props.tasks The available tasks rendered in the navigation list.
 * @param props.activeTaskId Optional active task identifier used for highlighting.
 * @param props.onCreateTask Optional callback that opens the task creation UI.
 * @param props.onEditWorkspace Optional callback that opens the workspace edit UI.
 * @param props.onOpenSettings Optional callback that opens the user settings UI.
 * @returns The sidebar panel.
 */
export function WorkspaceSidebar({
  workspace,
  user,
  tasks,
  activeTaskId,
  onCreateTask,
  onEditWorkspace,
  onOpenSettings,
}: {
  workspace: Workspace;
  user: User;
  tasks: Task[];
  activeTaskId?: string;
  onCreateTask?: () => void;
  onEditWorkspace?: () => void;
  onOpenSettings?: () => void;
}): JSX.Element {
  return (
    <Panel className="flex h-full min-h-[calc(100vh-3rem)] flex-col p-5">
      <div className="space-y-4 border-b border-line pb-5">
        <p className="text-xs uppercase tracking-[0.28em] text-pine/65">Workspace</p>
        <div className="space-y-2">
          <h2 className="font-display text-3xl text-ink">{workspace.name}</h2>
          <p className="text-sm leading-7 text-ink/65">{workspace.description}</p>
        </div>
        {onEditWorkspace ? (
          <Button className="w-full justify-center" onClick={onEditWorkspace} type="button" variant="secondary">
            编辑工作区
          </Button>
        ) : null}
      </div>

      <div className="mt-5 flex items-center justify-between">
        <div>
          <p className="text-xs uppercase tracking-[0.22em] text-ink/42">Tasks</p>
          <p className="mt-1 text-sm text-ink/56">{tasks.length} 个任务</p>
        </div>
        {onCreateTask ? (
          <Button aria-label="新建任务" onClick={onCreateTask} size="icon" type="button">
            <Plus className="h-4 w-4" />
          </Button>
        ) : null}
      </div>

      <div className="mt-5 flex-1 space-y-3 overflow-y-auto pr-1">
        {tasks.map((task) => (
          <TaskListItem
            href={`/workspace/tasks/${task.id}`}
            isActive={task.id === activeTaskId}
            key={task.id}
            task={task}
          />
        ))}
      </div>

      <div className="mt-5 flex items-center gap-3 border-t border-line pt-5">
        <Avatar name={user.username} />
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-semibold text-ink">{user.username}</p>
          <p className="truncate text-xs text-ink/52">{user.email}</p>
        </div>
        {onOpenSettings ? (
          <Button aria-label="打开用户设置" onClick={onOpenSettings} size="icon" type="button" variant="ghost">
            <Settings2 className="h-4 w-4" />
          </Button>
        ) : null}
      </div>
    </Panel>
  );
}
