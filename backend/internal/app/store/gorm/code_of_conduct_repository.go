package gorm

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/potibm/crewapparat/internal/app/domain"
	"github.com/potibm/crewapparat/internal/app/repository"
)

type codeOfConductRepository struct {
	db *gorm.DB
}

var codeOfConductSortColumns = map[string]string{
	"id":        "id",
	"status":    "status",
	"created":   "created_at",
	"published": "published_at",
}

func (s *Store) NewCodeOfConductRepository() repository.CodeOfConductRepository {
	return NewCodeOfConductRepository(s.db)
}

func NewCodeOfConductRepository(db *gorm.DB) repository.CodeOfConductRepository {
	return &codeOfConductRepository{db: db}
}

func (r *codeOfConductRepository) Create(ctx context.Context, coc *domain.CodeOfConduct) error {
	dbObj := fromDomainCodeOfConduct(coc)

	if err := r.db.WithContext(ctx).Create(dbObj).Error; err != nil {
		return fmt.Errorf("failed to create code of conduct: %w", err)
	}

	coc.ID = dbObj.ID

	return nil
}

func (r *codeOfConductRepository) Update(ctx context.Context, coc *domain.CodeOfConduct) error {
	var existing dbCodeOfConduct
	if err := r.db.WithContext(ctx).First(&existing, "id = ?", coc.ID).Error; err != nil {
		return fmt.Errorf("failed to load code of conduct %s: %w", coc.ID, err)
	}

	updated := fromDomainCodeOfConduct(coc)
	updated.CreatedAt = existing.CreatedAt
	updated.DeletedAt = existing.DeletedAt

	if err := r.db.WithContext(ctx).Save(updated).Error; err != nil {
		return fmt.Errorf("failed to update code of conduct %s: %w", coc.ID, err)
	}

	return nil
}

func (r *codeOfConductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&dbCodeOfConduct{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete code of conduct %s: %w", id, result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("delete code of conduct %s: %w", id, gorm.ErrRecordNotFound)
	}

	return nil
}

func (r *codeOfConductRepository) List(
	ctx context.Context,
	params repository.CodeOfConductListParams,
	filters repository.CodeOfConductListFilters,
) ([]domain.CodeOfConduct, int64, error) {
	var (
		dbRows []dbCodeOfConduct
		total  int64
	)

	query := r.db.WithContext(ctx).Model(&dbCodeOfConduct{})

	if filters.Status != nil {
		query = query.Where("status = ?", string(*filters.Status))
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count code of conducts: %w", err)
	}

	sorted, err := applySort(query, codeOfConductSortColumns, params.Sort, params.Order)
	if err != nil {
		return nil, 0, err
	}

	if err := applyPagination(sorted, params).Find(&dbRows).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list code of conducts: %w", err)
	}

	cocs := make([]domain.CodeOfConduct, 0, len(dbRows))
	for i := range dbRows {
		cocs = append(cocs, *toDomainCodeOfConduct(&dbRows[i]))
	}

	return cocs, total, nil
}

func (r *codeOfConductRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.CodeOfConduct, error) {
	var dbRow dbCodeOfConduct

	if err := r.db.WithContext(ctx).First(&dbRow, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return toDomainCodeOfConduct(&dbRow), nil
}

func (r *codeOfConductRepository) GetLatestPublished(ctx context.Context) (*domain.CodeOfConduct, error) {
	var dbRow dbCodeOfConduct

	err := r.db.WithContext(ctx).
		Where("status = ?", string(domain.CodeOfConductStatusPublished)).
		Order("published_at DESC").
		First(&dbRow).Error
	if err != nil {
		return nil, err
	}

	return toDomainCodeOfConduct(&dbRow), nil
}
