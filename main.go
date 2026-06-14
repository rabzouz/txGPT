package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/rabzouz/txGPT/internal/ai"
	openai "github.com/sashabaranov/go-openai"
)

type jsonResponse struct {
	Response string     `json:"response"`
	Data     [][]string `json:"data,omitempty"`
}

type toolPreset struct {
	Name        string
	Description string
	English     string
	French      string
}

var toolPresets = map[string]toolPreset{
	"bash": {
		Name:        "bash",
		Description: "Linux shell commands and scripts",
		English:     "Focus on Bash and POSIX shell commands. Prefer portable, readable commands and explain risky flags.",
		French:      "Concentre-toi sur Bash et les commandes shell POSIX. Prefere des commandes lisibles et portables, et explique les options risquées.",
	},
	"debug": {
		Name:        "debug",
		Description: "Step-by-step troubleshooting",
		English:     "Act as a debugging assistant. Ask for observable symptoms when needed, isolate causes, and give verification steps.",
		French:      "Agis comme un assistant de diagnostic. Demande les symptomes observables si necessaire, isole les causes et donne des etapes de verification.",
	},
	"docker": {
		Name:        "docker",
		Description: "Dockerfiles, Compose and containers",
		English:     "Focus on Docker, Dockerfiles, Compose, images, containers, volumes and networking. Prefer reproducible commands.",
		French:      "Concentre-toi sur Docker, Dockerfile, Compose, images, conteneurs, volumes et reseau. Prefere des commandes reproductibles.",
	},
	"git": {
		Name:        "git",
		Description: "Git commands and workflows",
		English:     "Focus on Git workflows. Protect user work, avoid destructive commands unless explicitly requested, and explain branch state clearly.",
		French:      "Concentre-toi sur les workflows Git. Protege le travail utilisateur, evite les commandes destructrices sauf demande explicite, et explique clairement l'etat des branches.",
	},
	"nmap": {
		Name:        "nmap",
		Description: "Authorized Nmap scanning",
		English:     "Focus on authorized Nmap usage. Keep commands scoped, explain scan impact, and remind the user to scan only permitted targets.",
		French:      "Concentre-toi sur l'usage autorise de Nmap. Garde les commandes ciblees, explique l'impact du scan et rappelle de scanner uniquement les cibles autorisees.",
	},
	"powershell": {
		Name:        "powershell",
		Description: "Windows PowerShell commands and scripts",
		English:     "Focus on Windows PowerShell. Prefer safe cmdlets, include commands that work in current PowerShell, and explain administrator requirements.",
		French:      "Concentre-toi sur Windows PowerShell. Prefere les cmdlets sures, donne des commandes compatibles PowerShell, et explique les besoins administrateur.",
	},
	"python": {
		Name:        "python",
		Description: "Python scripts and automation",
		English:     "Focus on Python scripts. Prefer standard library solutions first, include clear error handling, and keep examples runnable.",
		French:      "Concentre-toi sur les scripts Python. Prefere d'abord la bibliotheque standard, ajoute une gestion d'erreur claire et garde les exemples executables.",
	},
	"regex": {
		Name:        "regex",
		Description: "Regular expressions",
		English:     "Focus on regular expressions. State the target regex flavor, explain each group, and provide positive and negative examples.",
		French:      "Concentre-toi sur les expressions regulieres. Precise le moteur regex cible, explique chaque groupe et donne des exemples positifs et negatifs.",
	},
	"sql": {
		Name:        "sql",
		Description: "SQL queries and database troubleshooting",
		English:     "Focus on SQL. Prefer parameterized queries, mention indexes when relevant, and avoid unsafe data modification without a backup step.",
		French:      "Concentre-toi sur SQL. Prefere les requetes parametrees, mentionne les index si utile, et evite les modifications dangereuses sans etape de sauvegarde.",
	},
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cfg := ai.DefaultConfig()

	flags := flag.NewFlagSet("txgpt", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	stream := flags.Bool("stream", false, "stream the answer as it is generated")
	asJSON := flags.Bool("json", false, "print a JSON object with response and extracted data")
	execute := flags.Bool("exec", false, "ask for confirmation, then execute the first generated code block")
	lang := flags.String("lang", "en", "answer language: en or fr")
	role := flags.String("role", "", "role preset appended to the system prompt")
	tool := flags.String("tool", "", "tool preset to use; run --list-tools to see available tools")
	listTools := flags.Bool("list-tools", false, "list available tool presets and exit")
	model := flags.String("model", cfg.Model, "OpenAI model name")
	temperature := flags.Float64("temperature", 0.7, "sampling temperature from 0.0 to 2.0")
	maxTokens := flags.Int("max-tokens", cfg.MaxTokens, "maximum output tokens; 0 lets the API choose")
	baseURL := flags.String("base-url", os.Getenv("OPENAI_BASE_URL"), "optional OpenAI-compatible API base URL")

	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if *listTools {
		printToolList()
		return nil
	}
	if *execute && *asJSON {
		return fmt.Errorf("--exec cannot be combined with --json")
	}
	if *temperature < 0 || *temperature > 2 {
		return fmt.Errorf("--temperature must be between 0.0 and 2.0")
	}
	if *maxTokens < 0 {
		return fmt.Errorf("--max-tokens cannot be negative")
	}

	cfg.Stream = *stream
	cfg.Model = strings.TrimSpace(*model)
	cfg.Temperature = float32(*temperature)
	cfg.MaxTokens = *maxTokens
	cfg.BaseURL = strings.TrimSpace(*baseURL)
	systemPrompt, err := systemPromptFor(*lang, *role, *tool)
	if err != nil {
		return err
	}
	cfg.SystemPrompt = systemPrompt
	if *asJSON {
		cfg.Stream = false
	}

	prompt := strings.TrimSpace(strings.Join(flags.Args(), " "))
	if prompt == "" {
		return repl(cfg, *asJSON, *execute)
	}

	response, err := askAndPrint(prompt, nil, cfg, *asJSON)
	if err != nil {
		return err
	}
	if *execute {
		return confirmAndRun(response)
	}
	return nil
}

func repl(cfg ai.Config, asJSON bool, execute bool) error {
	fmt.Fprintln(os.Stderr, "txGPT interactive mode. Type exit or quit to leave.")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024), 1024*1024)

	var history []openai.ChatCompletionMessage
	for {
		fmt.Fprint(os.Stderr, "txgpt> ")
		if !scanner.Scan() {
			break
		}

		prompt := strings.TrimSpace(scanner.Text())
		if prompt == "" {
			continue
		}
		switch strings.ToLower(prompt) {
		case "exit", "quit":
			return nil
		}

		response, err := askAndPrint(prompt, history, cfg, asJSON)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			continue
		}

		history = append(history,
			openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: prompt},
			openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: response},
		)

		if execute {
			if err := confirmAndRun(response); err != nil {
				return err
			}
		}
	}

	return scanner.Err()
}

func askAndPrint(prompt string, history []openai.ChatCompletionMessage, cfg ai.Config, asJSON bool) (string, error) {
	response, err := ai.Ask(prompt, history, cfg)
	if err != nil {
		return "", err
	}

	if asJSON {
		return response, json.NewEncoder(os.Stdout).Encode(jsonResponse{
			Response: response,
			Data:     extractStructuredData(response),
		})
	}

	if !cfg.Stream {
		fmt.Println(response)
	}
	return response, nil
}

func systemPromptFor(lang string, role string, tool string) (string, error) {
	normalizedLang := strings.ToLower(strings.TrimSpace(lang))
	normalizedRole := strings.TrimSpace(role)
	normalizedTool := strings.ToLower(strings.TrimSpace(tool))

	var prompt string
	switch normalizedLang {
	case "fr", "fra", "fre", "french", "francais", "français":
		prompt = "Tu es txGPT, un assistant technique concis pour le code, Linux et les tests de sécurité autorisés. Refuse les demandes qui facilitent un usage nuisible ou non autorisé."
		if normalizedRole != "" {
			prompt += " Adopte le rôle suivant: " + normalizedRole + "."
		}
	default:
		prompt = "You are txGPT, a concise technical assistant for code, Linux and authorized security testing. Refuse requests that facilitate unauthorized harm."
		if normalizedRole != "" {
			prompt += " Adopt this role: " + normalizedRole + "."
		}
	}

	if normalizedTool == "" {
		return prompt, nil
	}

	preset, ok := toolPresets[normalizedTool]
	if !ok {
		return "", fmt.Errorf("unknown tool %q; run --list-tools", tool)
	}
	if isFrench(normalizedLang) {
		prompt += " Outil actif: " + preset.Name + ". " + preset.French
	} else {
		prompt += " Active tool: " + preset.Name + ". " + preset.English
	}
	return prompt, nil
}

func isFrench(normalizedLang string) bool {
	switch normalizedLang {
	case "fr", "fra", "fre", "french", "francais", "français":
		return true
	default:
		return false
	}
}

func printToolList() {
	names := make([]string, 0, len(toolPresets))
	for name := range toolPresets {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Println("Available txGPT tools:")
	for _, name := range names {
		preset := toolPresets[name]
		fmt.Printf("  %-10s %s\n", preset.Name, preset.Description)
	}
}

func extractStructuredData(text string) [][]string {
	re := regexp.MustCompile(`(?im)^\s*(\d{1,5})(?:/(tcp|udp))?\s+((?:open|closed|filtered|unfiltered)(?:\|filtered)?)\s+([a-z0-9._+-]+)\b`)
	matches := re.FindAllStringSubmatch(text, -1)

	rows := make([][]string, 0, len(matches))
	for _, match := range matches {
		portNumber, err := strconv.Atoi(match[1])
		if err != nil || portNumber < 1 || portNumber > 65535 {
			continue
		}

		port := match[1]
		if match[2] != "" {
			port += "/" + strings.ToLower(match[2])
		}
		rows = append(rows, []string{port, strings.ToLower(match[3]), match[4]})
	}
	return rows
}

func extractFirstCodeBlock(text string) string {
	start := strings.Index(text, "```")
	if start == -1 {
		return strings.TrimSpace(text)
	}

	rest := text[start+3:]
	if newline := strings.Index(rest, "\n"); newline >= 0 {
		rest = rest[newline+1:]
	}

	end := strings.Index(rest, "```")
	if end == -1 {
		return strings.TrimSpace(rest)
	}
	return strings.TrimSpace(rest[:end])
}

func confirmAndRun(response string) error {
	commandText := extractFirstCodeBlock(response)
	if commandText == "" {
		return fmt.Errorf("no command or code block found to execute")
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Generated command or script:")
	fmt.Fprintln(os.Stderr, commandText)
	fmt.Fprint(os.Stderr, "Run it locally? Type yes to continue: ")

	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "yes", "y", "oui", "o":
		return runShell(commandText)
	default:
		fmt.Fprintln(os.Stderr, "Execution cancelled.")
		return nil
	}
}

func runShell(commandText string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-NoProfile", "-Command", commandText)
	} else {
		cmd = exec.Command("sh", "-c", commandText)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
