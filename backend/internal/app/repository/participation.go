package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/potibm/crewapparat/internal/app/domain"
)

type ParticipationListFilters struct {
	Query       *string
	ID          *uuid.UUID
	OrganizerID *uuid.UUID
	EventYearID *string
	Status      *domain.ParticipationStatus
}

type ParticipationListParams = ListParams

type ParticipationRepository interface {
	Create(ctx context.Context, participation *domain.Participation) error
	Update(ctx context.Context, participation *domain.Participation) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(
		ctx context.Context,
		params ParticipationListParams,
		filters ParticipationListFilters,
	) ([]domain.Participation, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Participation, error)
	GetByOrganizerAndYear(
		ctx context.Context,
		organizerID uuid.UUID,
		eventYearID string,
	) (*domain.Participation, error)
	// SetTeams replaces the team assignment of a participation. Join rows are
	// deleted explicitly, because a soft delete of the involved teams would
	// otherwise leave orphaned rows behind.
	SetTeams(ctx context.Context, participationID uuid.UUID, teams []domain.Team) error
}
