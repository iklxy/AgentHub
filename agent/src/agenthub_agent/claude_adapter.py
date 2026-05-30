# Date: 2026-05-25
# Author: XinYang Li

"""Claude Agent SDK adapter and runtime helpers for AgentHub v0.2."""

from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path

from claude_agent_sdk import AssistantMessage, ClaudeAgentOptions, ResultMessage, SystemMessage, query


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


async def run_agent_sdk(
    context: AgentRuntimeContext,
    resume_session_id: str | None = None,
    timeout_seconds: int = 90,
) -> tuple[str, str]:
    """Execute one agent round via the Claude Agent SDK.

    Args:
        context: The session runtime context for this round.
        resume_session_id: SDK session ID to resume, or None for a new session.
        timeout_seconds: The maximum execution time before subprocess timeout.

    Returns:
        A tuple of (result_text, captured_session_id). The session_id is
        non-empty for the first call and empty for subsequent calls when
        the session already exists.

    Raises:
        RuntimeError: Raised when the SDK returns an error or empty result.
    """
    cwd = context.agent_workdir
    options = ClaudeAgentOptions(
        cwd=cwd,
        resume=resume_session_id,
    )

    prompt: str
    if not resume_session_id:
        prompt = build_first_message_prompt(context)
    else:
        prompt = context.user_input

    captured_session_id = ""
    text_parts: list[str] = []
    last_result = ""

    async for message in query(prompt=prompt, options=options):
        if isinstance(message, SystemMessage):
            if message.subtype == "init":
                session_id = message.data.get("session_id", "")
                if session_id:
                    captured_session_id = str(session_id)
        elif isinstance(message, AssistantMessage):
            for block in message.content:
                text = getattr(block, "text", "")
                if text:
                    text_parts.append(str(text))
        elif isinstance(message, ResultMessage):
            if message.result:
                last_result = message.result
            if message.is_error:
                error_text = "\n".join(message.errors) if message.errors else "SDK execution returned an error"
                raise RuntimeError(error_text)

    result = "".join(text_parts) or last_result
    if not result:
        raise RuntimeError("Claude Agent SDK returned empty output")

    return result, captured_session_id
