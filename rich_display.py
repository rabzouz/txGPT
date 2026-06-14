import argparse
import html
import json
from typing import Any, List, Optional, Tuple

from rich.console import Console
from rich.markdown import Markdown
from rich.panel import Panel
from rich.table import Table


def parse_output(raw_output: Optional[str]) -> Tuple[str, List[Any]]:
    if not raw_output:
        return "No response received.", []

    try:
        payload = json.loads(raw_output)
    except json.JSONDecodeError:
        return raw_output, []

    if not isinstance(payload, dict):
        return str(payload), []

    response = payload.get("response", "No response received.")
    rows = payload.get("data", [])
    return str(response), rows if isinstance(rows, list) else []


def print_rows(console: Console, rows: List[Any]) -> None:
    if not rows:
        return

    table = Table(title="Extracted Data")

    first_row = rows[0]
    if isinstance(first_row, dict):
        headers = list(first_row.keys())
        for header in headers:
            table.add_column(str(header), style="cyan")
        for row in rows:
            if isinstance(row, dict):
                table.add_row(*(str(row.get(header, "")) for header in headers))
        console.print(table)
        return

    if isinstance(first_row, list):
        for index in range(len(first_row)):
            table.add_column(f"Col{index + 1}", style="cyan" if index % 2 == 0 else "magenta")
        for row in rows:
            if isinstance(row, list):
                table.add_row(*(str(value) for value in row))
        console.print(table)


def main() -> None:
    parser = argparse.ArgumentParser(description="Render txGPT JSON output with Rich.")
    parser.add_argument("--prompt", default="", help="Original prompt")
    parser.add_argument("--output", default="", help="txGPT output as JSON or plain text")
    args = parser.parse_args()

    console = Console()
    response_text, rows = parse_output(args.output)
    response_text = html.unescape(response_text.replace("\\n", "\n").replace("\\r", ""))

    title = args.prompt or "txGPT"
    console.print(Markdown(f"# Results for: {title}"), style="bold blue")
    console.print(Panel(Markdown(response_text), title="txGPT Response", style="green", expand=True))
    print_rows(console, rows)


if __name__ == "__main__":
    main()
