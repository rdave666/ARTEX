package agent

import "testing"

func TestOllamaCompactionUsesRequestedRuntimeContext(t *testing.T) {
	for _, tc := range []struct{ k, want int }{{0, 32000}, {32, 32000}, {64, 64000}, {4, 4000}} {
		c := Config{Format: FormatOllama, ContextWindowK: tc.k}
		if got := c.CompactionWindow(); got != tc.want {
			t.Errorf("Ollama context %dK: got %d want %d", tc.k, got, tc.want)
		}
	}
	if got := (Config{}).CompactionWindow(); got != 200000 {
		t.Errorf("cloud default changed: %d", got)
	}
}
