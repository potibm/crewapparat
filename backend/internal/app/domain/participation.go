package domain

import "github.com/google/uuid"

type ParticipationStatus string

const (
	ParticipationStatusPending   ParticipationStatus = "pending"
	ParticipationStatusConfirmed ParticipationStatus = "confirmed"
	ParticipationStatusDeclined  ParticipationStatus = "declined"
)

// Participation bundles everything an organizer stated for one event year.
type Participation struct {
	ID          uuid.UUID           `json:"id"`
	OrganizerID uuid.UUID           `json:"organizer_id"`
	EventYearID string              `json:"event_year_id"`
	Status      ParticipationStatus `json:"status"`

	// Merch and public info
	TShirtSize           string `json:"t_shirt_size,omitempty"`
	AdditionalPaidShirt  bool   `json:"additional_paid_shirt"`
	PublicWebsiteConsent bool   `json:"public_website_consent"`

	// Shift planning: indices of half-hour blocks.
	ShiftAvailableTimes   []int    `json:"shift_available_times"`
	ShiftUnavailableTimes []int    `json:"shift_unavailable_times"`
	ShiftPreferredPartner []string `json:"shift_preferred_partner"` // UUIDs der Wunschpartner

	// Teams the organizer joined in this year.
	Teams []Team `json:"teams,omitempty"`
}
