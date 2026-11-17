package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDirectChatOmitsAgentToolsAndUsesOllamaChatCompletions(t *testing.T) {
	type request struct {
		Model    string            `json:"model"`
		Stream   bool              `json:"stream"`
		Tools    []json.RawMessage `json:"tools"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}
	got := make(chan request, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			http.Error(w, "wrong route", http.StatusNotFound)
			return
		}
		var body request
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		got <- body
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"fixture","choices":[{"index":0,"message":{"role":"assistant","content":"Ollama says OK"},"finish_reason":"stop"}]}`)
	}))
	defer server.Close()

	cfg := ConfigFrom("ollama", "gemma4:12b", server.URL, "", "")
	provider, err := cfg.NewProvider()
	if err != nil {
		t.Fatal(err)
	}
	chat := NewChatAgent(provider, cfg.Model, t.TempDir(), nil, cfg.CompactionWindow())
	chat.SetNonStreaming(func() bool { return true })
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reply, err := chat.ChatDirect(ctx, "auto", "conv-direct-fixture", "ping", 1, 5*time.Second, false, nil)
	if err != nil {
		t.Fatalf("direct chat failed: %v", err)
	}
	if reply != "Ollama says OK" {
		t.Fatalf("reply = %q, want Ollama says OK", reply)
	}
	select {
	case body := <-got:
		if body.Model != "gemma4:12b" || body.Stream || len(body.Tools) != 0 {
			t.Fatalf("unexpected request: model=%q stream=%v tools=%d", body.Model, body.Stream, len(body.Tools))
		}
		if len(body.Messages) != 2 || body.Messages[0].Role != "system" || body.Messages[1].Role != "user" || body.Messages[1].Content != "ping" {
			t.Fatalf("direct messages = %+v, want simple system + user without tools", body.Messages)
		}
	default:
		t.Fatal("direct chat never sent an HTTP request")
	}
}
