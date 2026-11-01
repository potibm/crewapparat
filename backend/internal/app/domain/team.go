package domain

import "github.com/google/uuid"

// Team is an organizational group such as "Infodesk" or "Network".
type Team struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	// Color is a pastel hex color used for the team chip in the admin UI.
	Color string `json:"color,omitempty"`
	// HasShifts marks teams whose members need shift planning (e.g. Infodesk).
	HasShifts bool `json:"has_shifts"`
}
