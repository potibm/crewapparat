package hub

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/potibm/crewapparat/internal/app/config"
)

type AppConfigPublic struct {
	Version            string                         `json:"version"`
	Environment        string                         `json:"environment"`
	EnvironmentMessage string                         `json:"environment_message"`
	DateLocale         string                         `json:"date_locale"`
	DateOptions        config.DateFormatOptionsConfig `json:"date_options"`
	Sentry             config.SentryConfig            `json:"sentry"`
	Auth               *config.AuthConfig             `json:"auth,omitempty"`
	PronounOptions     []string                       `json:"pronoun_options"`
	LanguageOptions    []string                       `json:"language_options"`
	TShirtSizeOptions  []string                       `json:"tshirt_size_options"`
	TShirtDeadline     string                         `json:"tshirt_deadline"`
	Shift              config.ShiftConfig             `json:"shift"`
}

func (s *Server) handleGetPublicConfig(c *gin.Context) {
	pub := mapToPublicConfig(&s.cfg)

	c.JSON(http.StatusOK, pub)
}

func mapToPublicConfig(cfg *config.Config) AppConfigPublic {
	return AppConfigPublic{
		Version:            cfg.App.Version,
		Environment:        cfg.App.Environment,
		EnvironmentMessage: cfg.App.EnvironmentMessage,
		Sentry:             cfg.Sentry,
		DateLocale:         cfg.Format.Date.Locale,
		DateOptions:        cfg.Format.Date.Options,
		Auth:               cfg.Auth,
		PronounOptions:     cfg.App.PronounOptions,
		LanguageOptions:    cfg.App.LanguageOptions,
		TShirtSizeOptions:  cfg.App.TShirtSizeOptions,
		TShirtDeadline:     cfg.TShirtDeadline,
		Shift:              cfg.Shift,
	}
}
