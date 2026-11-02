package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/potibm/crewapparat/internal/app/domain"
)

type TeamListFilters struct {
	Query  *string
	ID     *uuid.UUID
	Active *bool
}

type TeamListParams = ListParams

type TeamRepository interface {
	Create(ctx context.Context, team *domain.Team) error
	Update(ctx context.Context, team *domain.Team) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, params TeamListParams, filters TeamListFilters) ([]domain.Team, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Team, error)
	GetByName(ctx context.Context, name string) (*domain.Team, error)
}
