package gorm

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/potibm/crewapparat/internal/app/domain"
	"github.com/potibm/crewapparat/internal/app/repository"
)

type eventYearRepository struct {
	db *gorm.DB
}

var eventYearSortColumns = map[string]string{
	"year":     "year",
	"isactive": "is_active",
}

func (s *Store) NewEventYearRepository() repository.EventYearRepository {
	return NewEventYearRepository(s.db)
}

func NewEventYearRepository(db *gorm.DB) repository.EventYearRepository {
	return &eventYearRepository{db: db}
}

func (r *eventYearRepository) Create(ctx context.Context, year *domain.EventYear) error {
	dbObj := fromDomainEventYear(year)

	if err := r.db.WithContext(ctx).Create(dbObj).Error; err != nil {
		return fmt.Errorf("failed to create event year %s: %w", year.Year, err)
	}

	return nil
}

func (r *eventYearRepository) Update(ctx context.Context, year *domain.EventYear) error {
	var existing dbEventYear
	if err := r.db.WithContext(ctx).First(&existing, "year = ?", year.Year).Error; err != nil {
		return fmt.Errorf("failed to load event year %s: %w", year.Year, err)
	}

	updated := fromDomainEventYear(year)
	updated.CreatedAt = existing.CreatedAt
	updated.DeletedAt = existing.DeletedAt

	if err := r.db.WithContext(ctx).Save(updated).Error; err != nil {
		return fmt.Errorf("failed to update event year %s: %w", year.Year, err)
	}

	return nil
}

func (r *eventYearRepository) Delete(ctx context.Context, year string) error {
	result := r.db.WithContext(ctx).Delete(&dbEventYear{}, "year = ?", year)
	if result.Error != nil {
		return fmt.Errorf("failed to delete event year %s: %w", year, result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("delete event year %s: %w", year, gorm.ErrRecordNotFound)
	}

	return nil
}

func (r *eventYearRepository) List(
	ctx context.Context,
	params repository.EventYearListParams,
	filters repository.EventYearListFilters,
) ([]domain.EventYear, int64, error) {
	var (
		dbRows []dbEventYear
		total  int64
	)

	query := r.db.WithContext(ctx).Model(&dbEventYear{})

	if filters.Query != nil {
		like := "%" + strings.TrimSpace(*filters.Query) + "%"
		query = query.Where("(year LIKE ?)", like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count event years: %w", err)
	}

	sorted, err := applySort(query, eventYearSortColumns, params.Sort, params.Order)
	if err != nil {
		return nil, 0, err
	}

	if err := applyPagination(sorted, params).Find(&dbRows).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list event years: %w", err)
	}

	years := make([]domain.EventYear, 0, len(dbRows))
	for i := range dbRows {
		years = append(years, *toDomainEventYear(&dbRows[i]))
	}

	return years, total, nil
}

func (r *eventYearRepository) GetByYear(ctx context.Context, year string) (*domain.EventYear, error) {
	var dbRow dbEventYear

	if err := r.db.WithContext(ctx).First(&dbRow, "year = ?", year).Error; err != nil {
		return nil, err
	}

	return toDomainEventYear(&dbRow), nil
}

func (r *eventYearRepository) GetActive(ctx context.Context) (*domain.EventYear, error) {
	var dbRow dbEventYear

	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("year DESC").
		First(&dbRow).Error
	if err != nil {
		return nil, err
	}

	return toDomainEventYear(&dbRow), nil
}

// ActivateYear deaktiviert alle Jahre und aktiviert genau das angegebene. Das
// läuft in einer Transaktion, damit nie zwei Jahre aktiv sind - SQLite kann
// keinen partiellen Unique-Index darauf erzwingen.
func (r *eventYearRepository) ActivateYear(ctx context.Context, year string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing dbEventYear
		if err := tx.First(&existing, "year = ?", year).Error; err != nil {
			return fmt.Errorf("event year %s not found: %w", year, err)
		}

		if err := tx.Model(&dbEventYear{}).
			Where("is_active = ?", true).
			Update("is_active", false).Error; err != nil {
			return fmt.Errorf("failed to deactivate event years: %w", err)
		}

		if err := tx.Model(&dbEventYear{}).
			Where("year = ?", year).
			Update("is_active", true).Error; err != nil {
			return fmt.Errorf("failed to activate event year %s: %w", year, err)
		}

		return nil
	})
}
