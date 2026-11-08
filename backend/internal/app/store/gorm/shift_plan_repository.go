package gorm

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/potibm/crewapparat/internal/app/domain"
	"github.com/potibm/crewapparat/internal/app/repository"
)

type shiftPlanRepository struct {
	db *gorm.DB
}

func (s *Store) NewShiftPlanRepository() repository.ShiftPlanRepository {
	return NewShiftPlanRepository(s.db)
}

func NewShiftPlanRepository(db *gorm.DB) repository.ShiftPlanRepository {
	return &shiftPlanRepository{db: db}
}

func (r *shiftPlanRepository) Create(ctx context.Context, plan *domain.ShiftPlan) error {
	dbObj := fromDomainShiftPlan(plan)
	if err := r.db.WithContext(ctx).Create(dbObj).Error; err != nil {
		return fmt.Errorf("failed to create shift plan: %w", err)
	}

	plan.ID = dbObj.ID

	return nil
}

func (r *shiftPlanRepository) List(
	ctx context.Context,
	params repository.ShiftPlanListParams,
	filters repository.ShiftPlanListFilters,
) ([]domain.ShiftPlan, int64, error) {
	query := r.db.WithContext(ctx).Model(&dbShiftPlan{})

	if filters.EventYearID != nil {
		query = query.Where("event_year_id = ?", *filters.EventYearID)
	}

	if filters.Status != nil {
		query = query.Where("status = ?", string(*filters.Status))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count shift plans: %w", err)
	}

	sortColumn := "created_at"

	if params.Sort != "" {
		// Allow only known columns to avoid SQL injection.
		allowed := map[string]bool{
			"id": true, "event_year_id": true, "version": true,
			"status": true, "created_at": true, "updated_at": true,
		}
		if allowed[params.Sort] {
			sortColumn = params.Sort
		}
	}

	order := "ASC"
	if strings.EqualFold(params.Order, "DESC") {
		order = "DESC"
	}

	offset := params.Offset

	limit := params.Limit
	if limit <= 0 {
		limit = 100
	}

	var rows []dbShiftPlan
	if err := query.
		Order(sortColumn + " " + order).
		Offset(offset).
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list shift plans: %w", err)
	}

	result := make([]domain.ShiftPlan, 0, len(rows))
	for i := range rows {
		result = append(result, *toDomainShiftPlan(&rows[i]))
	}

	return result, total, nil
}

func (r *shiftPlanRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ShiftPlan, error) {
	var row dbShiftPlan
	if err := r.db.WithContext(ctx).First(&row, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return toDomainShiftPlan(&row), nil
}

func (r *shiftPlanRepository) GetLatestPublished(ctx context.Context, eventYearID string) (*domain.ShiftPlan, error) {
	var row dbShiftPlan

	query := r.db.WithContext(ctx).Where("status = ?", string(domain.ShiftPlanStatusPublished))
	if eventYearID != "" {
		query = query.Where("event_year_id = ?", eventYearID)
	}

	if err := query.Order("created_at DESC").First(&row).Error; err != nil {
		return nil, err
	}

	return toDomainShiftPlan(&row), nil
}
