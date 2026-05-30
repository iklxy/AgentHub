# Date: 2026-05-25
# Author: XinYang Li

"""Claude Agent SDK adapter and runtime helpers for AgentHub v0.2."""

from __future__ import annotations

import json
import sys
from collections.abc import AsyncIterable
from dataclasses import dataclass
from pathlib import Path

from claude_agent_sdk import AssistantMessage, ClaudeAgentOptions, ResultMessage, SystemMessage, query
from claude_agent_sdk.types import HookMatcher, PermissionResultAllow, PermissionResultDeny, ToolPermissionContext

from agenthub_agent.permissions import check_auto_allow, check_auto_deny


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


async def _prompt_stream(prompt_text: str) -> AsyncIterable[dict]:
    """Yield a single user message as an async iterable for SDK streaming mode."""
    yield {
        "type": "user",
        "message": {
            "role": "user",
            "content": prompt_text,
        },
    }


async def _dummy_hook(input_data: dict, tool_use_id: str, context: object) -> dict:
    """Keep the stream open so can_use_tool callbacks can fire."""
    return {"continue_": True}


async def run_agent_sdk_with_permissions(
    context: AgentRuntimeContext,
    resume_session_id: str | None = None,
    request_seq: int = 0,
) -> tuple[str, str]:
    """Execute one agent round via the Claude Agent SDK with interactive permission approval.

    Writes JSONL messages to stdout for backend consumption and reads JSONL
    responses from stdin. Permission decisions follow a three-tier check:
    1. Auto-allow (Read/Glob/Grep, safe Bash commands) → immediate return
    2. Auto-deny (sensitive paths, dangerous commands) → immediate return
    3. Needs approval → write to stdout, block on stdin response

    Args:
        context: The session runtime context for this round.
        resume_session_id: SDK session ID to resume, or None for a new session.
        request_seq: Starting sequence number for permission request IDs.

    Returns:
        A tuple of (result_text, captured_session_id).
    """
    cwd = context.agent_workdir
    session_ref = resume_session_id or context.session_id
    seq = request_seq

    async def can_use_tool(
        tool_name: str,
        input_data: dict,
        _tool_context: ToolPermissionContext,
    ) -> PermissionResultAllow | PermissionResultDeny:
        nonlocal seq

        auto_allow = check_auto_allow(tool_name, input_data)
        if auto_allow is not None:
            return auto_allow

        auto_deny = check_auto_deny(tool_name, input_data)
        if auto_deny is not None:
            return auto_deny

        seq += 1
        request_id = f"perm_{session_ref}_{seq}"

        request_msg = {
            "type": "permission_request",
            "requestId": request_id,
            "toolName": tool_name,
            "input": input_data,
        }
        _write_jsonl(request_msg)

        response = _read_jsonl()
        if response is None:
            return PermissionResultDeny(message="Backend communication failed, denying by default")

        behavior = response.get("behavior", "deny")
        if behavior == "allow":
            updated_input = response.get("updatedInput", input_data)
            return PermissionResultAllow(updated_input=updated_input)

        message = response.get("message", "User denied this action")
        return PermissionResultDeny(message=message)

    options = ClaudeAgentOptions(
        cwd=cwd,
        resume=resume_session_id,
        can_use_tool=can_use_tool,
        hooks={"PreToolUse": [HookMatcher(matcher=None, hooks=[_dummy_hook])]},
    )

    prompt: str
    if not resume_session_id:
        prompt = build_first_message_prompt(context)
    else:
        prompt = context.user_input

    captured_session_id = ""
    text_parts: list[str] = []
    last_result = ""

    async for message in query(prompt=_prompt_stream(prompt), options=options):
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


def _write_jsonl(data: dict) -> None:
    """Write a JSON line to stdout, flushing immediately for pipe consumers.

    Args:
        data: The dictionary to serialize and write.
    """
    line = json.dumps(data, ensure_ascii=False)
    sys.stdout.write(line + "\n")
    sys.stdout.flush()


def _read_jsonl() -> dict | None:
    """Read a JSON line from stdin.

    Returns:
        The parsed dictionary, or None if stdin is closed or invalid.
    """
    line = sys.stdin.readline()
    if not line:
        return None
    try:
        return json.loads(line.strip())
    except json.JSONDecodeError:
        return None
