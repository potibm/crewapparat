package gorm

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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

func NewOrganizerRepository(db *gorm.DB) repository.OrganizerRepository {
	return &organizerRepository{db: db}
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
	if err := r.db.WithContext(ctx).First(&existing, whereID, organizer.ID).Error; err != nil {
		return fmt.Errorf("failed to load organizer %s: %w", organizer.ID, err)
	}

	updated := fromDomainOrganizer(organizer)
	// Do not overwrite timestamp and audit columns.
	updated.CreatedAt = existing.CreatedAt
	updated.DeletedAt = existing.DeletedAt

	if err := r.db.WithContext(ctx).Save(updated).Error; err != nil {
		return fmt.Errorf("failed to update organizer %s: %w", organizer.ID, err)
	}

	return nil
}

func (r *organizerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&dbOrganizer{}, whereID, id)
	if result.Error != nil {
		return fmt.Errorf("delete organizer %s: %w", id, result.Error)
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
		org := toDomainOrganizer(&dbRows[i])

		// When filtered by event year, populate current participation data.
		if filters.EventYearID != nil {
			var participation dbParticipation

			err := r.db.WithContext(ctx).
				Preload("Teams").
				Where("organizer_id = ? AND event_year_id = ?", org.ID, *filters.EventYearID).
				First(&participation).Error
			if err == nil {
				org.CurrentParticipationStatus = participation.Status

				org.CurrentParticipationTeams = make([]domain.Team, 0, len(participation.Teams))
				for j := range participation.Teams {
					org.CurrentParticipationTeams = append(
						org.CurrentParticipationTeams,
						*toDomainTeam(&participation.Teams[j]),
					)
				}
			}
		}

		organizers = append(organizers, *org)
	}

	return organizers, total, nil
}

func (r *organizerRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Organizer, error) {
	var dbRow dbOrganizer

	err := r.db.WithContext(ctx).
		Preload("Participations.Teams").
		First(&dbRow, whereID, id).Error
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
		Where(whereEmail, email).
		First(&dbRow).Error
	if err != nil {
		return nil, err
	}

	return toDomainOrganizer(&dbRow), nil
}

func (r *organizerRepository) GetBySlug(ctx context.Context, slug string) (*domain.Organizer, error) {
	var dbRow dbOrganizer

	err := r.db.WithContext(ctx).
		Where("slug = ?", slug).
		First(&dbRow).Error
	if err != nil {
		return nil, err
	}

	return toDomainOrganizer(&dbRow), nil
}

func (r *organizerRepository) applyFilters(
	query *gorm.DB,
	filters repository.OrganizerListFilters,
) *gorm.DB {
	if filters.Query != nil {
		like := "%" + strings.TrimSpace(*filters.Query) + "%"
		// The parentheses are required: without them a later AND binds
		// tighter than the OR and the filter silently returns wrong rows.
		query = query.Where("(handle LIKE ? OR realname LIKE ? OR email LIKE ?)", like, like, like)
	}

	if filters.ID != nil {
		query = query.Where(whereID, *filters.ID)
	}

	if filters.Slug != nil {
		query = query.Where("slug = ?", *filters.Slug)
	}

	// Filter by participation data when event year is specified.
	if filters.EventYearID != nil {
		// Filter by status.
		if filters.Status != nil && *filters.Status != "all" {
			sub := r.db.Model(&dbParticipation{}).
				Select("organizer_id").
				Where("event_year_id = ?", *filters.EventYearID).
				Where("status = ?", *filters.Status)
			query = query.Where("id IN (?)", sub)
		}

		// Filter by team.
		if filters.TeamID != nil {
			sub := r.db.Model(&dbParticipation{}).
				Select("organizer_id").
				Joins("JOIN participation_teams ON participation_teams.participation_id = participations.id").
				Where("event_year_id = ?", *filters.EventYearID).
				Where("participation_teams.team_id = ?", *filters.TeamID)
			query = query.Where("id IN (?)", sub)
		}
	}

	return query
}

// backfillOrganizerSlugs populates the slug column for organizers that predate
// it (or were created without one). Slugs are immutable once set.
func backfillOrganizerSlugs(db *gorm.DB) error {
	var rows []dbOrganizer
	if err := db.Where("slug = '' OR slug IS NULL").Find(&rows).Error; err != nil {
		return err
	}

	used := map[string]bool{}

	var all []dbOrganizer
	if err := db.Select("slug").Find(&all).Error; err != nil {
		return err
	}

	for _, r := range all {
		if r.Slug != "" {
			used[r.Slug] = true
		}
	}

	for i := range rows {
		base := domain.Slugify(rows[i].Handle)

		slug := base
		for n := 2; used[slug]; n++ {
			slug = base + "-" + strconv.Itoa(n)
		}

		used[slug] = true
		rows[i].Slug = slug

		if err := db.Model(&dbOrganizer{}).Where(whereID, rows[i].ID).Update("slug", slug).Error; err != nil {
			return err
		}
	}

	return nil
}

// IsNotFound reports whether an error means the record does not exist.
func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
