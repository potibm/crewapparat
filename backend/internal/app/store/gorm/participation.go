package gorm

import (
	"github.com/google/uuid"

	"github.com/potibm/crewapparat/internal/app/domain"
)

type dbParticipation struct {
	TimestampModel
	AuditModel

	ID          uuid.UUID `gorm:"primaryKey;type:text;column:id"`
	OrganizerID uuid.UUID `gorm:"uniqueIndex:idx_participation_organizer_year;column:organizer_id"`
	EventYearID string    `gorm:"uniqueIndex:idx_participation_organizer_year;column:event_year_id"`
	Status      string    `gorm:"column:status"`

	// Merch and public info
	TShirtSize           string `gorm:"column:t_shirt_size"`
	AdditionalPaidShirt  bool   `gorm:"default:false;column:additional_paid_shirt"`
	PublicWebsiteConsent bool   `gorm:"default:false;column:public_website_consent"`

	ShiftAvailableTimes   []int    `gorm:"serializer:json;column:shift_available_times"`
	ShiftUnavailableTimes []int    `gorm:"serializer:json;column:shift_unavailable_times"`
	ShiftPreferredPartner []string `gorm:"serializer:json;column:shift_preferred_partner"`

	Teams []dbTeam `gorm:"many2many:participation_teams;joinForeignKey:participation_id;joinReferences:team_id"`
}

func (dbParticipation) TableName() string { return "participations" }

func (p dbParticipation) PrimaryKey() uuid.UUID { return p.ID }

func (p *dbParticipation) SetID(id uuid.UUID) { p.ID = id }

func fromDomainParticipation(p *domain.Participation) *dbParticipation {
	dbObj := &dbParticipation{
		ID:                    p.ID,
		OrganizerID:           p.OrganizerID,
		EventYearID:           p.EventYearID,
		Status:                string(p.Status),
		TShirtSize:            p.TShirtSize,
		AdditionalPaidShirt:   p.AdditionalPaidShirt,
		PublicWebsiteConsent:  p.PublicWebsiteConsent,
		ShiftAvailableTimes:   p.ShiftAvailableTimes,
		ShiftUnavailableTimes: p.ShiftUnavailableTimes,
		ShiftPreferredPartner: p.ShiftPreferredPartner,
	}

	for i := range p.Teams {
		dbObj.Teams = append(dbObj.Teams, *fromDomainTeam(&p.Teams[i]))
	}

	return dbObj
}

func toDomainParticipation(db *dbParticipation) *domain.Participation {
	domainObj := &domain.Participation{
		ID:                    db.ID,
		OrganizerID:           db.OrganizerID,
		EventYearID:           db.EventYearID,
		Status:                domain.ParticipationStatus(db.Status),
		TShirtSize:            db.TShirtSize,
		AdditionalPaidShirt:   db.AdditionalPaidShirt,
		PublicWebsiteConsent:  db.PublicWebsiteConsent,
		ShiftAvailableTimes:   db.ShiftAvailableTimes,
		ShiftUnavailableTimes: db.ShiftUnavailableTimes,
		ShiftPreferredPartner: db.ShiftPreferredPartner,
	}

	for i := range db.Teams {
		domainObj.Teams = append(domainObj.Teams, *toDomainTeam(&db.Teams[i]))
	}

	return domainObj
}
