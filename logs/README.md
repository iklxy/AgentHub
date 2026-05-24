<!-- Date: 2026-05-25 -->
<!-- Author: XinYang Li -->

# Logs

This folder is reserved for local runtime logs if a developer wants to redirect stdout or stderr into files.

Current default behavior:

- `frontend/`: browser console diagnostics only
- `backend/`: structured JSON logs to stdout
- `agent/`: structured JSON logs to stdout

Do not commit generated runtime log files.
