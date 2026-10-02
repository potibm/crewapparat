package gorm

import (
	"github.com/google/uuid"

	"github.com/potibm/crewapparat/internal/app/domain"
)

type dbOrganizer struct {
	TimestampModel
	AuditModel

	ID              uuid.UUID `gorm:"primaryKey;type:text;column:id"`
	OidcSub         *string   `gorm:"uniqueIndex;column:oidc_sub"`
	Handle          string    `gorm:"index;column:handle"`
	Realname        string    `gorm:"column:realname"`
	Email           string    `gorm:"uniqueIndex;column:email"`
	Phone           string    `gorm:"column:phone"`
	SlackHandle     string    `gorm:"column:slack_handle"`
	SlackProfileURL string    `gorm:"column:slack_profile_url"`
	AvatarURL       string    `gorm:"column:avatar_url"`
	DemozooID       *uint     `gorm:"column:demozoo_id"`
	Pronouns        string    `gorm:"column:pronouns"`

	Languages []string `gorm:"serializer:json;column:languages"`

	Participations []dbParticipation `gorm:"foreignKey:OrganizerID;references:ID"`
}

func (dbOrganizer) TableName() string { return "organizers" }

func (o dbOrganizer) PrimaryKey() uuid.UUID { return o.ID }

func (o *dbOrganizer) SetID(id uuid.UUID) { o.ID = id }

func fromDomainOrganizer(o *domain.Organizer) *dbOrganizer {
	dbObj := &dbOrganizer{
		ID:              o.ID,
		OidcSub:         o.OidcSub,
		Handle:          o.Handle,
		Realname:        o.Realname,
		Email:           o.Email,
		Phone:           o.Phone,
		SlackHandle:     o.SlackHandle,
		SlackProfileURL: o.SlackProfileURL,
		AvatarURL:       o.AvatarURL,
		DemozooID:       o.DemozooID,
		Pronouns:        o.Pronouns,
		Languages:       o.Languages,
	}

	for i := range o.Participations {
		dbObj.Participations = append(dbObj.Participations, *fromDomainParticipation(&o.Participations[i]))
	}

	return dbObj
}

func toDomainOrganizer(db *dbOrganizer) *domain.Organizer {
	domainObj := &domain.Organizer{
		ID:              db.ID,
		OidcSub:         db.OidcSub,
		Handle:          db.Handle,
		Realname:        db.Realname,
		Email:           db.Email,
		Phone:           db.Phone,
		SlackHandle:     db.SlackHandle,
		SlackProfileURL: db.SlackProfileURL,
		AvatarURL:       db.AvatarURL,
		DemozooID:       db.DemozooID,
		Pronouns:        db.Pronouns,
		Languages:       db.Languages,
	}

	for i := range db.Participations {
		domainObj.Participations = append(domainObj.Participations, *toDomainParticipation(&db.Participations[i]))
	}

	return domainObj
}
