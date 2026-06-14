#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -eq 0 ]; then
    echo "Usage: ./txgpt_rich.sh \"your prompt\""
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TXGPT_BIN="${TXGPT_BIN:-$SCRIPT_DIR/txgpt}"
PYTHON_BIN="${PYTHON:-python3}"
PROMPT="$*"

if ! command -v "$PYTHON_BIN" >/dev/null 2>&1; then
    PYTHON_BIN="python"
fi

TXGPT_OUTPUT="$("$TXGPT_BIN" --json "$PROMPT")"
"$PYTHON_BIN" "$SCRIPT_DIR/rich_display.py" --prompt "$PROMPT" --output "$TXGPT_OUTPUT"
