package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/potibm/crewapparat/internal/app/domain"
)

type ShiftPlanListFilters struct {
	EventYearID *string
	Status      *domain.ShiftPlanStatus
}

type ShiftPlanListParams = ListParams

type ShiftPlanRepository interface {
	Create(ctx context.Context, plan *domain.ShiftPlan) error
	List(
		ctx context.Context,
		params ShiftPlanListParams,
		filters ShiftPlanListFilters,
	) ([]domain.ShiftPlan, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ShiftPlan, error)
	GetLatestPublished(ctx context.Context, eventYearID string) (*domain.ShiftPlan, error)
}
