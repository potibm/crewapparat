package gorm

import (
	"time"

	"github.com/google/uuid"

	"github.com/potibm/crewapparat/internal/app/domain"
)

type dbInvite struct {
	TimestampModel
	AuditModel

	ID     uuid.UUID `gorm:"primaryKey;type:text;column:id"`
	Email  string    `gorm:"uniqueIndex;column:email"`
	Status string    `gorm:"index;column:status"`

	// Provided by the inviting organizer when creating the invite.
	Handle    string `gorm:"column:handle"`
	DemozooID *uint  `gorm:"column:demozoo_id"`
	Reason    string `gorm:"column:reason"`

	// Magic-link credential, validated when the link is opened.
	Token     string    `gorm:"uniqueIndex;column:token"`
	ExpiresAt time.Time `gorm:"column:expires_at"`

	// Submitted through the public form by the invitee.
	Notes         string     `gorm:"column:notes"`
	CoCAccepted   bool       `gorm:"default:false;column:coc_accepted"`
	CoCAcceptedAt *time.Time `gorm:"column:coc_accepted_at"`

	InvitedByID  uuid.UUID  `gorm:"index;column:invited_by_id"`
	ReviewedByID *uuid.UUID `gorm:"column:reviewed_by_id"`

	// Filled in by the reviewing admin.
	ReviewNote string `gorm:"column:review_note"`

	// Tracking fields for onboarding steps.
	OpenIDAccountCreated bool `gorm:"default:false;column:openid_account_created"`
	SlackInvited         bool `gorm:"default:false;column:slack_invited"`
}

func (dbInvite) TableName() string { return "invites" }

func (i dbInvite) PrimaryKey() uuid.UUID { return i.ID }

func (i *dbInvite) SetID(id uuid.UUID) { i.ID = id }

func fromDomainInvite(inv *domain.Invite) *dbInvite {
	return &dbInvite{
		ID:                   inv.ID,
		Email:                inv.Email,
		Status:               string(inv.Status),
		Handle:               inv.Handle,
		DemozooID:            inv.DemozooID,
		Reason:               inv.Reason,
		Token:                inv.Token,
		ExpiresAt:            inv.ExpiresAt,
		Notes:                inv.Notes,
		CoCAccepted:          inv.CoCAccepted,
		CoCAcceptedAt:        inv.CoCAcceptedAt,
		InvitedByID:          inv.InvitedByID,
		ReviewedByID:         inv.ReviewedByID,
		ReviewNote:           inv.ReviewNote,
		OpenIDAccountCreated: inv.OpenIDAccountCreated,
		SlackInvited:         inv.SlackInvited,
	}
}

func toDomainInvite(db *dbInvite) *domain.Invite {
	return &domain.Invite{
		ID:                   db.ID,
		Email:                db.Email,
		Status:               domain.InviteStatus(db.Status),
		Handle:               db.Handle,
		DemozooID:            db.DemozooID,
		Reason:               db.Reason,
		Token:                db.Token,
		ExpiresAt:            db.ExpiresAt,
		Notes:                db.Notes,
		CoCAccepted:          db.CoCAccepted,
		CoCAcceptedAt:        db.CoCAcceptedAt,
		InvitedByID:          db.InvitedByID,
		ReviewedByID:         db.ReviewedByID,
		ReviewNote:           db.ReviewNote,
		OpenIDAccountCreated: db.OpenIDAccountCreated,
		SlackInvited:         db.SlackInvited,
	}
}
