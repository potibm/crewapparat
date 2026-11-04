package domain

import (
	"time"

	"github.com/google/uuid"
)

type InviteStatus string

const (
	InviteStatusSent            InviteStatus = "sent"
	InviteStatusPendingApproval InviteStatus = "pending_approval"
	InviteStatusPendingReview   InviteStatus = "pending_review"
	InviteStatusApproved        InviteStatus = "approved"
	InviteStatusRejected        InviteStatus = "rejected"
)

// Invite is an invitation for someone who is not in the crew directory yet.
//
// Team assignment does not happen here: the reviewing organizer picks the teams
// on approval, and returning crew selects theirs in the annual check-in.
type Invite struct {
	ID     uuid.UUID    `json:"id"`
	Email  string       `json:"email"`
	Status InviteStatus `json:"status"`

	// Provided by the inviting organizer when creating the invite.
	Handle    string `json:"handle,omitempty"`
	DemozooID *uint  `json:"demozoo_id,omitempty"`
	Reason    string `json:"reason,omitempty"`

	// Filled in by the reviewing admin.
	ReviewNote string `json:"review_note,omitempty"`

	// Token is the magic-link credential. It travels in the invitation URL and
	// is validated when the link is opened.
	Token     string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`

	// Submitted through the public form by the invitee.
	ProvidedName  string     `json:"provided_name,omitempty"`
	ProvidedPhone string     `json:"provided_phone,omitempty"`
	Notes         string     `json:"notes,omitempty"`
	CoCAccepted   bool       `json:"coc_accepted"`
	CoCAcceptedAt *time.Time `json:"coc_accepted_at,omitempty"`

	InvitedByID  uuid.UUID  `json:"invited_by_id"`
	ReviewedByID *uuid.UUID `json:"reviewed_by_id,omitempty"`
}

// NewInviteToken returns a fresh magic-link token.
func NewInviteToken() string {
	return uuid.NewString()
}

// IsExpired reports whether the invitation link has expired.
func (i Invite) IsExpired(now time.Time) bool {
	return !i.ExpiresAt.IsZero() && i.ExpiresAt.Before(now)
}
