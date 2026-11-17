package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Autumn-27/norma/llm"
)

func TestNativeOllamaContextAndToolRoundTrip(t *testing.T) {
	var bodies []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var b map[string]any
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			t.Error(err)
		}
		bodies = append(bodies, b)
		if r.URL.Path != "/gateway/api/chat" {
			http.Error(w, "native path required", 404)
			return
		}
		if b["options"].(map[string]any)["num_ctx"] != float64(32000) {
			http.Error(w, "context remains4096", 400)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if len(bodies) == 1 {
			fmt.Fprint(w, `{"message":{"role":"assistant","content":"","tool_calls":[{"function":{"name":"Read","arguments":{"file_path":"fixture.txt"}}}]},"done":true,"done_reason":"stop","prompt_eval_count":12000,"eval_count":20}`)
		} else {
			fmt.Fprint(w, `{"message":{"role":"assistant","content":"read complete"},"done":true,"done_reason":"stop"}`)
		}
	}))
	defer server.Close()
	cfg := Config{BaseURL: server.URL + "/gateway/v1", Model: "fixture", Retry: RetryConfig{ConnectAttempts: -1}}
	p := newOllamaProvider(cfg, server.Client())
	req := llm.CompletionRequest{System: []string{"full agent instructions"}, Messages: []llm.Message{llm.UserText("read fixture")}, Tools: []llm.ToolSchema{{Name: "Read", Description: "read a fixture", InputSchema: map[string]any{"type": "object"}}}, MaxTokens: 256}
	answer, stop, usage, err := p.Complete(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}
	if stop != "tool_use" || len(answer.ToolUses()) != 1 || usage.InputTokens != 12000 {
		t.Fatalf("answer=%+v stop=%q usage=%+v", answer, stop, usage)
	}
	tool := answer.ToolUses()[0]
	req.Messages = append(req.Messages, answer, llm.Message{Role: llm.RoleUser, Content: []llm.ContentBlock{llm.ToolResultText(tool.ID, "fixture data", false)}})
	answer, stop, _, err = p.Complete(t.Context(), req)
	if err != nil || answer.Text() != "read complete" || stop != "end_turn" {
		t.Fatalf("answer=%+v stop=%q err=%v", answer, stop, err)
	}
	sent := bodies[1]
	if sent["truncate"] != false || sent["stream"] != false || len(sent["tools"].([]any)) != 1 {
		t.Fatalf("agent settings lost: %+v", sent)
	}
	messages := sent["messages"].([]any)
	result := messages[len(messages)-1].(map[string]any)
	if result["role"] != "tool" || result["tool_name"] != "Read" || result["content"] != "fixture data" {
		t.Fatalf("bad tool result: %+v", result)
	}
}

func TestNativeOllamaStreamThinkingToolsAndUsage(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["think"] != false || body["stream"] != true {
			t.Errorf("body=%+v", body)
		}
		w.Header().Set("Content-Type", "application/x-ndjson")
		fmt.Fprintln(w, `{"message":{"thinking":"checking"},"done":false}`)
		fmt.Fprintln(w, `{"message":{"content":"working","tool_calls":[{"function":{"name":"Read","arguments":{"file_path":"a"}}},{"function":{"name":"Read","arguments":{"file_path":"b"}}}]},"done":false}`)
		fmt.Fprintln(w, `{"message":{},"done":true,"done_reason":"stop","prompt_eval_count":42,"eval_count":9}`)
	}))
	defer s.Close()
	p := newOllamaProvider(Config{BaseURL: s.URL, Model: "fixture", ThinkingType: "enabled", ReasoningEffort: "high"}, s.Client())
	for round := range 2 {
		acc := llm.NewAccumulator()
		for ev, err := range p.Stream(t.Context(), llm.CompletionRequest{Thinking: "disabled", Messages: []llm.Message{llm.UserText("test")}}) {
			if err != nil {
				t.Fatal(err)
			}
			acc.Add(ev)
		}
		msg := acc.Message()
		tools := msg.ToolUses()
		if msg.Text() != "working" || len(tools) != 2 || tools[0].ID == tools[1].ID || !strings.HasSuffix(tools[0].ID, fmt.Sprintf("_%d_0", round+1)) || acc.StopReason != "tool_use" || acc.Usage.InputTokens != 42 || acc.Usage.OutputTokens != 9 {
			t.Fatalf("msg=%+v stop=%q usage=%+v", msg, acc.StopReason, acc.Usage)
		}
	}
}

func TestNativeOllamaStreamRejectsTruncationAndErrors(t *testing.T) {
	for _, test := range []struct{ name, body string }{{"eof", `{"message":{"content":"partial"},"done":false}`}, {"error", `{"error":"model unavailable"}`}, {"malformed", `{bad`}} {
		t.Run(test.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, test.body) }))
			defer s.Close()
			p := newOllamaProvider(Config{BaseURL: s.URL, Model: "fixture"}, s.Client())
			var got error
			for _, err := range p.Stream(t.Context(), llm.CompletionRequest{}) {
				if err != nil {
					got = err
				}
			}
			if got == nil {
				t.Fatal("expected terminal stream error")
			}
		})
	}
}

func TestNativeOllamaCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p := newOllamaProvider(Config{BaseURL: "http://127.0.0.1:1", Model: "fixture"}, http.DefaultClient)
	started := time.Now()
	_, _, _, err := p.Complete(ctx, llm.CompletionRequest{})
	if err == nil || time.Since(started) > time.Second {
		t.Fatalf("canceled request err=%v", err)
	}
}

func TestNativeOllamaRetriesOnlyBeforeConsumption(t *testing.T) {
	calls := 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			http.Error(w, "busy", http.StatusTooManyRequests)
			return
		}
		fmt.Fprint(w, `{"message":{"content":"OK"},"done":true}`)
	}))
	defer s.Close()
	p := newOllamaProvider(Config{BaseURL: s.URL, Model: "fixture", Retry: RetryConfig{ConnectAttempts: 1, ConnectInterval: time.Millisecond}}, s.Client())
	answer, _, _, err := p.Complete(t.Context(), llm.CompletionRequest{})
	if err != nil || answer.Text() != "OK" || calls != 2 {
		t.Fatalf("answer=%+v err=%v calls=%d", answer, err, calls)
	}
	calls = 0
	for _, err := range p.Stream(t.Context(), llm.CompletionRequest{}) {
		if err != nil {
			t.Fatal(err)
		}
	}
	if calls != 2 {
		t.Fatalf("retry count=%d", calls)
	}
}

func TestNativeOllamaURLNormalization(t *testing.T) {
	for _, suffix := range []string{"", "/v1", "/v1/chat/completions", "/api/chat", "/api/chat/v1", "/api/generate/v1", "/v1/chat/completions/"} {
		if got := ollamaNativeURL("http://host:11434/gateway" + suffix); got != "http://host:11434/gateway/api/chat" {
			t.Errorf("suffix %q: %s", suffix, got)
		}
	}
}

func TestNativeOllamaProfileOptionsAndErrorRedaction(t *testing.T) {
	cfg := Config{BaseURL: "http://host/v1", Model: "fixture", ContextWindowK: 64, APIKey: "fixture-private-key", ThinkingType: "enabled", ReasoningEffort: "high"}
	p := newOllamaProvider(cfg, http.DefaultClient).(*ollamaProvider)
	temperature := 0.2
	raw, err := p.requestBody(llm.CompletionRequest{MaxTokens: 512, Temperature: &temperature, Stop: []string{"END"}}, false)
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	_ = json.Unmarshal(raw, &body)
	options := body["options"].(map[string]any)
	if options["num_ctx"] != float64(64000) || options["num_predict"] != float64(512) || body["think"] != "high" {
		t.Fatalf("body=%s", raw)
	}
	err = p.emitChunk(ollamaChunk{Error: "invalid key fixture-private-key"}, 1, new(int), func(llm.StreamEvent) bool { return true })
	if err == nil || strings.Contains(err.Error(), cfg.APIKey) {
		t.Fatalf("error not redacted: %v", err)
	}
}
