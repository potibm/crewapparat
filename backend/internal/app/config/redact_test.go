package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_RedactConfigForDisplay(t *testing.T) {
	t.Run("redacts sensitive fields", func(t *testing.T) {
		original := Config{
			App: AppConfig{
				RedisURL: "redis://:secret@localhost:6379/0",
			},
			Sentry: SentryConfig{
				DSN: "https://key@sentry.io/123",
			},
		}

		result := original.RedactConfigForDisplay()

		assert.Equal(t, "***REDACTED***", result.Sentry.DSN)
		assert.Equal(t, RedisURL("redis://:%2A%2A%2AREDACTED%2A%2A%2A@localhost:6379/0"), result.App.RedisURL)

		// Verify original is not modified
		assert.Equal(t, "https://key@sentry.io/123", original.Sentry.DSN)
		assert.Equal(t, RedisURL("redis://:secret@localhost:6379/0"), original.App.RedisURL)
	})

	t.Run("redacts redis url without password", func(t *testing.T) {
		original := Config{
			App: AppConfig{
				RedisURL: "redis://localhost:6379/0",
			},
		}

		result := original.RedactConfigForDisplay()

		assert.Equal(t, RedisURL("redis://localhost:6379/0"), result.App.RedisURL)
	})

	t.Run("handles empty redis url", func(t *testing.T) {
		original := Config{
			App: AppConfig{
				RedisURL: "",
			},
		}

		result := original.RedactConfigForDisplay()

		assert.Equal(t, RedisURL(""), result.App.RedisURL)
	})
}

func TestRedactURLPassword(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "URL with password",
			input:    "redis://:secret@localhost:6379/0",
			expected: "redis://:%2A%2A%2AREDACTED%2A%2A%2A@localhost:6379/0",
		},
		{
			name:     "URL with username and password",
			input:    "redis://user:secret@localhost:6379/0",
			expected: "redis://user:%2A%2A%2AREDACTED%2A%2A%2A@localhost:6379/0",
		},
		{
			name:     "URL without credentials",
			input:    "redis://localhost:6379/0",
			expected: "redis://localhost:6379/0",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "invalid URL",
			input:    "://invalid",
			expected: "://invalid",
		},
		{
			name:     "http URL with password",
			input:    "http://user:pass@example.com/path",
			expected: "http://user:%2A%2A%2AREDACTED%2A%2A%2A@example.com/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := redactURLPassword(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
