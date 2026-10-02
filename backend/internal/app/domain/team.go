package domain

import "github.com/google/uuid"

// Team is an organizational group such as "Infodesk" or "Network".
type Team struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}
