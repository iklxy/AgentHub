# Date: 2026-05-25
# Author: XinYang Li

"""Minimal CLI entrypoint for the AgentHub session-aware agent layer."""

from __future__ import annotations

import argparse
import asyncio
import json
import sys

from agenthub_agent.claude_adapter import AgentRuntimeContext, run_agent_sdk_with_permissions
from agenthub_agent.logging import get_logger


def parse_args() -> argparse.Namespace:
    """Parse local development arguments.

    Args:
        None: Arguments are read from the command line.

    Returns:
        The parsed namespace containing session runtime metadata and current user input.
    """
    parser = argparse.ArgumentParser(description="AgentHub v0.2 Claude Agent SDK adapter")
    parser.add_argument("--agent-name", required=True)
    parser.add_argument("--session-id", required=True)
    parser.add_argument("--agent-workdir", required=True)
    parser.add_argument("--agent-rule-file", required=True)
    parser.add_argument("--task-title", required=True)
    parser.add_argument("--task-description", required=True)
    parser.add_argument("--session-title", required=True)
    parser.add_argument("--resume-session-id", default="")
    parser.add_argument("--user-input", required=True)
    return parser.parse_args()


def write_result(result: str, session_id: str) -> None:
    """Write the final agent result as a JSONL message to stdout.

    Args:
        result: The assistant response text.
        session_id: The SDK-generated session ID, or empty string.
    """
    msg = {"type": "agent_result", "result": result, "session_id": session_id}
    line = json.dumps(msg, ensure_ascii=False)
    sys.stdout.write(line + "\n")
    sys.stdout.flush()


async def main_async() -> int:
    """Run the local agent entrypoint via the Claude Agent SDK with permission approval.

    Permission requests are written as JSONL to stdout and responses are read
    from stdin. The final result is also written as JSONL to stdout.

    Returns:
        Exit code 0 on success, or 1 when the SDK call fails.
    """
    logger = get_logger()
    args = parse_args()

    context = AgentRuntimeContext(
        agent_name=args.agent_name,
        session_id=args.session_id,
        agent_workdir=args.agent_workdir,
        agent_rule_file=args.agent_rule_file,
        task_title=args.task_title,
        task_description=args.task_description,
        session_title=args.session_title,
        user_input=args.user_input,
    )

    resume_session_id = args.resume_session_id.strip() or None

    logger.info(
        "agent sdk request received",
        extra={
            "context": {
                "agent_name": args.agent_name,
                "session_id": args.session_id,
                "resume_session_id": resume_session_id,
            }
        },
    )

    try:
        result, captured_session_id = await run_agent_sdk_with_permissions(
            context, resume_session_id
        )
    except Exception as exc:  # noqa: BLE001
        logger.error(
            "agent sdk request failed",
            extra={
                "context": {
                    "agent_name": args.agent_name,
                    "session_id": args.session_id,
                    "resume_session_id": resume_session_id,
                    "error": str(exc),
                }
            },
        )
        return 1

    write_result(result, captured_session_id)
    logger.info(
        "agent sdk request finished",
        extra={
            "context": {
                "agent_name": args.agent_name,
                "session_id": args.session_id,
                "resume_session_id": resume_session_id,
            }
        },
    )
    return 0


def main() -> int:
    """Entrypoint wrapper that runs the async main on the event loop.

    Returns:
        Exit code 0 on success, 1 on failure.
    """
    return asyncio.run(main_async())


if __name__ == "__main__":
    raise SystemExit(main())
