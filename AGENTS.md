# AGENTS.md

## Project Goal

This project is AgentHub, a multi-agent collaboration platform whose core interaction paradigm is IM-style chat. Users can have one-on-one or group conversations with agents such as Claude Code, Codex, and OpenCode, and view text, code diffs, web previews, and file artifacts directly inside the chat stream.

## Code Style

- Use TypeScript throughout the project.
- Use functional components for React components.
- Prefer shadcn/ui components and do not recreate basic components unnecessarily.
- Prefer Tailwind CSS for styling.
- Use PascalCase for component names.
- Use camelCase for function and variable names.
- Use PascalCase for type names.
- Do not use `any` unless there is a clear reason. If `any` is used, explain the reason near the code.
- Do not write meaningless comments. Only add concise comments for complex business logic.
- Keep each file focused on a single responsibility and avoid overly large components.
- Prefer splitting code into modules such as `components`, `lib`, `types`, and `services`.
- Every generated function must include clear comments. The comments should describe the accepted parameters and the specific meaning of each parameter.
- Comments must be written in English and should be friendly for agents to understand.
- Each file must include a file header with the date and author. The author is XinYang Li.
- The front-end page uses Chinese.

## Output Format

After each modification, reply using the following format:

1. Which files were modified
2. What functionality was implemented
3. Whether tests or builds were run
4. Whether there are any unfinished items

Do not output long, unrelated explanations.
Do not repeatedly paste complete files unless explicitly requested by the user.
If a modification fails, explain the reason for the failure and suggest the next steps.

## Modification Requirements

- Before making changes, read the relevant files first and understand the existing structure.
- Do not modify shared types, database schema, or environment variables casually.
- If database schema changes are involved, explain the migration impact.
- If UI changes are involved, keep the overall visual style consistent.
- If Agent orchestration logic is involved, consider task status, error handling, and logging.

## SKILL

- For frontend page generation or frontend code modification tasks, prefer using the `frontend-design` skill.

## Python environment

- If any Python packages require installation, please install them under the AgentHub virtual environment managed by conda.