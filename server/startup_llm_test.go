package server

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Autumn-27/artex/agent"
	"github.com/Autumn-27/artex/db"
	"github.com/Autumn-27/norma/llm"
)

func TestLauncherOverrideDoesNotUseSavedPoolAndReportsEffectiveModel(t *testing.T) {
	skipWithoutDSN(t)
	var fallbackHits atomic.Int32
	fallback := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fallbackHits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"1","choices":[{"index":0,"message":{"role":"assistant","content":"unwanted fallback"},"finish_reason":"stop"}]}`))
	}))
	defer fallback.Close()
	custom := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusPaymentRequired)
		_, _ = w.Write([]byte(`{"error":"insufficient_quota"}`))
	}))
	defer custom.Close()
	m, err := NewManager(t.TempDir(), "")
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	for _, name := range []string{"launcher-old-active", "launcher-old-fallback"} {
		id, err := m.pg.SaveProfile(&db.LLMProfile{Name: name, Format: "openai", Model: "old-cloud-model", BaseURL: fallback.URL + "/v1", APIKey: "fixture-cloud-key"})
		if err != nil {
			t.Fatal(err)
		}
		defer func(id int64) {
			_, _ = m.pg.Exec(`UPDATE llm_profiles SET is_default=false WHERE id=$1`, id)
			_ = m.pg.DeleteProfile(id)
		}(id)
		if name == "launcher-old-active" {
			if err := m.pg.SetActiveProfile(id); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := m.SetLLMPoolEnabled(true); err != nil {
		t.Fatal(err)
	}
	defer m.SetLLMPoolEnabled(false)
	t.Setenv("ARTEX_LLM_OVERRIDE", "true")
	t.Setenv("ARTEX_LLM_PROVIDER", "openai")
	t.Setenv("ARTEX_LLM_NO_AUTH", "true")
	t.Setenv("ARTEX_LLM_BASE_URL", custom.URL+"/v1")
	t.Setenv("ARTEX_LLM_MODEL", "launcher-model")
	ctx, cancel := context.WithCancel(context.Background())
	dir := t.TempDir()
	s := New(ctx, m, dir, dir, dir)
	defer stopTestServer(t, s, cancel)
	resolution, err := s.resolveTaskRoleLLM(&Task{}, "worker")
	if err != nil {
		t.Fatal(err)
	}
	if resolution.Model != "launcher-model" || resolution.Source != "environment" {
		t.Errorf("reported %+v instead of the effective launcher provider", resolution)
	}
	provider, _, ok := s.globalProvider()
	if !ok {
		t.Fatal("launcher provider unavailable")
	}
	policyRecorder := httptest.NewRecorder()
	s.pgSaveLLMRetryPolicy(policyRecorder, httptest.NewRequest(http.MethodPut, "/api/llm/retry", bytes.NewBufferString(`{"connect":{"attempts":1}}`)))
	if policyRecorder.Code != http.StatusOK {
		t.Fatalf("retry policy update: HTTP %d", policyRecorder.Code)
	}
	provider, after, ok := s.globalProvider()
	if !ok || after.Model != "launcher-model" || after.BaseURL != custom.URL+"/v1" || !after.NoAuth {
		t.Fatal("retry policy refresh replaced the explicitly selected launcher endpoint")
	}
	requestCtx, stop := context.WithTimeout(context.Background(), 3*time.Second)
	defer stop()
	_, _, _, err = provider.Complete(requestCtx, llm.CompletionRequest{Messages: []llm.Message{llm.UserText("ping")}})
	if err == nil || fallbackHits.Load() != 0 {
		t.Fatalf("explicit launcher provider must fail on its own endpoint; fallback requests=%d err=%v", fallbackHits.Load(), err)
	}
}

func TestStartupLLMExplicitEnvironmentDoesNotUseSavedProfile(t *testing.T) {
	saved := func() (agent.Config, bool) {
		t.Fatal("saved profile must not override explicit Android settings")
		return agent.Config{}, false
	}
	env := func() (agent.Config, bool) { return agent.Config{Model: "android-model"}, true }
	cfg, source := selectStartupLLMConfig(true, saved, env)
	if source != "env" || cfg.Model != "android-model" {
		t.Fatalf("source=%q model=%q", source, cfg.Model)
	}
	_, source = selectStartupLLMConfig(true, saved, func() (agent.Config, bool) { return agent.Config{}, false })
	if source != "" {
		t.Fatal("invalid explicit settings must not fall back to a different saved provider")
	}
}

func TestStartupLLMDefaultPreservesSavedProfile(t *testing.T) {
	saved := func() (agent.Config, bool) { return agent.Config{Model: "saved-model"}, true }
	env := func() (agent.Config, bool) {
		t.Fatal("saved profile should retain default precedence")
		return agent.Config{}, false
	}
	cfg, source := selectStartupLLMConfig(false, saved, env)
	if source != "DB" || cfg.Model != "saved-model" {
		t.Fatalf("source=%q model=%q", source, cfg.Model)
	}
}
