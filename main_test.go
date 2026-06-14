package main

import (
	"strings"
	"testing"
)

func TestExtractStructuredDataFromNmapOutput(t *testing.T) {
	input := `PORT     STATE    SERVICE
22/tcp   open     ssh
80/tcp   open     http
443/tcp  filtered https
70000/tcp open    invalid`

	got := extractStructuredData(input)
	want := [][]string{
		{"22/tcp", "open", "ssh"},
		{"80/tcp", "open", "http"},
		{"443/tcp", "filtered", "https"},
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d rows, got %d: %#v", len(want), len(got), got)
	}
	for i := range want {
		for j := range want[i] {
			if got[i][j] != want[i][j] {
				t.Fatalf("row %d column %d: expected %q, got %q", i, j, want[i][j], got[i][j])
			}
		}
	}
}

func TestExtractFirstCodeBlock(t *testing.T) {
	input := "Run this:\n```bash\nnmap -sV 127.0.0.1\n```\nDone."

	got := extractFirstCodeBlock(input)
	if got != "nmap -sV 127.0.0.1" {
		t.Fatalf("unexpected code block: %q", got)
	}
}

func TestSystemPromptLanguageAndRole(t *testing.T) {
	got, err := systemPromptFor("fr", "expert Kali", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "Adopte le rôle suivant: expert Kali.") {
		t.Fatalf("expected French role prompt, got %q", got)
	}
}

func TestSystemPromptToolPreset(t *testing.T) {
	got, err := systemPromptFor("fr", "", "powershell")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "Outil actif: powershell.") {
		t.Fatalf("expected PowerShell tool prompt, got %q", got)
	}
	if !strings.Contains(got, "Windows PowerShell") {
		t.Fatalf("expected PowerShell guidance, got %q", got)
	}
}

func TestSystemPromptUnknownTool(t *testing.T) {
	_, err := systemPromptFor("en", "", "unknown")
	if err == nil {
		t.Fatal("expected an unknown tool error")
	}
	if !strings.Contains(err.Error(), "run --list-tools") {
		t.Fatalf("unexpected error: %v", err)
	}
}
