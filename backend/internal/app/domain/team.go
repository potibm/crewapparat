package domain

import "github.com/google/uuid"

// Team ist eine organisatorische Gruppe wie "Infodesk" oder "Network".
type Team struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}
