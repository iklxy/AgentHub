# Date: 2026-05-25
# Author: XinYang Li

"""Minimal CLI entrypoint for the AgentHub agent layer."""

from __future__ import annotations

import argparse
import json
from typing import Any

from agenthub_agent.claude_adapter import PromptContext, build_prompt, run_claude_prompt
from agenthub_agent.logging import get_logger


def parse_args() -> argparse.Namespace:
    """Parse local development arguments.

    Args:
        None: Arguments are read from the command line.

    Returns:
        The parsed namespace containing task metadata and current user input.
    """
    parser = argparse.ArgumentParser(description="AgentHub v0.1 Claude CLI adapter")
    parser.add_argument("--task-title", required=True)
    parser.add_argument("--task-description", required=True)
    parser.add_argument("--user-input", required=True)
    parser.add_argument("--history-json", default="[]")
    return parser.parse_args()


def load_history(raw_history: str) -> list[dict[str, str]]:
    """Load message history from a JSON string.

    Args:
        raw_history: The JSON string passed on the command line.

    Returns:
        A validated list of history items with role and content fields.
    """
    parsed: Any = json.loads(raw_history)

    if not isinstance(parsed, list):
        raise ValueError("history-json must be a JSON array")

    return [item for item in parsed if isinstance(item, dict)]


def main() -> int:
    """Run the local agent entrypoint.

    Args:
        None: Runtime state is loaded from process arguments.

    Returns:
        Exit code 0 on success, or 1 when the CLI call fails.
    """
    logger = get_logger()
    args = parse_args()
    history = load_history(args.history_json)

    context = PromptContext(
      task_title=args.task_title,
      task_description=args.task_description,
      message_history=history,
      user_input=args.user_input,
    )

    logger.info("agent request received", extra={"context": {"task_title": args.task_title}})

    try:
        prompt = build_prompt(context)
        output = run_claude_prompt(prompt=prompt)
    except Exception as exc:  # noqa: BLE001
        logger.error("agent request failed", extra={"context": {"error": str(exc)}})
        return 1

    print(output)
    logger.info("agent request finished")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
