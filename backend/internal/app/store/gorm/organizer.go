package gorm

import (
	"github.com/google/uuid"

	"github.com/potibm/crewapparat/internal/app/domain"
)

type dbOrganizer struct {
	TimestampModel
	AuditModel

	ID                   uuid.UUID `gorm:"primaryKey;type:text;column:id"`
	OidcSub              *string   `gorm:"uniqueIndex;column:oidc_sub"`
	Handle               string    `gorm:"index;column:handle"`
	Slug                 string    `gorm:"column:slug"`
	Realname             string    `gorm:"column:realname"`
	Email                string    `gorm:"uniqueIndex;column:email"`
	Phone                string    `gorm:"column:phone"`
	SlackHandle          string    `gorm:"column:slack_handle"`
	SlackProfileURL      string    `gorm:"column:slack_profile_url"`
	MediaWikiProfileURL  string    `gorm:"column:mediawiki_profile_url"`
	AvatarURL            string    `gorm:"column:avatar_url"`
	DemozooID            *uint     `gorm:"column:demozoo_id"`
	FirstYearAsOrganizer *int      `gorm:"column:first_year_as_organizer"`
	Pronouns             string    `gorm:"column:pronouns"`

	Languages []string `gorm:"serializer:json;column:languages"`

	Participations []dbParticipation `gorm:"foreignKey:OrganizerID;references:ID"`
}

func (dbOrganizer) TableName() string { return "organizers" }

func (o dbOrganizer) PrimaryKey() uuid.UUID { return o.ID }

func (o *dbOrganizer) SetID(id uuid.UUID) { o.ID = id }

func fromDomainOrganizer(o *domain.Organizer) *dbOrganizer {
	dbObj := &dbOrganizer{
		ID:                   o.ID,
		OidcSub:              o.OidcSub,
		Handle:               o.Handle,
		Slug:                 o.Slug,
		Realname:             o.Realname,
		Email:                o.Email,
		Phone:                o.Phone,
		SlackHandle:          o.SlackHandle,
		SlackProfileURL:      o.SlackProfileURL,
		MediaWikiProfileURL:  o.MediaWikiProfileURL,
		AvatarURL:            o.AvatarURL,
		DemozooID:            o.DemozooID,
		FirstYearAsOrganizer: o.FirstYearAsOrganizer,
		Pronouns:             o.Pronouns,
		Languages:            o.Languages,
	}

	for i := range o.Participations {
		dbObj.Participations = append(dbObj.Participations, *fromDomainParticipation(&o.Participations[i]))
	}

	return dbObj
}

func toDomainOrganizer(db *dbOrganizer) *domain.Organizer {
	domainObj := &domain.Organizer{
		ID:                   db.ID,
		OidcSub:              db.OidcSub,
		Handle:               db.Handle,
		Slug:                 db.Slug,
		Realname:             db.Realname,
		Email:                db.Email,
		Phone:                db.Phone,
		SlackHandle:          db.SlackHandle,
		SlackProfileURL:      db.SlackProfileURL,
		MediaWikiProfileURL:  db.MediaWikiProfileURL,
		AvatarURL:            db.AvatarURL,
		DemozooID:            db.DemozooID,
		FirstYearAsOrganizer: db.FirstYearAsOrganizer,
		Pronouns:             db.Pronouns,
		Languages:            db.Languages,
	}

	for i := range db.Participations {
		domainObj.Participations = append(domainObj.Participations, *toDomainParticipation(&db.Participations[i]))
	}

	return domainObj
}
