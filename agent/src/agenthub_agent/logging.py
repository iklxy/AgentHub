# Date: 2026-05-25
# Author: XinYang Li

"""Structured logging helpers for the AgentHub Python agent layer."""

from __future__ import annotations

import json
import logging
from datetime import datetime, timezone
from typing import Any


class JsonFormatter(logging.Formatter):
    """Formats Python log records into newline-safe JSON objects."""

    def format(self, record: logging.LogRecord) -> str:
        """Format one log record.

        Args:
            record: The log record created by the Python logging subsystem.

        Returns:
            A JSON string with stable keys for downstream log consumers.
        """
        payload: dict[str, Any] = {
            "service": "agent",
            "level": record.levelname.lower(),
            "message": record.getMessage(),
            "at": datetime.now(timezone.utc).isoformat(),
        }

        if hasattr(record, "context"):
            payload["context"] = getattr(record, "context")

        return json.dumps(payload, ensure_ascii=False)


def get_logger() -> logging.Logger:
    """Build the shared agent logger.

    Args:
        None: The logger writes JSON lines to standard output.

    Returns:
        A configured Python logger for the agent layer.
    """
    logger = logging.getLogger("agenthub-agent")

    if logger.handlers:
        return logger

    logger.setLevel(logging.INFO)
    handler = logging.StreamHandler()
    handler.setFormatter(JsonFormatter())
    logger.addHandler(handler)
    logger.propagate = False
    return logger
