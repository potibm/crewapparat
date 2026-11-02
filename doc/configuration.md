# Crewapparat

## Configuration

The application can be configured in multiple ways. This includes CLI parameters, environment variables and config files.

They are processed in the following order:

- CLI parameters
- Environment variables
- Config file
- Defaults.
  
### CLI Parameters

You can call the `crewapparat serve` command with the following flags

```
-p, --port          int             # the port for the server
--otel-endpoint     string          # hostname and port for otel
--db-file           string          # the database file to use (without suffix)
--log-format        string          # log format, any of json, text
--log-level         string          # log level, any of debug, info, warn, error
```

For details call `crewapparat help serve`.

### Config

The application will use the config in `config/config.yaml` and `config/config.local.yaml` (which is ignored by git and useful for local overrides).

To generate a basic config you can use `crewapparat config create`.

The yaml consists of those parts:

#### app

The app section defines core server settings, database paths, and basic network policies.

````
app:
    cors_allow_origins:
        - https://localhost:3400            
    db_filename: crewapparat
    env: production 
    frontend_url: https://localhost:3400
    log_format: json
    log_level: info
    redis_url: "redis://localhost:6379/0"
````

#### format

Used to localize how dates and times are rendered in the frontend.

````
format:
    date:
        locale: da-DK           # the locale to use to display dates in the frontend
        options:                # javascript DateTimeFormat-Options, see below
            hour: 2-digit
            minute: 2-digit
            weekday: short
````

The date options are according to the documentation at https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Intl/DateTimeFormat/DateTimeFormat#options.

#### sentry

When using Sentry, provide the project specific DSN and further information here:

`````
sentry:
    dsn: ""
    environment: production
    replay_error_sample_rate: 0.1
    replay_session_sample_rate: 0.1
    trace_sample_rate: 0.1
    version: dev
`````

#### auth

Enables OIDC for the admin panel. Omit the whole section to disable authentication (see
[auth.md](auth.md)). ````
auth:
    type: oidc
    name: Dex
    authority: https://dex.crewapparat.test/dex
    client_id: react-admin-client
    skip_tls_verify: false
````

## Environment

You can overwrite all the config properties using environment variables.

`app.env` translates to `APP_ENV`,  `app.redis_url` to `APP_REDIS_URL` and so on.

`APP_CORS_ALLOW_ORIGINS` can be provided as a comma separated list of values.

For more complex types (the `auth` section, for example) better use the config file mentioned above.