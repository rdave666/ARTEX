package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Autumn-27/norma/llm"
)

func TestKeylessCustomOpenAIUsesExactEndpointWithoutCloudKey(t *testing.T) {
	type request struct{ path, authorization, model string }
	requests := make(chan request, 1)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Model string `json:"model"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		requests <- request{r.URL.Path, r.Header.Get("Authorization"), body.Model}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"1","choices":[{"index":0,"message":{"role":"assistant","content":"OK"},"finish_reason":"stop"}]}`))
	}))
	defer upstream.Close()
	t.Setenv("ARTEX_LLM_PROVIDER", "openai")
	t.Setenv("ARTEX_LLM_BASE_URL", upstream.URL+"/gateway/v1")
	t.Setenv("ARTEX_LLM_MODEL", "local-model")
	t.Setenv("ARTEX_LLM_NO_AUTH", "true")
	t.Setenv("OPENAI_API_KEY", "fixture-cloud-key-must-not-leak")
	cfg, ok := FromEnv()
	if !ok || cfg.APIKey != "" || cfg.NeedsAPIKey() {
		t.Fatal("explicit keyless custom endpoint must be usable without inheriting a cloud key")
	}
	provider, err := cfg.NewProvider()
	if err != nil {
		t.Fatal(err)
	}
	_, _, _, err = provider.Complete(context.Background(), llm.CompletionRequest{Messages: []llm.Message{llm.UserText("ping")}})
	if err != nil {
		t.Fatal(err)
	}
	got := <-requests
	if got.path != "/gateway/v1/chat/completions" || got.authorization != "" || got.model != "local-model" {
		t.Fatalf("request path=%q auth=%q model=%q", got.path, got.authorization, got.model)
	}
}

func TestKeylessCustomOpenAIStreamingOmitsAuthorization(t *testing.T) {
	auth := make(chan string, 1)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth <- r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"id\":\"s1\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"OK\"},\"finish_reason\":null}]}\n\n")
		fmt.Fprint(w, "data: {\"id\":\"s1\",\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"stop\"}]}\n\ndata: [DONE]\n\n")
	}))
	defer upstream.Close()
	t.Setenv("OPENAI_API_KEY", "fixture-inherited-cloud-key")
	cfg := ConfigFrom("openai", "local-model", upstream.URL+"/v1", "", "")
	cfg.NoAuth = true
	provider, err := cfg.NewProvider()
	if err != nil {
		t.Fatal(err)
	}
	text := ""
	for event, err := range provider.Stream(context.Background(), llm.CompletionRequest{Messages: []llm.Message{llm.UserText("ping")}}) {
		if err != nil {
			t.Fatal(err)
		}
		if event.Type == llm.SETextDelta {
			text += event.Text
		}
	}
	if text != "OK" || <-auth != "" {
		t.Fatal("streaming keyless response or Authorization isolation failed")
	}
}

func TestKeylessOpenAIRejectsMissingOrInvalidEndpoint(t *testing.T) {
	t.Setenv("ARTEX_LLM_PROVIDER", "openai")
	t.Setenv("ARTEX_LLM_NO_AUTH", "true")
	t.Setenv("OPENAI_API_KEY", "fixture-cloud-key")
	for _, endpoint := range []string{"", "/v1", "ftp://localhost/v1", "http://user:pass@localhost/v1"} {
		t.Run(endpoint, func(t *testing.T) {
			t.Setenv("ARTEX_LLM_BASE_URL", endpoint)
			if _, ok := FromEnv(); ok {
				t.Fatal("keyless mode must require an explicit HTTP(S) endpoint without URL credentials")
			}
		})
	}
}

func TestOfficialOpenAIPreservesConfiguredKey(t *testing.T) {
	t.Setenv("ARTEX_LLM_PROVIDER", "openai")
	t.Setenv("ARTEX_LLM_NO_AUTH", "")
	t.Setenv("OPENAI_API_KEY", "fixture-official-key")
	t.Setenv("ARTEX_LLM_BASE_URL", "")
	cfg, ok := FromEnv()
	if !ok || cfg.APIKey != "fixture-official-key" || !cfg.NeedsAPIKey() {
		t.Fatal("official OpenAI must retain its explicit key requirement")
	}
}
