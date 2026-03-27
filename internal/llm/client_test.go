package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestGenerate_GoldenFile(t *testing.T) {
	// Set a test system prompt
	SetSystemPrompt("You are a test prompt.")

	// Read the golden file — this is what our mock server will return
	golden, err := os.ReadFile("../../testdata/golden/grep.1")
	if err != nil {
		t.Fatalf("reading golden file: %v", err)
	}

	// Mock OpenAI-compatible server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/chat/completions") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}

		// Verify the request has the right structure
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decoding request: %v", err)
		}
		if len(req.Messages) != 2 {
			t.Errorf("expected 2 messages, got %d", len(req.Messages))
		}
		if req.Messages[0].Role != "system" {
			t.Errorf("expected system message first, got %s", req.Messages[0].Role)
		}
		if !strings.Contains(req.Messages[1].Content, "grep") {
			t.Errorf("expected user message to mention grep")
		}

		resp := chatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: string(golden)}},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(Config{
		APIURL: server.URL,
		APIKey: "test-key",
		Model:  "test-model",
	})

	helpText, err := os.ReadFile("../../testdata/help/grep.txt")
	if err != nil {
		t.Fatalf("reading help text: %v", err)
	}

	result, err := client.Generate(context.Background(), GenerateRequest{
		Sources: []Source{{Type: "help", Content: string(helpText)}},
		Name:    "grep",
		Section: 1,
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Validate the output has essential mdoc structure
	assertContains(t, result.Content, ".Dd ")
	assertContains(t, result.Content, ".Dt GREP 1")
	assertContains(t, result.Content, ".Os")
	assertContains(t, result.Content, ".Sh NAME")
	assertContains(t, result.Content, ".Nm grep")
	assertContains(t, result.Content, ".Sh SYNOPSIS")
	assertContains(t, result.Content, ".Sh DESCRIPTION")
	assertContains(t, result.Content, ".Sh OPTIONS")
	assertContains(t, result.Content, ".Sh EXAMPLES")
	assertContains(t, result.Content, "Fl i")
	assertContains(t, result.Content, "Fl R")
}

func TestStripCodeFences(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no fences",
			input: ".Dd March 26, 2026\n.Dt TEST 1",
			want:  ".Dd March 26, 2026\n.Dt TEST 1",
		},
		{
			name:  "mdoc fences",
			input: "```mdoc\n.Dd March 26, 2026\n.Dt TEST 1\n```",
			want:  ".Dd March 26, 2026\n.Dt TEST 1",
		},
		{
			name:  "plain fences",
			input: "```\n.Dd March 26, 2026\n```",
			want:  ".Dd March 26, 2026",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripCodeFences(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildUserPrompt(t *testing.T) {
	prompt := BuildUserPrompt(GenerateRequest{
		Sources: []Source{
			{Type: "help", Content: "usage: mytool [options]"},
			{Type: "readme", Content: "# MyTool\nA great tool."},
		},
		Name:    "mytool",
		Section: 1,
	})

	assertContains(t, prompt, "mytool")
	assertContains(t, prompt, "section 1")
	assertContains(t, prompt, "## --help output")
	assertContains(t, prompt, "usage: mytool [options]")
	assertContains(t, prompt, "## README")
	assertContains(t, prompt, "# MyTool")
}

func TestGenerateUsageStats(t *testing.T) {
	SetSystemPrompt("test prompt")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": ".Dd\n.Dt TEST 1\n.Os"}},
			},
			"usage": map[string]int{
				"prompt_tokens":     100,
				"completion_tokens": 50,
				"total_tokens":      150,
			},
		})
	}))
	defer server.Close()

	client := NewClient(Config{APIURL: server.URL, APIKey: "k", Model: "m"})
	result, err := client.Generate(context.Background(), GenerateRequest{
		Sources: []Source{{Type: "help", Content: "test"}},
		Name:    "test",
		Section: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Usage.TotalTokens != 150 {
		t.Errorf("expected 150 total tokens, got %d", result.Usage.TotalTokens)
	}
	if result.Usage.PromptTokens != 100 {
		t.Errorf("expected 100 prompt tokens, got %d", result.Usage.PromptTokens)
	}
	if result.Usage.CompletionTokens != 50 {
		t.Errorf("expected 50 completion tokens, got %d", result.Usage.CompletionTokens)
	}
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("output missing %q", substr)
	}
}
