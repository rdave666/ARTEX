package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// The connection test must test the model, not silently install the SDK's
// default filesystem/shell tools (nil Tools means DefaultTools in agentcore).
func TestConnectionIsLightweightAndNeverRunsTools(t *testing.T) {
	for _, stream := range []bool{false, true} {
		t.Run(fmt.Sprint(stream), func(t *testing.T) {
			var body map[string]json.RawMessage
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
					return
				}
				if stream {
					w.Header().Set("Content-Type", "text/event-stream")
					fmt.Fprint(w, "data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"OK\"},\"finish_reason\":null}]}\n\ndata: {\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"stop\"}]}\n\ndata: [DONE]\n\n")
				} else {
					w.Header().Set("Content-Type", "application/json")
					fmt.Fprint(w, `{"choices":[{"message":{"role":"assistant","content":"OK"},"finish_reason":"stop"}]}`)
				}
			}))
			defer upstream.Close()
			cfg := ConfigFrom("ollama", "small-model", upstream.URL, "", "")
			cfg.Stream = stream
			cfg.MaxTokens = 128
			_, reply, err := TestConnection(context.Background(), cfg)
			if err != nil || reply != "OK" {
				t.Fatalf("reply=%q err=%v", reply, err)
			}
			if tools := body["tools"]; len(tools) > 0 && string(tools) != "[]" {
				t.Errorf("connection test unexpectedly advertised tools (%d bytes)", len(tools))
			}
			if string(body["max_tokens"]) != "128" {
				t.Errorf("test ignored profile output cap: %s", body["max_tokens"])
			}
			if len(body["messages"]) > 1024 {
				t.Errorf("unexpected agent prompt injected: %d bytes", len(body["messages"]))
			}
			if string(body["stream"]) != fmt.Sprint(stream) {
				t.Errorf("stream setting ignored: %s", body["stream"])
			}
		})
	}
}
