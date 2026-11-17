package db

import (
	"regexp"
	"testing"
)

func TestModelChatBuiltinIsSelectableAssistantWithoutInteractiveShell(t *testing.T) {
	for _, a := range builtinAgents {
		if a.key == "model_chat" {
			if a.role != "assistant" || a.name != "Model chat (no tools)" || a.interactiveShell {
				t.Fatalf("unexpected lightweight chat metadata: %+v", a)
			}
			return
		}
	}
	t.Fatal("model_chat missing from seeded agent selector")
}

func TestModelChatBuiltinKeyMatchesSchema(t *testing.T) {
	constraint := regexp.MustCompile(`key\s+TEXT NOT NULL UNIQUE CHECK \(key ~ '([^']+)'\)`).FindStringSubmatch(schemaSQL)
	if len(constraint) != 2 {
		t.Fatal("agent key constraint not found in embedded schema")
	}
	valid := regexp.MustCompile(constraint[1])
	for _, a := range builtinAgents {
		if !valid.MatchString(a.key) {
			t.Errorf("builtin agent key %q violates schema %s", a.key, constraint[1])
		}
	}
}

func TestModelChatDatabaseSeed(t *testing.T) {
	d, err := Open(testDSN(t))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	var name, role string
	var builtin, interactive bool
	if err := d.QueryRow(`SELECT name,role,builtin,interactive_shell FROM agents WHERE key='model_chat'`).Scan(&name, &role, &builtin, &interactive); err != nil {
		t.Fatal(err)
	}
	if name != "Model chat (no tools)" || role != "assistant" || !builtin || interactive {
		t.Fatalf("seed name=%q role=%q builtin=%v interactive=%v", name, role, builtin, interactive)
	}
	if err := d.seedBuiltins(); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := d.QueryRow(`SELECT count(*) FROM agents WHERE key='model_chat'`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("idempotent seed count=%d err=%v", count, err)
	}
}
