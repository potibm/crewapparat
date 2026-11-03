package domain

import (
	"time"

	"github.com/google/uuid"
)

type CodeOfConductStatus string

const (
	CodeOfConductStatusPending   CodeOfConductStatus = "pending"
	CodeOfConductStatusPublished CodeOfConductStatus = "published"
)

// CodeOfConduct is a versioned legal document (e.g. the Code of Conduct).
// Multiple versions may exist; the latest published one is considered active.
type CodeOfConduct struct {
	ID          uuid.UUID           `json:"id"`
	Version     string              `json:"version"` // semantic version, e.g. "1.2.0"
	Status      CodeOfConductStatus `json:"status"`
	ChangeNote  string              `json:"change_note"` // free-text summary of what changed in this version
	Text        string              `json:"text"`        // Markdown
	PublishedAt *time.Time          `json:"published_at,omitempty"`

	// AcceptedCount is the number of organizers who accepted this version.
	// It is populated by the repository and never persisted.
	AcceptedCount int64 `json:"accepted_count"`
}
