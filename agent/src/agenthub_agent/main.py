# Date: 2026-05-25
# Author: XinYang Li

"""Minimal CLI entrypoint for the AgentHub session-aware agent layer."""

from __future__ import annotations

import argparse

from agenthub_agent.claude_adapter import AgentRuntimeContext, run_agent_command
from agenthub_agent.logging import get_logger


def parse_args() -> argparse.Namespace:
    """Parse local development arguments.

    Args:
        None: Arguments are read from the command line.

    Returns:
        The parsed namespace containing session runtime metadata and current user input.
    """
    parser = argparse.ArgumentParser(description="AgentHub v0.2 Claude CLI adapter")
    parser.add_argument("--agent-name", required=True)
    parser.add_argument("--session-id", required=True)
    parser.add_argument("--agent-workdir", required=True)
    parser.add_argument("--agent-rule-file", required=True)
    parser.add_argument("--task-title", required=True)
    parser.add_argument("--task-description", required=True)
    parser.add_argument("--session-title", required=True)
    parser.add_argument("--command-mode", choices=["start", "resume"], required=True)
    parser.add_argument("--user-input", required=True)
    return parser.parse_args()


def main() -> int:
    """Run the local agent entrypoint.

    Args:
        None: Runtime state is loaded from process arguments.

    Returns:
        Exit code 0 on success, or 1 when the CLI call fails.
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
        command_mode=args.command_mode,
        user_input=args.user_input,
    )

    logger.info(
        "agent request received",
        extra={
            "context": {
                "agent_name": args.agent_name,
                "session_id": args.session_id,
                "command_mode": args.command_mode,
            }
        },
    )

    try:
        output = run_agent_command(context)
    except Exception as exc:  # noqa: BLE001
        logger.error(
            "agent request failed",
            extra={
                "context": {
                    "agent_name": args.agent_name,
                    "session_id": args.session_id,
                    "command_mode": args.command_mode,
                    "error": str(exc),
                }
            },
        )
        return 1

    print(output)
    logger.info(
        "agent request finished",
        extra={
            "context": {
                "agent_name": args.agent_name,
                "session_id": args.session_id,
                "command_mode": args.command_mode,
            }
        },
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
