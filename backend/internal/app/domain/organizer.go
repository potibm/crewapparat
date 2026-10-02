package domain

import "github.com/google/uuid"

// Organizer ist der Stammdatensatz eines Crew-Mitglieds und gilt über alle
// Jahre hinweg. Jährliche Angaben wie T-Shirt-Größe oder Schichtzeiten stehen
// in der Participation des jeweiligen Jahres.
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

	// Languages wird als JSON-Array in SQLite gespeichert. API: ["de", "en"]
	Languages []string `json:"languages"`

	Participations []Participation `json:"participations,omitempty"`
}
