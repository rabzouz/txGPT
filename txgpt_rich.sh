#!/usr/bin/env bash
set -euo pipefail

LANGUAGE="${TXGPT_LANG:-fr}"
TOOL=""

while [ "$#" -gt 0 ]; do
    case "$1" in
        --lang)
            LANGUAGE="${2:-}"
            shift 2
            ;;
        --tool)
            TOOL="${2:-}"
            shift 2
            ;;
        --help|-h)
            echo "Usage: ./txgpt_rich.sh [--lang fr] [--tool nmap] \"your prompt\""
            exit 0
            ;;
        *)
            break
            ;;
    esac
done

if [ "$#" -eq 0 ]; then
    echo "Usage: ./txgpt_rich.sh [--lang fr] [--tool nmap] \"your prompt\""
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TXGPT_BIN="${TXGPT_BIN:-$SCRIPT_DIR/txgpt}"
PYTHON_BIN="${PYTHON:-python3}"
PROMPT="$*"

if ! command -v "$PYTHON_BIN" >/dev/null 2>&1; then
    PYTHON_BIN="python"
fi

TXGPT_ARGS=(--json --lang "$LANGUAGE")
if [ -n "$TOOL" ]; then
    TXGPT_ARGS+=(--tool "$TOOL")
fi
TXGPT_ARGS+=("$PROMPT")

TXGPT_OUTPUT="$("$TXGPT_BIN" "${TXGPT_ARGS[@]}")"
"$PYTHON_BIN" "$SCRIPT_DIR/rich_display.py" --prompt "$PROMPT" --output "$TXGPT_OUTPUT" --title "txGPT Rich"
