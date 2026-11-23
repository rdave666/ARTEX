package agent

import (
	"strings"
	"testing"
)

func TestRenderStalledIntentTrigger(t *testing.T) {
	got := renderTriggers([]TriggerEvent{{Kind: "stalled", IntentID: 42}}, newDataEnvelope(), triggerLookup{
		intentSummary: func(id int64) string {
			if id != 42 {
				t.Fatalf("unexpected intent id %d", id)
			}
			return "check target reachability"
		},
	})
	for _, want := range []string{"stalled intent #42", "last pending tool action", "target reachability", "Do not assume", "check target reachability"} {
		if !strings.Contains(got, want) {
			t.Fatalf("stalled recovery guidance missing %q: %s", want, got)
		}
	}
}
