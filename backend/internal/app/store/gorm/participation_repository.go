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

type participationRepository struct {
	db *gorm.DB
}

var participationSortColumns = map[string]string{
	"id":      "id",
	"status":  "status",
	"year":    "event_year_id",
	"created": "created_at",
}

func (s *Store) NewParticipationRepository() repository.ParticipationRepository {
	return NewParticipationRepository(s.db)
}

func (r *participationRepository) Create(ctx context.Context, participation *domain.Participation) error {
	teams := participation.Teams
	participation.Teams = nil

	dbObj := fromDomainParticipation(participation)

	if err := r.db.WithContext(ctx).Create(dbObj).Error; err != nil {
		return fmt.Errorf("failed to create participation: %w", err)
	}

	participation.ID = dbObj.ID
	participation.Teams = teams

	if len(teams) > 0 {
		if err := r.replaceTeams(ctx, dbObj.ID, teams); err != nil {
			return err
		}
	}

	return nil
}

func (r *participationRepository) Update(ctx context.Context, participation *domain.Participation) error {
	var existing dbParticipation
	if err := r.db.WithContext(ctx).First(&existing, "id = ?", participation.ID).Error; err != nil {
		return fmt.Errorf("failed to load participation %s: %w", participation.ID, err)
	}

	teams := participation.Teams

	updated := fromDomainParticipation(participation)
	updated.Teams = nil
	updated.CreatedAt = existing.CreatedAt
	updated.DeletedAt = existing.DeletedAt

	if err := r.db.WithContext(ctx).Omit("Teams").Save(updated).Error; err != nil {
		return fmt.Errorf("failed to update participation %s: %w", participation.ID, err)
	}

	if len(teams) > 0 {
		if err := r.replaceTeams(ctx, participation.ID, teams); err != nil {
			return err
		}
	}

	participation.Teams = teams

	return nil
}

func (r *participationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&dbParticipation{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete participation %s: %w", id, result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("delete participation %s: %w", id, gorm.ErrRecordNotFound)
	}

	if err := r.clearTeams(ctx, id); err != nil {
		return err
	}

	return nil
}

func (r *participationRepository) List(
	ctx context.Context,
	params repository.ParticipationListParams,
	filters repository.ParticipationListFilters,
) ([]domain.Participation, int64, error) {
	var (
		dbRows []dbParticipation
		total  int64
	)

	query := r.db.WithContext(ctx).Model(&dbParticipation{})

	if filters.Query != nil {
		like := "%" + strings.TrimSpace(*filters.Query) + "%"
		query = query.Where("(t_shirt_size LIKE ? OR status LIKE ?)", like, like)
	}

	if filters.ID != nil {
		query = query.Where("id = ?", *filters.ID)
	}

	if filters.OrganizerID != nil {
		query = query.Where("organizer_id = ?", *filters.OrganizerID)
	}

	if filters.EventYearID != nil {
		query = query.Where("event_year_id = ?", *filters.EventYearID)
	}

	if filters.Status != nil {
		query = query.Where("status = ?", string(*filters.Status))
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count participations: %w", err)
	}

	sorted, err := applySort(query, participationSortColumns, params.Sort, params.Order)
	if err != nil {
		return nil, 0, err
	}

	if err := applyPagination(sorted, params).Preload("Teams").Find(&dbRows).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list participations: %w", err)
	}

	participations := make([]domain.Participation, 0, len(dbRows))
	for i := range dbRows {
		participations = append(participations, *toDomainParticipation(&dbRows[i]))
	}

	return participations, total, nil
}

func (r *participationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Participation, error) {
	var dbRow dbParticipation

	err := r.db.WithContext(ctx).
		Preload("Teams").
		First(&dbRow, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	return toDomainParticipation(&dbRow), nil
}

func (r *participationRepository) GetByOrganizerAndYear(
	ctx context.Context,
	organizerID uuid.UUID,
	eventYearID string,
) (*domain.Participation, error) {
	var dbRow dbParticipation

	err := r.db.WithContext(ctx).
		Preload("Teams").
		Where("organizer_id = ? AND event_year_id = ?", organizerID, eventYearID).
		First(&dbRow).Error
	if err != nil {
		return nil, err
	}

	return toDomainParticipation(&dbRow), nil
}

func (r *participationRepository) SetTeams(
	ctx context.Context,
	participationID uuid.UUID,
	teams []domain.Team,
) error {
	var existing dbParticipation
	if err := r.db.WithContext(ctx).First(&existing, "id = ?", participationID).Error; err != nil {
		return fmt.Errorf("failed to load participation %s: %w", participationID, err)
	}

	return r.replaceTeams(ctx, participationID, teams)
}

// replaceTeams schreibt die Zuordnung neu. Die alten Join-Zeilen werden vorher
// hart gelöscht: Ein Soft-Delete der Teams würde sie sonst als Waisen
// zurücklassen, weil die Join-Tabelle keine DeletedAt-Spalte hat.

func NewParticipationRepository(db *gorm.DB) repository.ParticipationRepository {
	return &participationRepository{db: db}
}

func (r *participationRepository) replaceTeams(
	ctx context.Context,
	participationID uuid.UUID,
	teams []domain.Team,
) error {
	rows := make([]participationTeams, 0, len(teams))
	for i := range teams {
		rows = append(rows, participationTeams{
			ParticipationID: participationID,
			TeamID:          teams[i].ID,
		})
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			"DELETE FROM participation_teams WHERE participation_id = ?",
			participationID,
		).Error; err != nil {
			return fmt.Errorf("failed to clear participation teams: %w", err)
		}

		if len(rows) == 0 {
			return nil
		}

		if err := tx.CreateInBatches(&rows, joinBatchSize).Error; err != nil {
			return fmt.Errorf("failed to insert participation teams: %w", err)
		}

		return nil
	})
}

func (r *participationRepository) clearTeams(ctx context.Context, participationID uuid.UUID) error {
	err := r.db.WithContext(ctx).Exec(
		"DELETE FROM participation_teams WHERE participation_id = ?",
		participationID,
	).Error
	if err != nil {
		return fmt.Errorf("failed to clear participation teams: %w", err)
	}

	return nil
}
