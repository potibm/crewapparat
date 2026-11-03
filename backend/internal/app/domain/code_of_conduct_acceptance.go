package domain

import (
	"time"

	"github.com/google/uuid"
)

// CodeOfConductAcceptance records that an organizer accepted a specific
// published version of the Code of Conduct. The referenced version is
// immutable, so the record always describes the exact text that was accepted.
//
// Deliberately no IP address or user agent: the timestamp and the version
// reference are sufficient evidence and keep the personal data minimal.
type CodeOfConductAcceptance struct {
	ID              uuid.UUID `json:"id"`
	OrganizerID     uuid.UUID `json:"organizer_id"`
	CodeOfConductID uuid.UUID `json:"code_of_conduct_id"`
	AcceptedAt      time.Time `json:"accepted_at"`
}
