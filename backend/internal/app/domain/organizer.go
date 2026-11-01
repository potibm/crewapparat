package domain

import "github.com/google/uuid"

// Organizer is the master record of a crew member and spans all years.
// Per-year details such as t-shirt size or shift times live in the
// Participation of that year.
type Organizer struct {
	ID              uuid.UUID `json:"id"`
	OidcSub         *string   `json:"oidc_sub,omitempty"` // set on first login against the IdP
	Handle          string    `json:"handle"`             // Demoscene-Handle
	Slug            string    `json:"slug"`               // immutable, slugified handle, unique
	Realname        string    `json:"realname"`           // legal name
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

	// CurrentParticipationStatus is the status of the participation in the
	// requested event year (only populated when filtered by event_year_id).
	CurrentParticipationStatus string `json:"current_participation_status,omitempty"`

	// CurrentParticipationTeams are the teams of the participation in the
	// requested event year (only populated when filtered by event_year_id).
	CurrentParticipationTeams []Team `json:"current_participation_teams,omitempty"`
}
