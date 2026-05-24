<!-- Date: 2026-05-25 -->
<!-- Author: XinYang Li -->

# AgentHub v0.1 Architecture Notes

## Folder Isolation

- `frontend/` only contains page rendering, UI components, and front-end utilities.
- `backend/` only contains HTTP APIs, in-memory data storage, and backend logging.
- `agent/` only contains prompt assembly, Claude CLI invocation, and agent logging.

## Logging Strategy

- Front-end logs are diagnostic-only and should not become business state.
- Back-end logs are emitted as JSON to stdout through the shared logger.
- Agent logs are emitted as JSON to stdout so they can be collected separately from backend logs.

## Current Runtime Scope

- Front-end uses mock data for the first UI pass.
- Back-end exposes v0.1 endpoints with in-memory data.
- Agent layer contains the Claude CLI adapter but is not yet wired into backend HTTP handlers.

## Next Integration Step

1. Replace front-end mock data with API calls to the Go backend.
2. Replace the backend message mock reply with a call into the Python agent layer.
3. Persist users, workspaces, tasks, conversations, and messages to PostgreSQL.
