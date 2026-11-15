# Android launcher configuration

The Android supervisor executes the engine from Android's read-only native
library directory. Set `ARTEX_JWT_KEY_DIR` to a private writable directory
outside the browsable data directory; it holds both `jwt.key` and `secret.key`.
The Android app uses its private `files/engine-keys` directory and preserves
these files across restarts and app updates.

The normal OpenAI default is `https://api.openai.com/v1`, using Chat Completions.
Custom OpenAI-compatible gateways use `ARTEX_LLM_BASE_URL` and `ARTEX_LLM_MODEL`.

For an explicitly keyless endpoint, including Ollama's compatible API:

```sh
ARTEX_LLM_PROVIDER=openai
ARTEX_LLM_BASE_URL=http://127.0.0.1:11434/v1
ARTEX_LLM_MODEL=qwen3:8b
ARTEX_LLM_NO_AUTH=true
```

Keyless mode requires an explicit HTTP(S) base URL. It ignores inherited
OpenAI credentials and removes the Authorization header for both streaming
and nonstreaming requests. It never falls back to the public OpenAI endpoint.
With `ARTEX_LLM_NO_AUTH` unset/false, custom endpoints use the explicitly
configured `OPENAI_API_KEY` as before.

`ARTEX_LLM_OVERRIDE=true` selects the launcher's environment configuration as
the global provider at startup, even when PostgreSQL has an older active
profile. This global configuration does not inherit that profile's failover
pool. Stored profiles remain unchanged; explicit task chains and agent or
conversation profile bindings retain their precedence. Without the override,
the original saved-profile-first behavior is preserved.

Editing retry policy or unrelated profiles preserves the selected launcher
endpoint. Explicitly activating a profile in the web console switches the
running global provider until the next Android engine restart.
