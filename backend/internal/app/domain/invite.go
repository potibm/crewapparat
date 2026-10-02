package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

type InviteStatus string

const (
	InviteStatusSent          InviteStatus = "sent"
	InviteStatusPendingReview InviteStatus = "pending_review"
	InviteStatusApproved      InviteStatus = "approved"
	InviteStatusRejected      InviteStatus = "rejected"
)

// InviteTokenHashLength ist die Hex-Länge von SHA-256.
const InviteTokenHashLength = 64

// Invite ist eine Einladung an eine Person, die es im aktiven Crew-Verzeichnis
// noch nicht gibt.
type Invite struct {
	ID     uuid.UUID    `json:"id"`
	Email  string       `json:"email"`
	Status InviteStatus `json:"status"`

	// TokenHash ist der SHA-256-Hash des Magic-Link-Tokens. Der Klartext wird
	// ausschließlich per E-Mail verschickt und nie gespeichert, sonst ergeben
	// Datenbank-Dumps oder Backups gültige Einladungslinks.
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`

	// Vom Eingeladenen im öffentlichen Formular angegeben.
	ProvidedName  string     `json:"provided_name,omitempty"`
	ProvidedPhone string     `json:"provided_phone,omitempty"`
	Notes         string     `json:"notes,omitempty"`
	CoCAccepted   bool       `json:"coc_accepted"`
	CoCAcceptedAt *time.Time `json:"coc_accepted_at,omitempty"`

	InvitedByID  uuid.UUID  `json:"invited_by_id"`
	ReviewedByID *uuid.UUID `json:"reviewed_by_id,omitempty"`

	// Vom Eingeladenen gewünschte Teams. Beim Approve werden sie in die
	// Participation des aktiven Jahres übernommen.
	Teams []Team `json:"teams,omitempty"`
}

// HashInviteToken bildet den Magic-Link-Token auf seine Speicherform ab.
func HashInviteToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))

	return hex.EncodeToString(sum[:])
}

// NewInviteToken erzeugt einen frischen Token und den Hash, der gespeichert
// wird. Nur `raw` gehört in die E-Mail, `hash` in die Datenbank.
func NewInviteToken() (raw, hash string) {
	raw = uuid.NewString()

	return raw, HashInviteToken(raw)
}

// IsExpired meldet, ob der Einladungslink abgelaufen ist.
func (i Invite) IsExpired(now time.Time) bool {
	return !i.ExpiresAt.IsZero() && i.ExpiresAt.Before(now)
}
