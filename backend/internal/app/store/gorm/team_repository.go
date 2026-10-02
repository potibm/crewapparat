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

type teamRepository struct {
	db *gorm.DB
}

var teamSortColumns = map[string]string{
	"id":      "id",
	"name":    "name",
	"created": "created_at",
}

func (s *Store) NewTeamRepository() repository.TeamRepository {
	return NewTeamRepository(s.db)
}

func NewTeamRepository(db *gorm.DB) repository.TeamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) Create(ctx context.Context, team *domain.Team) error {
	dbObj := fromDomainTeam(team)

	if err := r.db.WithContext(ctx).Create(dbObj).Error; err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}

	team.ID = dbObj.ID

	return nil
}

func (r *teamRepository) Update(ctx context.Context, team *domain.Team) error {
	var existing dbTeam
	if err := r.db.WithContext(ctx).First(&existing, "id = ?", team.ID).Error; err != nil {
		return fmt.Errorf("failed to load team %s: %w", team.ID, err)
	}

	updated := fromDomainTeam(team)
	updated.CreatedAt = existing.CreatedAt
	updated.DeletedAt = existing.DeletedAt

	if err := r.db.WithContext(ctx).Save(updated).Error; err != nil {
		return fmt.Errorf("failed to update team %s: %w", team.ID, err)
	}

	return nil
}

func (r *teamRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&dbTeam{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete team %s: %w", id, result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("delete team %s: %w", id, gorm.ErrRecordNotFound)
	}

	return nil
}

func (r *teamRepository) List(
	ctx context.Context,
	params repository.TeamListParams,
	filters repository.TeamListFilters,
) ([]domain.Team, int64, error) {
	var (
		dbRows []dbTeam
		total  int64
	)

	query := r.db.WithContext(ctx).Model(&dbTeam{})

	if filters.Query != nil {
		like := "%" + strings.TrimSpace(*filters.Query) + "%"
		query = query.Where("(name LIKE ?)", like)
	}

	if filters.ID != nil {
		query = query.Where("id = ?", *filters.ID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count teams: %w", err)
	}

	sorted, err := applySort(query, teamSortColumns, params.Sort, params.Order)
	if err != nil {
		return nil, 0, err
	}

	if err := applyPagination(sorted, params).Find(&dbRows).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list teams: %w", err)
	}

	teams := make([]domain.Team, 0, len(dbRows))
	for i := range dbRows {
		teams = append(teams, *toDomainTeam(&dbRows[i]))
	}

	return teams, total, nil
}

func (r *teamRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Team, error) {
	var dbRow dbTeam

	if err := r.db.WithContext(ctx).First(&dbRow, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return toDomainTeam(&dbRow), nil
}

func (r *teamRepository) GetByName(ctx context.Context, name string) (*domain.Team, error) {
	var dbRow dbTeam

	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&dbRow).Error; err != nil {
		return nil, err
	}

	return toDomainTeam(&dbRow), nil
}
