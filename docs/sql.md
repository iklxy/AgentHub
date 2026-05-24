<!-- Date: 2026-05-25 -->
<!-- Author: XinYang Li -->

# AgentHub SQL 字段说明

本文档基于当前确认的数据库结构整理，目标是为后续前端、后端、Agent 编排和产物能力开发提供统一字段参考。

## 1. users

用途：

- 全局用户表
- 承载登录、注册、头像、角色、状态等基础信息

字段说明：

| 字段名 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `UUID` | 用户主键 |
| `email` | `VARCHAR(255)` | 登录邮箱，唯一 |
| `name` | `VARCHAR(128)` | 用户名称 |
| `avatar_url` | `TEXT` | 用户头像地址 |
| `role` | `VARCHAR(32)` | 全局用户角色，默认 `member` |
| `status` | `VARCHAR(32)` | 用户状态，默认 `active` |
| `password_hash` | `TEXT` | 登录密码哈希，当前后端已使用 |
| `created_at` | `TIMESTAMPTZ` | 创建时间 |
| `updated_at` | `TIMESTAMPTZ` | 更新时间 |

索引：

- `idx_users_status_updated_at`

## 2. workspaces

用途：

- workspace 主表
- 表示一个协作工作区

字段说明：

| 字段名 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `UUID` | workspace 主键 |
| `name` | `VARCHAR(128)` | workspace 名称 |
| `description` | `TEXT` | workspace 描述 |
| `status` | `VARCHAR(32)` | workspace 状态，默认 `active` |
| `created_by` | `UUID` | 创建人，关联 `users.id` |
| `created_at` | `TIMESTAMPTZ` | 创建时间 |
| `updated_at` | `TIMESTAMPTZ` | 更新时间 |

索引：

- `idx_workspaces_status_updated_at`
- `idx_workspaces_created_by`

## 3. workspace_members

用途：

- workspace 成员关系表
- 表示用户与 workspace 的归属关系

字段说明：

| 字段名 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `UUID` | 成员关系主键 |
| `workspace_id` | `UUID` | 关联 `workspaces.id` |
| `user_id` | `UUID` | 关联 `users.id` |
| `member_role` | `VARCHAR(32)` | 成员角色，默认 `member` |
| `joined_at` | `TIMESTAMPTZ` | 加入时间 |

约束：

- `UNIQUE (workspace_id, user_id)`

索引：

- `idx_workspace_members_workspace_id`
- `idx_workspace_members_user_id`

## 4. agent_connectors

用途：

- 外部 Agent 接入连接器
- 承载外部平台、CLI、Webhook 等接入配置

字段说明：

| 字段名 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `UUID` | connector 主键 |
| `workspace_id` | `UUID` | 所属 workspace |
| `name` | `VARCHAR(128)` | connector 名称 |
| `connector_type` | `VARCHAR(64)` | connector 类型 |
| `base_url` | `TEXT` | 服务基础地址 |
| `auth_type` | `VARCHAR(32)` | 鉴权方式，默认 `none` |
| `config_json` | `JSONB` | connector 运行配置 |
| `capabilities_json` | `JSONB` | connector 能力列表 |
| `enabled` | `BOOLEAN` | 是否启用 |
| `created_at` | `TIMESTAMPTZ` | 创建时间 |
| `updated_at` | `TIMESTAMPTZ` | 更新时间 |

索引：

- `idx_agent_connectors_workspace_type`

## 5. agents

用途：

- Agent wrapper
- 表示系统内部管理的 Agent 实体

字段说明：

| 字段名 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `UUID` | agent 主键 |
| `workspace_id` | `UUID` | 所属 workspace |
| `name` | `VARCHAR(128)` | agent 名称 |
| `kind` | `VARCHAR(32)` | agent 类型 |
| `status` | `VARCHAR(32)` | agent 状态，默认 `active` |
| `description` | `TEXT` | agent 描述 |
| `avatar_url` | `TEXT` | agent 头像 |
| `capability_tags` | `JSONB` | 能力标签 |
| `connector_id` | `UUID` | 关联 `agent_connectors.id` |
| `source_kind` | `VARCHAR(32)` | 来源类型，默认 `external_cli` |
| `provider_type` | `VARCHAR(64)` | provider 类型 |
| `remote_agent_id` | `TEXT` | 外部平台上的 agent id |
| `default_model` | `VARCHAR(128)` | 默认模型 |
| `default_rule_id` | `UUID` | 默认规则 id |
| `system_prompt_md` | `TEXT` | system prompt |
| `tool_schema_json` | `JSONB` | tool schema |
| `created_by` | `UUID` | 创建人 |
| `created_at` | `TIMESTAMPTZ` | 创建时间 |
| `updated_at` | `TIMESTAMPTZ` | 更新时间 |

索引：

- `idx_agents_workspace_kind`
- `idx_agents_workspace_status`

## 6. tasks

用途：

- Task 主表
- 表示一个协作型任务

字段说明：

| 字段名 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `UUID` | task 主键 |
| `workspace_id` | `UUID` | 所属 workspace |
| `title` | `VARCHAR(255)` | task 标题 |
| `description` | `TEXT` | task 描述 |
| `status` | `VARCHAR(32)` | task 状态，默认 `draft` |
| `priority` | `INTEGER` | 优先级，默认 `0` |
| `current_session_id` | `UUID` | 当前会话 id |
| `current_primary_agent_id` | `UUID` | 当前主 agent id |
| `tags` | `JSONB` | task 标签 |
| `created_by` | `UUID` | 创建人 |
| `created_at` | `TIMESTAMPTZ` | 创建时间 |
| `updated_at` | `TIMESTAMPTZ` | 更新时间 |
| `completed_at` | `TIMESTAMPTZ` | 完成时间 |

索引：

- `idx_tasks_workspace_status_updated_at`
- `idx_tasks_workspace_created_by`

## 7. task_sessions

用途：

- Session / Conversation
- 表示一个 task 下的对话会话

字段说明：

| 字段名 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `UUID` | session 主键 |
| `workspace_id` | `UUID` | 所属 workspace |
| `task_id` | `UUID` | 所属 task |
| `title` | `VARCHAR(255)` | 会话标题 |
| `chat_mode` | `VARCHAR(16)` | 聊天模式，默认 `single` |
| `status` | `VARCHAR(32)` | 会话状态，默认 `active` |
| `is_pinned` | `BOOLEAN` | 是否置顶 |
| `archived_at` | `TIMESTAMPTZ` | 归档时间 |
| `last_active_at` | `TIMESTAMPTZ` | 最近活跃时间 |
| `last_message_at` | `TIMESTAMPTZ` | 最近消息时间 |
| `last_message_preview` | `TEXT` | 最近消息预览 |
| `primary_agent_id` | `UUID` | 主 agent id |
| `created_by` | `UUID` | 创建人 |
| `created_at` | `TIMESTAMPTZ` | 创建时间 |
| `updated_at` | `TIMESTAMPTZ` | 更新时间 |

索引：

- `idx_task_sessions_workspace_task_last_active`
- `idx_task_sessions_workspace_pinned`

## 8. session_participants

用途：

- Session participants
- 表示一个 session 里参与协作的 agent 列表

字段说明：

| 字段名 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `UUID` | participant 主键 |
| `workspace_id` | `UUID` | 所属 workspace |
| `session_id` | `UUID` | 所属 session |
| `agent_id` | `UUID` | 参与的 agent |
| `participant_role` | `VARCHAR(32)` | 参与角色 |
| `display_order` | `INTEGER` | 展示顺序 |
| `joined_at` | `TIMESTAMPTZ` | 加入时间 |
| `left_at` | `TIMESTAMPTZ` | 离开时间 |

约束：

- `UNIQUE (session_id, agent_id)`

索引：

- `idx_session_participants_session_id`
- `idx_session_participants_agent_id`

## 9. task_steps

用途：

- Task steps
- 表示一个 task 被拆分后的步骤

字段说明：

| 字段名 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `UUID` | step 主键 |
| `workspace_id` | `UUID` | 所属 workspace |
| `task_id` | `UUID` | 所属 task |
| `parent_step_id` | `UUID` | 父 step id |
| `title` | `VARCHAR(255)` | step 标题 |
| `description` | `TEXT` | step 描述 |
| `order_index` | `INTEGER` | 排序索引 |
| `status` | `VARCHAR(32)` | step 状态，默认 `queued` |
| `planned_by_agent_id` | `UUID` | 规划此 step 的 agent |
| `assigned_agent_id` | `UUID` | 执行此 step 的 agent |
| `input_snapshot` | `JSONB` | 输入快照 |
| `output_snapshot` | `JSONB` | 输出快照 |
| `started_at` | `TIMESTAMPTZ` | 开始时间 |
| `finished_at` | `TIMESTAMPTZ` | 完成时间 |
| `created_at` | `TIMESTAMPTZ` | 创建时间 |
| `updated_at` | `TIMESTAMPTZ` | 更新时间 |

索引：

- `idx_task_steps_task_order`

## 10. task_step_runs

用途：

- Task step runs
- 记录 step 的每次运行

字段说明：

| 字段名 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `UUID` | run 主键 |
| `workspace_id` | `UUID` | 所属 workspace |
| `task_step_id` | `UUID` | 所属 step |
| `agent_id` | `UUID` | 执行 agent |
| `run_no` | `INTEGER` | 第几次运行 |
| `status` | `VARCHAR(32)` | 运行状态 |
| `input_json` | `JSONB` | 输入 |
| `output_json` | `JSONB` | 输出 |
| `error_message` | `TEXT` | 错误信息 |
| `started_at` | `TIMESTAMPTZ` | 开始时间 |
| `finished_at` | `TIMESTAMPTZ` | 结束时间 |
| `usage_json` | `JSONB` | token/调用量等使用数据 |

约束：

- `UNIQUE (task_step_id, run_no)`

索引：

- `idx_task_step_runs_step_id`

## 11. artifacts

用途：

- Artifacts
- 表示任务产物主表

字段说明：

| 字段名 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `UUID` | artifact 主键 |
| `workspace_id` | `UUID` | 所属 workspace |
| `task_id` | `UUID` | 所属 task |
| `type` | `VARCHAR(32)` | 产物类型 |
| `title` | `VARCHAR(255)` | 产物标题 |
| `status` | `VARCHAR(32)` | 产物状态 |
| `current_version_id` | `UUID` | 当前版本 id |
| `preview_schema` | `JSONB` | 预览结构 |
| `export_url` | `TEXT` | 导出地址 |
| `created_by_agent_id` | `UUID` | 生成 agent |
| `created_by_message_id` | `UUID` | 来源消息 id |
| `created_at` | `TIMESTAMPTZ` | 创建时间 |
| `updated_at` | `TIMESTAMPTZ` | 更新时间 |

索引：

- `idx_artifacts_task_type`

## 12. artifact_versions

用途：

- Artifact versions
- 记录产物的多版本内容

字段说明：

| 字段名 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `UUID` | version 主键 |
| `workspace_id` | `UUID` | 所属 workspace |
| `artifact_id` | `UUID` | 所属 artifact |
| `version_no` | `INTEGER` | 版本号 |
| `content_md` | `TEXT` | markdown 内容 |
| `content_json` | `JSONB` | 结构化内容 |
| `content_html` | `TEXT` | html 内容 |
| `diff_base_version_id` | `UUID` | diff 基准版本 |
| `generated_by_run_id` | `UUID` | 生成来源 run |
| `created_by_user_id` | `UUID` | 创建用户 |
| `created_at` | `TIMESTAMPTZ` | 创建时间 |

约束：

- `UNIQUE (artifact_id, version_no)`

索引：

- `idx_artifact_versions_artifact_version`

补充约束：

- `artifacts.current_version_id -> artifact_versions.id`

## 13. deployments

用途：

- Deployments
- 记录部署与预览发布

字段说明：

| 字段名 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `UUID` | deployment 主键 |
| `workspace_id` | `UUID` | 所属 workspace |
| `task_id` | `UUID` | 所属 task |
| `artifact_id` | `UUID` | 所属 artifact |
| `deployment_type` | `VARCHAR(32)` | 部署类型 |
| `status` | `VARCHAR(32)` | 部署状态 |
| `preview_url` | `TEXT` | 预览地址 |
| `deployment_url` | `TEXT` | 正式部署地址 |
| `package_url` | `TEXT` | 打包文件地址 |
| `log_url` | `TEXT` | 部署日志地址 |
| `triggered_by_agent_id` | `UUID` | 触发 agent |
| `triggered_by_user_id` | `UUID` | 触发用户 |
| `created_at` | `TIMESTAMPTZ` | 创建时间 |
| `updated_at` | `TIMESTAMPTZ` | 更新时间 |

索引：

- `idx_deployments_task_status`

## 14. messages

用途：

- Messages
- 表示会话中的消息流

字段说明：

| 字段名 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `UUID` | 消息主键 |
| `workspace_id` | `UUID` | 所属 workspace |
| `task_id` | `UUID` | 所属 task |
| `session_id` | `UUID` | 所属 session |
| `sender_type` | `VARCHAR(32)` | 发送者类型 |
| `sender_id` | `UUID` | 发送者 id |
| `role` | `VARCHAR(32)` | 消息角色 |
| `message_type` | `VARCHAR(32)` | 消息类型，默认 `text` |
| `content_md` | `TEXT` | markdown 内容 |
| `content_json` | `JSONB` | 结构化内容 |
| `reply_to_message_id` | `UUID` | 回复目标消息 |
| `step_id` | `UUID` | 关联 step |
| `artifact_id` | `UUID` | 关联 artifact |
| `mentioned_agent_ids` | `JSONB` | 提及的 agent id 列表 |
| `is_pinned` | `BOOLEAN` | 是否置顶 |
| `pinned_at` | `TIMESTAMPTZ` | 置顶时间 |
| `created_at` | `TIMESTAMPTZ` | 创建时间 |

索引：

- `idx_messages_task_session_created_at`
- `idx_messages_session_created_at`

## 15. session_context_pins

用途：

- Session pins
- 记录 session 长期上下文 pin

字段说明：

| 字段名 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `UUID` | pin 主键 |
| `workspace_id` | `UUID` | 所属 workspace |
| `session_id` | `UUID` | 所属 session |
| `message_id` | `UUID` | 被 pin 的消息 |
| `created_at` | `TIMESTAMPTZ` | 创建时间 |

约束：

- `UNIQUE (session_id, message_id)`

索引：

- `idx_session_context_pins_session_id`

## 16. agent_rules

用途：

- Agent rules
- 表示 agent / task / workspace 范围的规则

字段说明：

| 字段名 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `UUID` | 规则主键 |
| `workspace_id` | `UUID` | 所属 workspace |
| `agent_id` | `UUID` | 关联 agent |
| `task_id` | `UUID` | 关联 task |
| `scope` | `VARCHAR(32)` | 规则作用域 |
| `title` | `VARCHAR(255)` | 规则标题 |
| `content_md` | `TEXT` | 规则 markdown 内容 |
| `parsed_rules_json` | `JSONB` | 解析后的结构化规则 |
| `version` | `INTEGER` | 当前版本号 |
| `status` | `VARCHAR(32)` | 规则状态 |
| `created_by` | `UUID` | 创建人 |
| `updated_by` | `UUID` | 更新人 |
| `created_at` | `TIMESTAMPTZ` | 创建时间 |
| `updated_at` | `TIMESTAMPTZ` | 更新时间 |

索引：

- `idx_agent_rules_scope_agent`
- `idx_agent_rules_scope_task`

## 17. agent_rule_versions

用途：

- Agent rule versions
- 记录规则版本变化

字段说明：

| 字段名 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `UUID` | 规则版本主键 |
| `workspace_id` | `UUID` | 所属 workspace |
| `agent_rule_id` | `UUID` | 所属规则 |
| `version_no` | `INTEGER` | 版本号 |
| `content_md` | `TEXT` | 规则内容 |
| `parsed_rules_json` | `JSONB` | 结构化规则 |
| `change_reason` | `TEXT` | 变更原因 |
| `created_by` | `UUID` | 创建人 |
| `created_at` | `TIMESTAMPTZ` | 创建时间 |

约束：

- `UNIQUE (agent_rule_id, version_no)`

索引：

- `idx_agent_rule_versions_rule_version`

## 18. task_attachments

用途：

- Task attachments
- 记录任务附件和外部来源文件

字段说明：

| 字段名 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `UUID` | 附件主键 |
| `workspace_id` | `UUID` | 所属 workspace |
| `task_id` | `UUID` | 所属 task |
| `file_name` | `VARCHAR(255)` | 文件名称 |
| `file_type` | `VARCHAR(64)` | 文件类型 |
| `storage_key` | `TEXT` | 对象存储 key |
| `source_type` | `VARCHAR(32)` | 来源类型 |
| `source_url` | `TEXT` | 来源地址 |
| `meta_json` | `JSONB` | 扩展元信息 |
| `created_by` | `UUID` | 上传人 |
| `created_at` | `TIMESTAMPTZ` | 创建时间 |

索引：

- `idx_task_attachments_task_created_at`

## 19. audit_events

用途：

- Audit events
- 记录系统审计和关键事件

字段说明：

| 字段名 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `UUID` | 审计事件主键 |
| `workspace_id` | `UUID` | 所属 workspace |
| `task_id` | `UUID` | 关联 task |
| `agent_id` | `UUID` | 关联 agent |
| `event_type` | `VARCHAR(64)` | 事件类型 |
| `payload_json` | `JSONB` | 事件载荷 |
| `created_by` | `UUID` | 触发用户 |
| `created_at` | `TIMESTAMPTZ` | 创建时间 |

索引：

- `idx_audit_events_workspace_type_created_at`

## 当前 v0.1 已直接使用的表

当前代码链路已经直接依赖这些表：

- `users`
- `workspaces`
- `workspace_members`
- `tasks`
- `task_sessions`
- `messages`

当前代码已直接依赖的关键字段：

- `users.name`
- `users.email`
- `users.password_hash`
- `workspaces.created_by`
- `workspace_members.user_id`
- `workspace_members.workspace_id`
- `task_sessions.task_id`
- `messages.session_id`
- `messages.content_md`

## 当前开发注意点

- 这版数据库并不是“一个用户一个 workspace”的模型，而是 `workspace_members` 关系模型
- 对话表实际名称是 `task_sessions`，不是旧版简化结构里的 `conversations`
- 消息正文当前主字段是 `content_md`，不是简化模型里的 `content`
- 如果后续继续补 Agent、Artifact、Rules、Deployment 功能，应优先对齐本文档字段，不要再回退到旧版简化表结构
