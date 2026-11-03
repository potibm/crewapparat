package gorm

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/potibm/crewapparat/internal/app/domain"
	"github.com/potibm/crewapparat/internal/app/repository"
)

type codeOfConductAcceptanceRepository struct {
	db *gorm.DB
}

var codeOfConductAcceptanceSortColumns = map[string]string{
	"id":          "id",
	"accepted":    "accepted_at",
	"accepted_at": "accepted_at",
	"organizer":   "organizer_id",
	"version":     "code_of_conduct_id",
}

func (s *Store) NewCodeOfConductAcceptanceRepository() repository.CodeOfConductAcceptanceRepository {
	return NewCodeOfConductAcceptanceRepository(s.db)
}

func NewCodeOfConductAcceptanceRepository(db *gorm.DB) repository.CodeOfConductAcceptanceRepository {
	return &codeOfConductAcceptanceRepository{db: db}
}

func (r *codeOfConductAcceptanceRepository) Create(
	ctx context.Context,
	acceptance *domain.CodeOfConductAcceptance,
) error {
	dbObj := fromDomainCodeOfConductAcceptance(acceptance)

	if err := r.db.WithContext(ctx).Create(dbObj).Error; err != nil {
		return fmt.Errorf("failed to create code of conduct acceptance: %w", err)
	}

	acceptance.ID = dbObj.ID

	return nil
}

func (r *codeOfConductAcceptanceRepository) GetByOrganizerAndCodeOfConduct(
	ctx context.Context,
	organizerID, codeOfConductID uuid.UUID,
) (*domain.CodeOfConductAcceptance, error) {
	var row dbCodeOfConductAcceptance

	err := r.db.WithContext(ctx).
		Where("organizer_id = ? AND code_of_conduct_id = ?", organizerID, codeOfConductID).
		First(&row).Error
	if err != nil {
		return nil, err
	}

	return toDomainCodeOfConductAcceptance(&row), nil
}

func (r *codeOfConductAcceptanceRepository) ListByOrganizerID(
	ctx context.Context,
	organizerID uuid.UUID,
) ([]domain.CodeOfConductAcceptance, error) {
	var rows []dbCodeOfConductAcceptance

	err := r.db.WithContext(ctx).
		Where("organizer_id = ?", organizerID).
		Order("accepted_at DESC").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list code of conduct acceptances: %w", err)
	}

	acceptances := make([]domain.CodeOfConductAcceptance, 0, len(rows))
	for i := range rows {
		acceptances = append(acceptances, *toDomainCodeOfConductAcceptance(&rows[i]))
	}

	return acceptances, nil
}

func (r *codeOfConductAcceptanceRepository) List(
	ctx context.Context,
	params repository.CodeOfConductAcceptanceListParams,
	filters repository.CodeOfConductAcceptanceListFilters,
) ([]repository.CodeOfConductAcceptanceWithOrganizer, int64, error) {
	var (
		rows  []dbCodeOfConductAcceptance
		total int64
	)

	query := r.db.WithContext(ctx).Model(&dbCodeOfConductAcceptance{})

	if filters.OrganizerID != nil {
		query = query.Where("organizer_id = ?", *filters.OrganizerID)
	}

	if filters.CodeOfConductID != nil {
		query = query.Where("code_of_conduct_id = ?", *filters.CodeOfConductID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count code of conduct acceptances: %w", err)
	}

	sorted, err := applySort(query, codeOfConductAcceptanceSortColumns, params.Sort, params.Order)
	if err != nil {
		return nil, 0, err
	}

	if err := applyPagination(sorted, params).Find(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list code of conduct acceptances: %w", err)
	}

	result, err := r.withOrganizers(ctx, rows)
	if err != nil {
		return nil, 0, err
	}

	return result, total, nil
}

// withOrganizers resolves the organizer data for the given acceptances in one
// query instead of joining, which keeps the sortable columns unambiguous.
func (r *codeOfConductAcceptanceRepository) withOrganizers(
	ctx context.Context,
	rows []dbCodeOfConductAcceptance,
) ([]repository.CodeOfConductAcceptanceWithOrganizer, error) {
	result := make([]repository.CodeOfConductAcceptanceWithOrganizer, 0, len(rows))

	if len(rows) == 0 {
		return result, nil
	}

	organizerIDs := make([]uuid.UUID, 0, len(rows))
	for i := range rows {
		organizerIDs = append(organizerIDs, rows[i].OrganizerID)
	}

	var organizers []dbOrganizer
	if err := r.db.WithContext(ctx).Where("id IN ?", organizerIDs).Find(&organizers).Error; err != nil {
		return nil, fmt.Errorf("failed to load organizers: %w", err)
	}

	byID := make(map[uuid.UUID]dbOrganizer, len(organizers))
	for i := range organizers {
		byID[organizers[i].ID] = organizers[i]
	}

	for i := range rows {
		organizer := byID[rows[i].OrganizerID]

		result = append(result, repository.CodeOfConductAcceptanceWithOrganizer{
			CodeOfConductAcceptance: *toDomainCodeOfConductAcceptance(&rows[i]),
			Handle:                  organizer.Handle,
			Realname:                organizer.Realname,
			Email:                   organizer.Email,
		})
	}

	return result, nil
}
