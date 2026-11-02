package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_PlaylistDefaultsAndValidation(t *testing.T) {
	cfg := &Config{
		App: AppConfig{
			GinMode:     "debug",
			Environment: "development",
			LogLevel:    "info",
			LogFormat:   "text",
			DbFilename:  "test.db",
			FrontendURL: "http://localhost:3000",
			Port:        8080,
		},
		Sentry: SentryConfig{
			DSN:                     "https://test@sentry.io/123",
			TraceSampleRate:         0.1,
			ReplaySessionSampleRate: 0.1,
			ReplayErrorSampleRate:   0.1,
			Environment:             "development",
			Version:                 "1.2.3",
		},
		Auth: &AuthConfig{
			Type:         "oidc",
			Name:         "Dex",
			AuthorityURL: "https://dex.crewapparat.test",
			ClientID:     "react-admin-client",
		},
		Format: FormatConfig{
			Date: DateFormatConfig{
				Locale: "en-US",
				Options: DateFormatOptionsConfig{
					"weekday": "long",
					"hour":    "2-digit",
				},
			},
		},
	}

	// 1. trigger validation
	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestAppConfig_Validate(t *testing.T) {
	cfg := AppConfig{
		GinMode:     "debug",
		Environment: "development",
		LogLevel:    "info",
		LogFormat:   "text",
		DbFilename:  "test.db",
		FrontendURL: "http://localhost:3000",
		Port:        8080,
	}

	err := cfg.Validate()
	assert.NoError(t, err)

	cfg.DbFilename = "../invalid-filename"
	err = cfg.Validate()
	assert.Error(t, err)
}

func TestConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := &Config{
			App: AppConfig{
				GinMode:     "debug",
				Environment: "development",
				LogLevel:    "info",
				LogFormat:   "text",
				DbFilename:  "test.db",
				FrontendURL: "http://localhost:3000",
				Port:        8080,
			},
			Sentry: SentryConfig{
				Environment: "development",
				Version:     "1.0.0",
			},
			Format: FormatConfig{
				Date: DateFormatConfig{
					Locale:  "en-US",
					Options: DateFormatOptionsConfig{},
				},
			},
		}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("missing required fields", func(t *testing.T) {
		cfg := &Config{}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid configuration")
	})

	t.Run("invalid app config", func(t *testing.T) {
		cfg := &Config{
			App: AppConfig{
				GinMode:     "debug",
				Environment: "development",
				LogLevel:    "info",
				LogFormat:   "text",
				DbFilename:  "../invalid",
				FrontendURL: "http://localhost:3000",
				Port:        8080,
			},
			Sentry: SentryConfig{
				Environment: "development",
				Version:     "1.0.0",
			},
			Format: FormatConfig{
				Date: DateFormatConfig{
					Locale:  "en-US",
					Options: DateFormatOptionsConfig{},
				},
			},
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db_filename")
	})

	t.Run("invalid format config", func(t *testing.T) {
		cfg := &Config{
			App: AppConfig{
				GinMode:     "debug",
				Environment: "development",
				LogLevel:    "info",
				LogFormat:   "text",
				DbFilename:  "test.db",
				FrontendURL: "http://localhost:3000",
				Port:        8080,
			},
			Sentry: SentryConfig{
				Environment: "development",
				Version:     "1.0.0",
			},
			Format: FormatConfig{
				Date: DateFormatConfig{
					Locale:  "invalid-locale",
					Options: DateFormatOptionsConfig{},
				},
			},
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "date_locale")
	})
}

func TestAppConfig_Validate_Extended(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := AppConfig{
			GinMode:     "debug",
			Environment: "development",
			LogLevel:    "info",
			LogFormat:   "text",
			DbFilename:  "test.db",
			FrontendURL: "http://localhost:3000",
			Port:        8080,
		}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("invalid db filename with path traversal", func(t *testing.T) {
		cfg := AppConfig{
			GinMode:     "debug",
			Environment: "development",
			LogLevel:    "info",
			LogFormat:   "text",
			DbFilename:  "../etc/passwd",
			FrontendURL: "http://localhost:3000",
			Port:        8080,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db_filename")
	})

	t.Run("invalid db filename with special chars", func(t *testing.T) {
		cfg := AppConfig{
			GinMode:     "debug",
			Environment: "development",
			LogLevel:    "info",
			LogFormat:   "text",
			DbFilename:  "test@db",
			FrontendURL: "http://localhost:3000",
			Port:        8080,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db_filename")
	})

	t.Run("valid redis url", func(t *testing.T) {
		cfg := AppConfig{
			GinMode:     "debug",
			Environment: "development",
			LogLevel:    "info",
			LogFormat:   "text",
			DbFilename:  "test.db",
			FrontendURL: "http://localhost:3000",
			RedisURL:    "redis://localhost:6379/0",
			Port:        8080,
		}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("empty redis url is valid", func(t *testing.T) {
		cfg := AppConfig{
			GinMode:     "debug",
			Environment: "development",
			LogLevel:    "info",
			LogFormat:   "text",
			DbFilename:  "test.db",
			FrontendURL: "http://localhost:3000",
			RedisURL:    "",
			Port:        8080,
		}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("invalid redis url", func(t *testing.T) {
		cfg := AppConfig{
			GinMode:     "debug",
			Environment: "development",
			LogLevel:    "info",
			LogFormat:   "text",
			DbFilename:  "test.db",
			FrontendURL: "http://localhost:3000",
			RedisURL:    "://invalid",
			Port:        8080,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "redis_url")
	})
}

func TestFormatConfig_Validate(t *testing.T) {
	t.Run("valid locale", func(t *testing.T) {
		cfg := FormatConfig{
			Date: DateFormatConfig{
				Locale:  "da-DK",
				Options: DateFormatOptionsConfig{},
			},
		}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("invalid locale", func(t *testing.T) {
		cfg := FormatConfig{
			Date: DateFormatConfig{
				Locale:  "invalid",
				Options: DateFormatOptionsConfig{},
			},
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "date_locale")
	})

	t.Run("locale with wrong format", func(t *testing.T) {
		cfg := FormatConfig{
			Date: DateFormatConfig{
				Locale:  "en_US",
				Options: DateFormatOptionsConfig{},
			},
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "date_locale")
	})
}

func TestDateFormatConfig_Validate(t *testing.T) {
	t.Run("valid locale", func(t *testing.T) {
		cfg := DateFormatConfig{
			Locale:  "en-US",
			Options: DateFormatOptionsConfig{},
		}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("invalid locale", func(t *testing.T) {
		cfg := DateFormatConfig{
			Locale:  "foo",
			Options: DateFormatOptionsConfig{},
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "date_locale")
	})

	t.Run("locale too short", func(t *testing.T) {
		cfg := DateFormatConfig{
			Locale:  "en",
			Options: DateFormatOptionsConfig{},
		}
		err := cfg.Validate()
		assert.Error(t, err)
	})

	t.Run("locale too long", func(t *testing.T) {
		cfg := DateFormatConfig{
			Locale:  "eng-USA",
			Options: DateFormatOptionsConfig{},
		}
		err := cfg.Validate()
		assert.Error(t, err)
	})
}
