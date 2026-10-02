package gorm

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/potibm/crewapparat/internal/app/domain"
	"github.com/potibm/crewapparat/internal/app/repository"
)

type organizerRepository struct {
	db *gorm.DB
}

var organizerSortColumns = map[string]string{
	"id":       "id",
	"handle":   "handle",
	"realname": "realname",
	"email":    "email",
	"created":  "created_at",
	"updated":  "updated_at",
}

func (s *Store) NewOrganizerRepository() repository.OrganizerRepository {
	return NewOrganizerRepository(s.db)
}

func (r *organizerRepository) Create(ctx context.Context, organizer *domain.Organizer) error {
	dbObj := fromDomainOrganizer(organizer)

	if err := r.db.WithContext(ctx).Create(dbObj).Error; err != nil {
		return fmt.Errorf("failed to create organizer: %w", err)
	}

	organizer.ID = dbObj.ID

	return nil
}

func (r *organizerRepository) Update(ctx context.Context, organizer *domain.Organizer) error {
	var existing dbOrganizer
	if err := r.db.WithContext(ctx).First(&existing, "id = ?", organizer.ID).Error; err != nil {
		return fmt.Errorf("failed to load organizer %s: %w", organizer.ID, err)
	}

	updated := fromDomainOrganizer(organizer)
	// Zeitstempel und Audit-Spalten nicht überschreiben.
	updated.CreatedAt = existing.CreatedAt
	updated.DeletedAt = existing.DeletedAt

	if err := r.db.WithContext(ctx).Save(updated).Error; err != nil {
		return fmt.Errorf("failed to update organizer %s: %w", organizer.ID, err)
	}

	return nil
}

func (r *organizerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&dbOrganizer{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete organizer %s: %w", id, result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("delete organizer %s: %w", id, gorm.ErrRecordNotFound)
	}

	return nil
}

func (r *organizerRepository) List(
	ctx context.Context,
	params repository.OrganizerListParams,
	filters repository.OrganizerListFilters,
) ([]domain.Organizer, int64, error) {
	var (
		dbRows []dbOrganizer
		total  int64
	)

	query := r.db.WithContext(ctx).Model(&dbOrganizer{})

	query = r.applyFilters(query, filters)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count organizers: %w", err)
	}

	sorted, err := applySort(query, organizerSortColumns, params.Sort, params.Order)
	if err != nil {
		return nil, 0, err
	}

	if err := applyPagination(sorted, params).Find(&dbRows).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list organizers: %w", err)
	}

	organizers := make([]domain.Organizer, 0, len(dbRows))
	for i := range dbRows {
		organizers = append(organizers, *toDomainOrganizer(&dbRows[i]))
	}

	return organizers, total, nil
}

func (r *organizerRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Organizer, error) {
	var dbRow dbOrganizer

	err := r.db.WithContext(ctx).
		Preload("Participations").
		First(&dbRow, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	return toDomainOrganizer(&dbRow), nil
}

func (r *organizerRepository) GetByOidcSub(ctx context.Context, oidcSub string) (*domain.Organizer, error) {
	var dbRow dbOrganizer

	err := r.db.WithContext(ctx).
		Where("oidc_sub = ?", oidcSub).
		First(&dbRow).Error
	if err != nil {
		return nil, err
	}

	return toDomainOrganizer(&dbRow), nil
}

func (r *organizerRepository) GetByEmail(ctx context.Context, email string) (*domain.Organizer, error) {
	var dbRow dbOrganizer

	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&dbRow).Error
	if err != nil {
		return nil, err
	}

	return toDomainOrganizer(&dbRow), nil
}

func NewOrganizerRepository(db *gorm.DB) repository.OrganizerRepository {
	return &organizerRepository{db: db}
}

func (r *organizerRepository) applyFilters(
	query *gorm.DB,
	filters repository.OrganizerListFilters,
) *gorm.DB {
	if filters.Query != nil {
		like := "%" + strings.TrimSpace(*filters.Query) + "%"
		// Klammern sind Pflicht: ohne sie bindet das spätere AND stärker als das
		// OR und der Filter liefert stillschweigend falsche Treffer.
		query = query.Where("(handle LIKE ? OR realname LIKE ? OR email LIKE ?)", like, like, like)
	}

	if filters.ID != nil {
		query = query.Where("id = ?", *filters.ID)
	}

	return query
}

// IsNotFound meldet, ob ein Fehler auf einen fehlenden Datensatz zurückgeht.
func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
