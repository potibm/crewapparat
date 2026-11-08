package domain

import (
	"time"

	"github.com/google/uuid"
)

type ShiftPlanStatus string

const (
	ShiftPlanStatusPending   ShiftPlanStatus = "pending"
	ShiftPlanStatusPublished ShiftPlanStatus = "published"
)

// ShiftPlanSlot is one half-hour block of the shift grid. MainOrganizerID is the
// main organizer on duty during that slot (if any); Organizers is the Infodesk
// crew assigned to it.
type ShiftPlanSlot struct {
	Slot            int      `json:"slot"`
	MainOrganizerID string   `json:"main_organizer_id,omitempty"`
	Organizers      []string `json:"organizers"`
}

// ShiftPlan is a versioned shift schedule for one event year. The slots are
// stored as a JSON array; the plan is read-only from the admin perspective.
type ShiftPlan struct {
	ID          uuid.UUID       `json:"id"`
	EventYearID string          `json:"event_year_id"`
	Version     string          `json:"version"`
	Status      ShiftPlanStatus `json:"status"`
	GeneratedAt *time.Time      `json:"generated_at,omitempty"`
	Slots       []ShiftPlanSlot `json:"slots"`
}
