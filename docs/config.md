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

## Sections

### `app`

Core server settings, database paths, and basic network policies.

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `version` | string | yes | `dev` | Application version. |
| `gin_mode` | string | yes | `debug` | Gin mode: `debug`, `release`, `test`. |
| `env` | string | yes | `development` | Environment: `development`, `staging`, `production`, `test`. |
| `log_level` | string | yes | `debug` | Log level: `debug`, `info`, `warn`, `error`. |
| `log_format` | string | yes | `json` | Log format: `json`, `text`. |
| `db_filename` | string | yes | `crewapparat` | SQLite database filename (without suffix). |
| `frontend_url` | string (URL) | yes | `https://crewapparat.test` | Public frontend URL. |
| `event_name` | string | no | – | Event name shown in invites/emails. |
| `cors_allow_origins` | list of strings | yes | `https://crewapparat.test` | Allowed CORS origins. |
| `environment_message` | string | no | – | Banner message for non-production environments. |
| `redis_url` | string (URL) | no | – | Redis URL for caching/sessions. |
| `otel_endpoint` | string | no | – | OpenTelemetry collector endpoint. |
| `pronoun_options` | list of strings | no | – | Available pronoun options for organizers. |
| `language_options` | list of strings | no | – | Available language options. |
| `tshirt_size_options` | list of strings | no | – | Available T-shirt sizes. |
| `port` | int | yes | `3401` | HTTP listen port (1–65535). |

### `auth`

OIDC authentication. Omit entirely for header-based dev auth.

| Field | Type | Required | Description |
|---|---|---|---|
| `type` | string | yes | Must be `oidc`. |
| `name` | string | yes | Display name of the provider. |
| `authority` | string (URL) | yes | OIDC authority URL (e.g. Dex). |
| `client_id` | string | yes | OIDC client ID. |
| `skip_tls_verify` | bool | no | Disable TLS verification (dev only). |
| `oidc_admin_group` | string | no | Group name granting admin rights. |
| `proxy_admins` | list of strings | no | User subs treated as admins (dev). |
| `header_user_sub` | string | no | Header name for the user sub (proxy auth). |
| `header_user_groups` | string | no | Header name for the user groups (proxy auth). |

### `mailer`

SMTP mailer used for invite emails.

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `dsn` | string | yes | – | SMTP DSN, e.g. `smtp://user:pass@host:587`. |
| `from` | string (email) | no | `crewapparat@example.com` | Sender address. |
| `subject_prefix` | string | yes | `[Crewapparat] ` | Subject prefix. |
| `frontend_url` | string (URL) | yes | `http://localhost:3000` | Base URL for magic links. |
| `admin_recipients` | list of strings (email) | no | – | Recipients for pending-approval notifications. |

### `sentry`

Error tracking.

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `dsn` | string (URL) | no | – | Sentry DSN. |
| `environment` | string | yes | `development` | Environment name. |
| `version` | string | yes | `dev` | Release version. |
| `trace_sample_rate` | float (0–1) | no | `0.1` | Tracing sample rate. |
| `replay_session_sample_rate` | float (0–1) | no | `0.1` | Session replay sample rate. |
| `replay_error_sample_rate` | float (0–1) | no | `0.1` | Error replay sample rate. |

### `format`

Frontend date/time formatting.

| Field | Type | Required | Description |
|---|---|---|---|
| `date.locale` | string | yes | Locale, e.g. `en-GB`. |
| `date.options` | map | no | `Intl.DateTimeFormat` options. |

### `shift`

Shift grid for participation planning.

| Field | Type | Required | Description |
|---|---|---|---|
| `start` | string (ISO 8601) | no | Timestamp of block 0. |
| `block_count` | int | no | Total number of half-hour blocks. |

## Example

See `backend/config/config.yaml` for a complete, committed default.
```