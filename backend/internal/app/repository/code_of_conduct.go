package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/potibm/crewapparat/internal/app/domain"
)

type CodeOfConductListFilters struct {
	Status *domain.CodeOfConductStatus
}

type CodeOfConductListParams = ListParams

type CodeOfConductRepository interface {
	Create(ctx context.Context, coc *domain.CodeOfConduct) error
	Update(ctx context.Context, coc *domain.CodeOfConduct) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(
		ctx context.Context,
		params CodeOfConductListParams,
		filters CodeOfConductListFilters,
	) ([]domain.CodeOfConduct, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.CodeOfConduct, error)
	GetLatestPublished(ctx context.Context) (*domain.CodeOfConduct, error)
	CountAcceptances(ctx context.Context, id uuid.UUID) (int64, error)
}
