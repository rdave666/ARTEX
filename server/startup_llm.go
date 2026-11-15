package server

import "github.com/Autumn-27/artex/agent"

// Explicit launcher settings select their endpoint even when an older active
// profile is stored in PostgreSQL. Default console behavior keeps DB precedence.
func selectStartupLLMConfig(envOnly bool, saved, env func() (agent.Config, bool)) (agent.Config, string) {
	if !envOnly {
		if cfg, ok := saved(); ok {
			return cfg, "DB"
		}
	}
	if cfg, ok := env(); ok {
		return cfg, "env"
	}
	return agent.Config{}, ""
}
