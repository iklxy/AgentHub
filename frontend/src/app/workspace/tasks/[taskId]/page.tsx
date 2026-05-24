// Date: 2026-05-25
// Author: XinYang Li

import { TaskPageClient } from "@/components/workspace/task-page-client";

/**
 * Renders the task workspace page for a given task identifier.
 * @param props.params The dynamic route parameters containing the task identifier.
 * @returns The task detail workspace UI.
 */
export default async function TaskDetailPage({
  params,
}: {
  params: Promise<{ taskId: string }>;
}): Promise<JSX.Element> {
  const { taskId } = await params;
  return <TaskPageClient taskId={taskId} />;
}
