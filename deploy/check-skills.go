// Read-only verification using the same skill library as ARTEX.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Autumn-27/norma/skill"
)

func main() {
	if len(os.Args) != 2 { panic("usage: check-skills SKILL_DIR") }
	reg, err := skill.LoadDir(os.Args[1])
	if err != nil { panic(err) }
	var verified []map[string]any
	for _, s := range reg.List() {
		if s.Dir == "" { continue }
		if s.Name != filepath.Base(s.Dir) || s.Description == "" || s.Instructions == "" {
			panic(fmt.Sprintf("invalid parsed skill: %s", s.Dir))
		}
		input, _ := json.Marshal(map[string]string{"name": s.Name})
		result, err := reg.Tool().Call(context.Background(), input, nil)
		if err != nil || result.IsError || len(result.Extra) == 0 {
			panic(fmt.Sprintf("cannot load skill %s: %v", s.Name, err))
		}
		if len(s.Name) >= 10 && s.Name[:10] == "offensive-" {
			ref, err := os.ReadFile(filepath.Join(s.Dir, "references", "methodology.md"))
			if err != nil || len(ref) == 0 { panic(fmt.Sprintf("missing reference: %s", s.Name)) }
		}
		verified = append(verified, map[string]any{"name": s.Name, "description_parsed": true, "skill_invocation": "passed", "mcps": s.MCPs})
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(verified); err != nil { panic(err) }
}
