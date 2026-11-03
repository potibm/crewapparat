package gorm

import (
	"time"

	"github.com/google/uuid"

	"github.com/potibm/crewapparat/internal/app/domain"
)

type dbCodeOfConductAcceptance struct {
	TimestampModel
	AuditModel

	ID          uuid.UUID `gorm:"primaryKey;type:text;column:id"`
	OrganizerID uuid.UUID `gorm:"column:organizer_id;uniqueIndex:idx_coc_acceptance_organizer_version"`
	// CodeOfConductID references an immutable published version, so the
	// accepted text can never change underneath the acceptance.
	CodeOfConductID uuid.UUID `gorm:"column:code_of_conduct_id;uniqueIndex:idx_coc_acceptance_organizer_version"`
	AcceptedAt      time.Time `gorm:"column:accepted_at"`
}

func (dbCodeOfConductAcceptance) TableName() string { return "code_of_conduct_acceptances" }

func (a dbCodeOfConductAcceptance) PrimaryKey() uuid.UUID { return a.ID }

func (a *dbCodeOfConductAcceptance) SetID(id uuid.UUID) { a.ID = id }

func fromDomainCodeOfConductAcceptance(a *domain.CodeOfConductAcceptance) *dbCodeOfConductAcceptance {
	return &dbCodeOfConductAcceptance{
		ID:              a.ID,
		OrganizerID:     a.OrganizerID,
		CodeOfConductID: a.CodeOfConductID,
		AcceptedAt:      a.AcceptedAt,
	}
}

func toDomainCodeOfConductAcceptance(a *dbCodeOfConductAcceptance) *domain.CodeOfConductAcceptance {
	return &domain.CodeOfConductAcceptance{
		ID:              a.ID,
		OrganizerID:     a.OrganizerID,
		CodeOfConductID: a.CodeOfConductID,
		AcceptedAt:      a.AcceptedAt,
	}
}
