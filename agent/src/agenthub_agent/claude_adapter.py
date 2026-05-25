# Date: 2026-05-25
# Author: XinYang Li

"""Claude CLI adapter and runtime helpers for AgentHub v0.2."""

from __future__ import annotations

import shutil
import subprocess
from dataclasses import dataclass
from pathlib import Path


@dataclass(slots=True)
class AgentRuntimeContext:
    """Carries the minimum runtime context needed to execute one agent round."""

    agent_name: str
    session_id: str
    agent_workdir: str
    agent_rule_file: str
    task_title: str
    task_description: str
    session_title: str
    command_mode: str
    user_input: str


def build_first_message_prompt(context: AgentRuntimeContext) -> str:
    """Build the first-round prompt that establishes task and session context.

    Args:
        context: The session runtime metadata and current user input.

    Returns:
        The prompt string used when a Claude session starts for the first time.
    """
    rule_text = read_rule_file(context.agent_rule_file)

    return (
        f"你是 AgentHub 中的 {context.agent_name}。\n"
        "你正在一个持久 session 中工作，请围绕当前 task 与当前 session 持续协作。\n\n"
        f"Task Title: {context.task_title}\n"
        f"Task Description: {context.task_description}\n"
        f"Session Title: {context.session_title}\n\n"
        "Agent Rules:\n"
        f"{rule_text}\n\n"
        "User Input:\n"
        f"{context.user_input}\n"
    )


def read_rule_file(rule_file: str) -> str:
    """Read the current Agent.md file for the active runtime target.

    Args:
        rule_file: The absolute path of the Agent.md file.

    Returns:
        The rule file content, or a fallback sentence when the file is missing.
    """
    path = Path(rule_file)
    if not path.exists():
        return "No agent-specific rules were found."
    return path.read_text(encoding="utf-8").strip() or "No agent-specific rules were found."


def ensure_claude_cli_available(executable: str = "claude") -> None:
    """Validate that the Claude CLI is installed before invocation.

    Args:
        executable: The CLI executable name or absolute path.

    Raises:
        FileNotFoundError: Raised when the executable cannot be resolved.
    """
    if shutil.which(executable) is None:
        raise FileNotFoundError(f"Claude CLI executable not found: {executable}")


def run_agent_command(context: AgentRuntimeContext, executable: str = "claude", timeout_seconds: int = 90) -> str:
    """Execute one agent round against the Claude CLI.

    Args:
        context: The session runtime context for this round.
        executable: The CLI executable name or absolute path.
        timeout_seconds: The maximum execution time before subprocess timeout.

    Returns:
        The stripped CLI stdout content.

    Raises:
        RuntimeError: Raised when the CLI exits with an error or returns empty stdout.
    """
    ensure_claude_cli_available(executable)

    if context.command_mode == "start":
        prompt = build_first_message_prompt(context)
        command = [executable, "--session-id", context.session_id, "-p", prompt]
    else:
        command = [executable, "-p", "--resume", context.session_id, context.user_input]

    result = subprocess.run(
        command,
        capture_output=True,
        text=True,
        timeout=timeout_seconds,
        check=False,
        cwd=context.agent_workdir,
    )

    if result.returncode != 0:
        raise RuntimeError(result.stderr.strip() or "Claude CLI execution failed")

    output = result.stdout.strip()
    if not output:
        raise RuntimeError("Claude CLI returned empty output")

    return output
