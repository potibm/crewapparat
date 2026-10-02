package domain

import "github.com/google/uuid"

type ParticipationStatus string

const (
	ParticipationStatusPending   ParticipationStatus = "pending"
	ParticipationStatusConfirmed ParticipationStatus = "confirmed"
	ParticipationStatusDeclined  ParticipationStatus = "declined"
)

// Participation bündelt alle jahresbezogenen Angaben eines Organizers.
type Participation struct {
	ID          uuid.UUID           `json:"id"`
	OrganizerID uuid.UUID           `json:"organizer_id"`
	EventYearID string              `json:"event_year_id"`
	Status      ParticipationStatus `json:"status"`

	// Merch & öffentliche Info
	TShirtSize           string `json:"t_shirt_size,omitempty"`
	AdditionalPaidShirt  bool   `json:"additional_paid_shirt"`
	PublicWebsiteConsent bool   `json:"public_website_consent"`

	// Schichtplanung: Indizes von Halbstundenblöcken.
	ShiftAvailableTimes   []int    `json:"shift_available_times"`
	ShiftUnavailableTimes []int    `json:"shift_unavailable_times"`
	ShiftPreferredPartner []string `json:"shift_preferred_partner"` // UUIDs der Wunschpartner

	// Teams, denen der Organizer in diesem Jahr beigetreten ist.
	Teams []Team `json:"teams,omitempty"`
}
