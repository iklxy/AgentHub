<!-- Date: 2026-05-25 -->
<!-- Author: XinYang Li -->

# AgentHub

AgentHub v0.1 is organized as a three-layer workspace:

- `frontend/`: Next.js 15 + TypeScript UI
- `backend/`: Go API service
- `agent/`: Python Claude CLI adapter and workflow layer

## Directory Boundaries

- The front-end only handles page rendering, local UI state, and API calls.
- The back-end owns auth, workspace, task, conversation, and message APIs.
- The agent layer owns prompt assembly, Claude CLI invocation, and agent-side logs.

## Logging

- Front-end uses browser-safe console wrappers for UI diagnostics only.
- Back-end writes structured JSON logs through the shared logger package.
- Agent writes structured JSON logs through a lightweight Python formatter.

## v0.1 Startup Order

1. Start the Go API service in `backend/`
2. Start the Python agent entrypoint in `agent/` when CLI integration is needed
3. Start the Next.js front-end in `frontend/`
