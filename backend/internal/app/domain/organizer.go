package domain

import "github.com/google/uuid"

// Organizer is the master record of a crew member and spans all years.
// Per-year details such as t-shirt size or shift times live in the
// Participation of that year.
type Organizer struct {
	ID              uuid.UUID `json:"id"`
	OidcSub         *string   `json:"oidc_sub,omitempty"` // wird beim ersten Login gegen Rauthy gesetzt
	Handle          string    `json:"handle"`             // Demoscene-Handle
	Realname        string    `json:"realname"`           // bürgerlicher Name
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	SlackHandle     string    `json:"slack_handle,omitempty"`
	SlackProfileURL string    `json:"slack_profile_url,omitempty"`
	AvatarURL       string    `json:"avatar_url,omitempty"`
	DemozooID       *uint     `json:"demozoo_id,omitempty"`
	Pronouns        string    `json:"pronouns,omitempty"`

	// Languages is stored as a JSON array in SQLite. API: ["de", "en"]
	Languages []string `json:"languages"`

	Participations []Participation `json:"participations,omitempty"`
}
