package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/potibm/crewapparat/internal/app/domain"
)

type CodeOfConductAcceptanceListFilters struct {
	OrganizerID     *uuid.UUID
	CodeOfConductID *uuid.UUID
}

type CodeOfConductAcceptanceListParams = ListParams

// CodeOfConductAcceptanceWithOrganizer enriches an acceptance with the
// organizer data the admin overview needs.
type CodeOfConductAcceptanceWithOrganizer struct {
	domain.CodeOfConductAcceptance

	Handle   string `json:"handle"`
	Realname string `json:"realname"`
	Email    string `json:"email"`
}

type CodeOfConductAcceptanceRepository interface {
	Create(ctx context.Context, acceptance *domain.CodeOfConductAcceptance) error
	GetByOrganizerAndCodeOfConduct(
		ctx context.Context,
		organizerID, codeOfConductID uuid.UUID,
	) (*domain.CodeOfConductAcceptance, error)
	ListByOrganizerID(ctx context.Context, organizerID uuid.UUID) ([]domain.CodeOfConductAcceptance, error)
	List(
		ctx context.Context,
		params CodeOfConductAcceptanceListParams,
		filters CodeOfConductAcceptanceListFilters,
	) ([]CodeOfConductAcceptanceWithOrganizer, int64, error)
}
