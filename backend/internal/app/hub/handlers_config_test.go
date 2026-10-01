package hub

import (
	"encoding/json"
	"testing"

	"github.com/potibm/crewapparat/internal/app/config"
	"github.com/stretchr/testify/assert"
)

func TestMapToPublicConfig(t *testing.T) {
	internalCfg := &config.Config{
		App: config.AppConfig{
			Version:            "1.2.3",
			Environment:        "production",
			EnvironmentMessage: "Hello World",
			RedisURL:           "redis://:supersecret@localhost:6379/0",
		},
		Sentry: config.SentryConfig{
			DSN:         "https://secret@sentry.io/123",
			Environment: "prod",
			Version:     "v1",
		},
	}

	public := mapToPublicConfig(internalCfg)

	// Verification
	assert.Equal(t, "1.2.3", public.Version)
	assert.Equal(t, "production", public.Environment)
	assert.Equal(t, "Hello World", public.EnvironmentMessage)
	assert.Equal(t, "https://secret@sentry.io/123", public.Sentry.DSN)

	// Secrets dürfen die öffentliche Config nie verlassen.
	payload, err := json.Marshal(public)
	assert.NoError(t, err)
	assert.NotContains(t, string(payload), "supersecret")
	assert.NotContains(t, string(payload), "redis_url")
}
