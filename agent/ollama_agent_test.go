package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// Exercise the real tool harness, not only provider serialization: the model
// requests a harmless Bash command, ARTEX executes it and returns its result.
func TestOllamaFullAgentExecutesToolAndReturnsResult(t *testing.T) {
	for _, stream := range []bool{false, true} {
		t.Run(fmt.Sprint(stream), func(t *testing.T) {
			var mu sync.Mutex
			var requests []map[string]json.RawMessage
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body map[string]json.RawMessage
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
					return
				}
				mu.Lock()
				requests = append(requests, body)
				n := len(requests)
				mu.Unlock()
				if r.URL.Path != "/api/chat" {
					t.Errorf("path %s", r.URL.Path)
				}
				msg := `{"role":"assistant","content":"TOOL_VERIFIED"}`
				if n == 1 {
					msg = `{"role":"assistant","content":"","tool_calls":[{"function":{"name":"Bash","arguments":{"command":"printf ARTEX_OLLAMA_TOOL_PROOF","description":"Print a harmless verification marker"}}}]}`
				}
				w.Header().Set("Content-Type", "application/x-ndjson")
				fmt.Fprintf(w, `{"message":%s,"done":true,"done_reason":"stop","prompt_eval_count":12430,"eval_count":12}`+"\n", msg)
			}))
			defer srv.Close()
			cfg := ConfigFrom("ollama", "tool-model", srv.URL, "", "")
			cfg.Stream = stream
			cfg.ContextWindowK = 32
			p, err := cfg.NewProvider()
			if err != nil {
				t.Fatal(err)
			}
			chat := NewChatAgent(p, "tool-model", t.TempDir(), nil, cfg.CompactionWindow())
			chat.SetNonStreaming(func() bool { return !stream })
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			reply, err := chat.Chat(ctx, "auto", "ollama-agent-test", "Print a verification marker using Bash then confirm it.", 3, 15*time.Second, false, nil)
			if err != nil || reply != "TOOL_VERIFIED" {
				t.Fatalf("reply=%q err=%v", reply, err)
			}
			mu.Lock()
			defer mu.Unlock()
			if len(requests) != 2 {
				t.Fatalf("requests=%d want2", len(requests))
			}
			if len(requests[0]["tools"]) < 100 {
				t.Fatal("full tools missing")
			}
			var opts struct {
				NumCtx int `json:"num_ctx"`
			}
			json.Unmarshal(requests[0]["options"], &opts)
			if opts.NumCtx != 32000 {
				t.Fatalf("num_ctx=%d", opts.NumCtx)
			}
			var msgs []struct {
				Role    string `json:"role"`
				Name    string `json:"tool_name"`
				Content string `json:"content"`
			}
			if err := json.Unmarshal(requests[1]["messages"], &msgs); err != nil {
				t.Fatal(err)
			}
			found := false
			for _, m := range msgs {
				if m.Role == "tool" && m.Name == "Bash" && strings.Contains(m.Content, "ARTEX_OLLAMA_TOOL_PROOF") {
					found = true
				}
			}
			if !found {
				t.Fatalf("tool result not returned to model: %s", requests[1]["messages"])
			}
		})
	}
}
