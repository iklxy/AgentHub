# Date: 2026-05-25
# Author: XinYang Li

"""Claude CLI adapter and prompt assembly helpers for AgentHub v0.1."""

from __future__ import annotations

import shutil
import subprocess
from dataclasses import dataclass


@dataclass(slots=True)
class PromptContext:
    """Carries the minimum context needed to build a Claude prompt."""

    task_title: str
    task_description: str
    message_history: list[dict[str, str]]
    user_input: str


def build_prompt(context: PromptContext) -> str:
    """Assemble the v0.1 prompt string.

    Args:
        context: The task metadata, recent message history, and current user input.

    Returns:
        A prompt string that keeps the task and transcript context together.
    """
    history_lines = []

    for item in context.message_history[-8:]:
        role = item.get("role", "unknown")
        content = item.get("content", "")
        history_lines.append(f"{role}: {content}")

    history_text = "\n".join(history_lines) if history_lines else "暂无历史消息"

    return (
        "你是 AgentHub 中的主 Agent 银河。\n"
        "你需要围绕当前 task 上下文持续回答用户，并保持表达清晰、直接、可继续追问。\n\n"
        f"Task Title: {context.task_title}\n"
        f"Task Description: {context.task_description}\n\n"
        f"History:\n{history_text}\n\n"
        f"User Input:\n{context.user_input}\n"
    )


def ensure_claude_cli_available(executable: str = "claude") -> None:
    """Validate that the Claude CLI is installed before invocation.

    Args:
        executable: The CLI executable name or absolute path.

    Raises:
        FileNotFoundError: Raised when the executable cannot be resolved.
    """
    if shutil.which(executable) is None:
        raise FileNotFoundError(f"Claude CLI executable not found: {executable}")


def run_claude_prompt(prompt: str, executable: str = "claude", timeout_seconds: int = 60) -> str:
    """Execute a prompt against the Claude CLI.

    Args:
        prompt: The fully assembled prompt string passed to Claude.
        executable: The CLI executable name or absolute path.
        timeout_seconds: The maximum execution time before subprocess timeout.

    Returns:
        The stripped CLI stdout content.

    Raises:
        RuntimeError: Raised when the CLI exits with an error or returns empty stdout.
    """
    ensure_claude_cli_available(executable)

    result = subprocess.run(
        [executable, "-p", prompt],
        capture_output=True,
        text=True,
        timeout=timeout_seconds,
        check=False,
    )

    if result.returncode != 0:
        raise RuntimeError(result.stderr.strip() or "Claude CLI execution failed")

    output = result.stdout.strip()

    if not output:
        raise RuntimeError("Claude CLI returned empty output")

    return output
