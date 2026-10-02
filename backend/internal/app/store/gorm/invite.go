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

	// Nur der SHA-256-Hash des Magic-Link-Tokens, nie der Klartext.
	TokenHash string    `gorm:"uniqueIndex;column:token_hash"`
	ExpiresAt time.Time `gorm:"column:expires_at"`

	// Vom Eingeladenen im öffentlichen Formular angegeben.
	ProvidedName  string     `gorm:"column:provided_name"`
	ProvidedPhone string     `gorm:"column:provided_phone"`
	Notes         string     `gorm:"column:notes"`
	CoCAccepted   bool       `gorm:"default:false;column:coc_accepted"`
	CoCAcceptedAt *time.Time `gorm:"column:coc_accepted_at"`

	InvitedByID  uuid.UUID  `gorm:"index;column:invited_by_id"`
	ReviewedByID *uuid.UUID `gorm:"column:reviewed_by_id"`

	Teams []dbTeam `gorm:"many2many:invite_teams;joinForeignKey:invite_id;joinReferences:team_id"`
}

func (dbInvite) TableName() string { return "invites" }

func (i dbInvite) PrimaryKey() uuid.UUID { return i.ID }

func (i *dbInvite) SetID(id uuid.UUID) { i.ID = id }

func fromDomainInvite(inv *domain.Invite) *dbInvite {
	dbObj := &dbInvite{
		ID:            inv.ID,
		Email:         inv.Email,
		Status:        string(inv.Status),
		TokenHash:     inv.TokenHash,
		ExpiresAt:     inv.ExpiresAt,
		ProvidedName:  inv.ProvidedName,
		ProvidedPhone: inv.ProvidedPhone,
		Notes:         inv.Notes,
		CoCAccepted:   inv.CoCAccepted,
		CoCAcceptedAt: inv.CoCAcceptedAt,
		InvitedByID:   inv.InvitedByID,
		ReviewedByID:  inv.ReviewedByID,
	}

	for idx := range inv.Teams {
		dbObj.Teams = append(dbObj.Teams, *fromDomainTeam(&inv.Teams[idx]))
	}

	return dbObj
}

func toDomainInvite(db *dbInvite) *domain.Invite {
	domainObj := &domain.Invite{
		ID:            db.ID,
		Email:         db.Email,
		Status:        domain.InviteStatus(db.Status),
		TokenHash:     db.TokenHash,
		ExpiresAt:     db.ExpiresAt,
		ProvidedName:  db.ProvidedName,
		ProvidedPhone: db.ProvidedPhone,
		Notes:         db.Notes,
		CoCAccepted:   db.CoCAccepted,
		CoCAcceptedAt: db.CoCAcceptedAt,
		InvitedByID:   db.InvitedByID,
		ReviewedByID:  db.ReviewedByID,
	}

	for idx := range db.Teams {
		domainObj.Teams = append(domainObj.Teams, *toDomainTeam(&db.Teams[idx]))
	}

	return domainObj
}
