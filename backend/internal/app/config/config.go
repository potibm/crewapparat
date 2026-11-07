package config

import (
	"strings"

	"github.com/spf13/viper"
)

const (
	OtelServiceName        = "crewapparat"
	OtelBackendServiceName = OtelServiceName + "-backend"

	DefaultPort = 8080

	DefaultTraceSampleRate         = 0.1
	DefaultReplaySessionSampleRate = 0.1
	DefaultReplayErrorSampleRate   = 0.1

	DataDirname = "./data"

	DefaultDBFilename = "crewapparat"

	DefaultShiftStart      = "2026-08-07T16:00:00+02:00"
	DefaultShiftBlockCount = 90

	DataDirPerm = 0o755
)

var DefaultDateOptions = DateFormatOptionsConfig{
	"weekday": "short",
	"hour":    "2-digit",
	"minute":  "2-digit",
}

func InitViper() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("app.port", DefaultPort)
	viper.SetDefault("app.otel_endpoint", "")
	viper.SetDefault("app.gin_mode", "release")
	viper.SetDefault("app.log_level", "info")
	viper.SetDefault("app.env", "production")
	viper.SetDefault("app.db_filename", DefaultDBFilename)
	viper.SetDefault("app.frontend_url", "")
	viper.SetDefault("app.event_name", "Evoke")
	viper.SetDefault("app.cors_allow_origins", []string{})
	viper.SetDefault("app.redis_url", "")

	viper.SetDefault("mailer.dsn", "smtp://localhost:1026")
	viper.SetDefault("mailer.from", "crewapparat@example.com")
	viper.SetDefault("mailer.subject_prefix", "[Crewapparat]")
	viper.SetDefault("mailer.admin_recipients", []string{})
	viper.SetDefault("app.pronoun_options", []string{"she/her", "he/him", "they/them"})
	viper.SetDefault("app.language_options", []string{"de", "en", "fr", "hu", "no", "nl", "ru", "pl"})
	viper.SetDefault("app.tshirt_size_options", []string{"XXS", "XS", "S", "M", "L", "XL", "XXL", "3XL"})

	viper.SetDefault("format.date.locale", "da-DK")
	viper.SetDefault("format.date.options", DefaultDateOptions)

	viper.SetDefault("shift.start", DefaultShiftStart)
	viper.SetDefault("shift.block_count", DefaultShiftBlockCount)

	viper.SetDefault("sentry.dsn", "")
	viper.SetDefault("sentry.trace_sample_rate", DefaultTraceSampleRate)

	viper.SetDefault("auth.oidc_admin_group", "crewapparat-admins")
	viper.SetDefault("sentry.replay_session_sample_rate", DefaultReplaySessionSampleRate)
	viper.SetDefault("sentry.replay_error_sample_rate", DefaultReplayErrorSampleRate)

	viper.RegisterAlias("mailer.frontend_url", "app.frontend_url")
	viper.RegisterAlias("sentry.environment", "app.env")
	viper.RegisterAlias("sentry.version", "app.version")
}
