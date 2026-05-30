# Date: 2026-05-30
# Author: XinYang Li

"""Permission rules for Claude Agent SDK tool approval."""

from __future__ import annotations

from claude_agent_sdk.types import PermissionResultAllow, PermissionResultDeny

# Tools that are always allowed without user interaction
AUTO_ALLOW_TOOLS: set[str] = {"Read", "Glob", "Grep"}

# Bash commands that are always allowed without user interaction
AUTO_ALLOW_COMMANDS: set[str] = {
    "git status",
    "git diff",
    "ls",
    "cat",
    "pwd",
    "find",
    "grep",
}

# Bash commands that are always denied
AUTO_DENY_COMMANDS: set[str] = {
    "rm -rf",
    "git push",
    "sudo",
    "ssh",
    "scp",
    "curl | bash",
}

# File path patterns that trigger auto-deny for Read/Write/Edit
DENY_PATH_PATTERNS: list[str] = [".env", ".ssh"]


def check_auto_allow(tool_name: str, input_data: dict) -> PermissionResultAllow | None:
    """Check whether a tool call can be auto-allowed without user interaction.

    Args:
        tool_name: The name of the tool Claude wants to use.
        input_data: The tool input parameters.

    Returns:
        PermissionResultAllow if the tool can be auto-allowed, None otherwise.
    """
    if tool_name in AUTO_ALLOW_TOOLS:
        return PermissionResultAllow(updated_input=input_data)

    if tool_name == "Bash":
        command = _normalize_command(input_data.get("command", ""))
        for allowed in AUTO_ALLOW_COMMANDS:
            if command.startswith(allowed):
                return PermissionResultAllow(updated_input=input_data)

    return None


def check_auto_deny(tool_name: str, input_data: dict) -> PermissionResultDeny | None:
    """Check whether a tool call should be auto-denied without user interaction.

    Args:
        tool_name: The name of the tool Claude wants to use.
        input_data: The tool input parameters.

    Returns:
        PermissionResultDeny with a reason if the tool should be auto-denied, None otherwise.
    """
    file_path = input_data.get("file_path", "")

    for pattern in DENY_PATH_PATTERNS:
        if pattern in file_path:
            return PermissionResultDeny(
                message=f"Access to files matching '{pattern}' is not allowed"
            )

    if tool_name == "Bash":
        command = _normalize_command(input_data.get("command", ""))
        for denied in AUTO_DENY_COMMANDS:
            if denied in command:
                return PermissionResultDeny(
                    message=f"Command '{denied}' is not allowed"
                )

    return None


def _normalize_command(command: str) -> str:
    """Normalize a bash command for pattern matching.

    Args:
        command: The raw command string from tool input.

    Returns:
        Lowercased, stripped command string.
    """
    return command.strip().lower()
