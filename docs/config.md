# Configuration

## Loading order

1. `backend/config/config.yaml` (committed defaults)
2. `backend/config/config.local.yaml` (gitignored overrides)
3. `.env` (loaded by godotenv)
4. Environment variables (`APP_LOG_LEVEL` maps to `app.log_level`)
5. CLI flags (`--log-level`, `--port`, …)

Use `config/config.local.yaml` for local secrets; do not edit `config.yaml`.
Generate a fresh config with `go run . config create`.

## Secrets

Secrets (SMTP credentials, OIDC client secret, …) belong in
`config.local.yaml` or environment variables, never in `config.yaml`.
```