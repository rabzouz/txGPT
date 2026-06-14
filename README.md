# txGPT - AI-Powered CLI Assistant for Kali Linux

txGPT is a Go-based command-line assistant that uses the OpenAI API to generate scripts, commands and technical explanations. It is tuned for Linux, Kali and authorized security workflows, while remaining usable on macOS and Windows.

<p align="center">
  <img src="images/TX_GPT.png" alt="txGPT demo" width="650">
</p>

## Features

| Capability | Notes |
| --- | --- |
| English or French output | Default: English. Use `--lang fr` for French. |
| Streaming mode | Use `--stream` to print long answers as they arrive. |
| Role presets | Use `--role "kali expert"` to steer the assistant. |
| Interactive REPL | Run `txgpt` with no prompt. Type `exit` or `quit` to leave. |
| Safe execution | `--exec` prints the generated command or code block and requires confirmation before running it. |
| JSON output | `--json` returns `{ "response": "...", "data": [...] }` for scripts and post-processing. |
| Rich display | `txgpt_rich.sh` pipes JSON output into `rich_display.py` for a formatted terminal view. |
| Nmap data extraction | Common `port/state/service` lines are extracted into JSON rows. |
| Tool presets | `--tool` focuses txGPT on PowerShell, Bash, Nmap, Git, Docker, Regex, SQL, Python or debugging. |

## Prerequisites

- Go 1.24.3 or newer.
- Git.
- An OpenAI API key in `OPENAI_API_KEY`.
- Optional: Python 3 and Rich for formatted output: `pip install rich`.

## Installation

```bash
git clone https://github.com/rabzouz/txGPT.git
cd txGPT
go mod tidy
go build -o txgpt
```

For Windows:

```powershell
go build -o txgpt.exe
```

Optional Unix install:

```bash
sudo install -m 0755 txgpt /usr/local/bin/txgpt
```

## API Key

Linux or macOS:

```bash
export OPENAI_API_KEY="sk-proj-YOUR_KEY"
```

Windows PowerShell:

```powershell
$Env:OPENAI_API_KEY = "sk-proj-YOUR_KEY"
```

You can also set `OPENAI_BASE_URL` or pass `--base-url` when using an OpenAI-compatible endpoint.

## Usage

```bash
txgpt "Generate a Bash script that backs up /var/www to /tmp."
txgpt --lang fr "Explique chmod 750 avec un exemple."
txgpt --stream --role "Kali Linux expert" "Explain a safe Nmap service scan."
txgpt --json "Show sample Nmap output with open SSH and HTTP ports."
txgpt --list-tools
txgpt --tool powershell --lang fr "Ecris un script pour lister les processus"
```

Interactive mode:

```bash
txgpt
```

Rich display:

```bash
pip install rich
chmod +x txgpt_rich.sh
./txgpt_rich.sh "Generate a safe Nmap scan command for my own lab host."
```

Safe execution:

```bash
txgpt --exec "Create a harmless command that prints the current directory."
```

`--exec` is intentionally interactive: txGPT displays what it is about to run and only continues after an explicit confirmation.

## Options

| Flag | Description |
| --- | --- |
| `--stream` | Stream tokens as they arrive. Disabled automatically for `--json`. |
| `--json` | Print machine-readable output. Cannot be combined with `--exec`. |
| `--exec` | Confirm and run the first generated code block or command. |
| `--lang en\|fr` | Choose English or French. |
| `--role "..."` | Add a role instruction to the system prompt. |
| `--tool "..."` | Use a focused tool preset. |
| `--list-tools` | List available tool presets and exit. |
| `--model "..."` | Override the default model. |
| `--temperature 0.7` | Set sampling temperature from `0.0` to `2.0`. |
| `--max-tokens 1000` | Set a response token limit. `0` lets the API choose. |
| `--base-url "..."` | Use an OpenAI-compatible base URL. |

## Tool Presets

Tool presets tune the system prompt without changing the model or requiring extra dependencies.

```bash
txgpt --list-tools
txgpt --tool powershell --lang fr "Explique Get-Process"
txgpt --tool nmap --lang fr "Propose un scan de service pour mon lab local"
txgpt --tool git "Help me understand this merge conflict"
```

Available tools:

| Tool | Focus |
| --- | --- |
| `bash` | Linux shell commands and scripts. |
| `debug` | Step-by-step troubleshooting. |
| `docker` | Dockerfiles, Compose and containers. |
| `git` | Git commands and workflows. |
| `nmap` | Authorized Nmap scanning. |
| `powershell` | Windows PowerShell commands and scripts. |
| `python` | Python scripts and automation. |
| `regex` | Regular expressions. |
| `sql` | SQL queries and database troubleshooting. |

## Development

```bash
go test ./...
go build -o txgpt
```

If Go cannot write to the default cache in a restricted environment, set a local cache first:

```bash
GOCACHE=.gocache go test ./...
```

## Troubleshooting

| Issue | Fix |
| --- | --- |
| `OPENAI_API_KEY is not set` | Export the API key in your shell before running txGPT. |
| `401 Unauthorized` | Regenerate the API key and update your environment. |
| Flags seem ignored | Rebuild the binary with `go build -o txgpt`. |
| Rich output is unavailable | Install Rich with `pip install rich`. |
| Git reports dubious ownership | Run Git as the repository owner or add a safe-directory exception. |

## License

GPL-3.0-or-later. See [LICENSE](LICENSE).

Use txGPT only on systems and networks where you have permission.
