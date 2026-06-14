import argparse
import html
import json
from typing import Any, List, Optional, Tuple

try:
    from rich.console import Console
    from rich.markdown import Markdown
    from rich.panel import Panel
    from rich.rule import Rule
    from rich.table import Table
    from rich.text import Text

    HAS_RICH = True
except ImportError:
    HAS_RICH = False


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


def print_rows(console: Any, rows: List[Any]) -> None:
    if not rows:
        return

    table = Table(title="Extracted Data", show_lines=False, header_style="bold cyan")

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
        headers = ["Port", "State", "Service"] if len(first_row) == 3 else []
        for index in range(len(first_row)):
            header = headers[index] if index < len(headers) else f"Col{index + 1}"
            table.add_column(header, style="cyan" if index % 2 == 0 else "magenta")
        for row in rows:
            if isinstance(row, list):
                table.add_row(*(str(value) for value in row))
        console.print(table)


def print_header(console: Any, prompt: str) -> None:
    title = Text("txGPT Terminal View", style="bold white on blue")
    console.print(Rule(title, style="blue"))
    if prompt:
        console.print(Panel(prompt, title="Prompt", border_style="blue", expand=True))


def print_plain(prompt: str, response_text: str, rows: List[Any]) -> None:
    print("=== txGPT Terminal View ===")
    if prompt:
        print("\n--- Prompt ---")
        print(prompt)
    print("\n--- Response ---")
    print(response_text)
    if rows:
        print("\n--- Extracted Data ---")
        for row in rows:
            if isinstance(row, list):
                print(" | ".join(str(value) for value in row))
            elif isinstance(row, dict):
                print(" | ".join(f"{key}={value}" for key, value in row.items()))
    print("\nTip: install Rich for graphical terminal panels: pip install rich")


def main() -> None:
    parser = argparse.ArgumentParser(description="Render txGPT JSON output with Rich.")
    parser.add_argument("--prompt", default="", help="Original prompt")
    parser.add_argument("--output", default="", help="txGPT output as JSON or plain text")
    parser.add_argument("--title", default="txGPT Response", help="Response panel title")
    args = parser.parse_args()

    response_text, rows = parse_output(args.output)
    response_text = html.unescape(response_text.replace("\\n", "\n").replace("\\r", ""))

    if not HAS_RICH:
        print_plain(args.prompt, response_text, rows)
        return

    console = Console()
    print_header(console, args.prompt)
    console.print(Panel(Markdown(response_text), title=args.title, border_style="green", expand=True))
    print_rows(console, rows)
    console.print(Rule("done", style="green"))


if __name__ == "__main__":
    main()
