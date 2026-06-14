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
	got := systemPromptFor("fr", "expert Kali")

	if !strings.Contains(got, "Adopte le rôle suivant: expert Kali.") {
		t.Fatalf("expected French role prompt, got %q", got)
	}
}
