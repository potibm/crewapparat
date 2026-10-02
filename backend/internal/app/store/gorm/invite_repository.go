package gorm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/potibm/crewapparat/internal/app/domain"
	"github.com/potibm/crewapparat/internal/app/repository"
)

type inviteRepository struct {
	db *gorm.DB
}

var inviteSortColumns = map[string]string{
	"id":      "id",
	"email":   "email",
	"status":  "status",
	"expires": "expires_at",
	"created": "created_at",
}

func (s *Store) NewInviteRepository() repository.InviteRepository {
	return NewInviteRepository(s.db)
}

func (s *Store) NewInviteTokenRepository() repository.InviteTokenRepository {
	return NewInviteTokenRepository(s.db)
}

func (s *Store) NewInviteWorkflowRepository() repository.InviteWorkflowRepository {
	return NewInviteWorkflowRepository(s.db)
}

func (s *Store) NewInviteTeamRepository() repository.InviteTeamRepository {
	return NewInviteTeamRepository(s.db)
}

func (r *inviteRepository) Create(ctx context.Context, invite *domain.Invite) error {
	teams := invite.Teams
	invite.Teams = nil

	dbObj := fromDomainInvite(invite)

	if err := r.db.WithContext(ctx).Omit("Teams").Create(dbObj).Error; err != nil {
		return fmt.Errorf("failed to create invite: %w", err)
	}

	invite.ID = dbObj.ID
	invite.Teams = teams

	if len(teams) > 0 {
		if err := r.replaceTeams(ctx, dbObj.ID, teams); err != nil {
			return err
		}
	}

	return nil
}

func (r *inviteRepository) Update(ctx context.Context, invite *domain.Invite) error {
	var existing dbInvite
	if err := r.db.WithContext(ctx).First(&existing, "id = ?", invite.ID).Error; err != nil {
		return fmt.Errorf("failed to load invite %s: %w", invite.ID, err)
	}

	teams := invite.Teams

	updated := fromDomainInvite(invite)
	updated.Teams = nil
	updated.CreatedAt = existing.CreatedAt
	updated.DeletedAt = existing.DeletedAt

	if err := r.db.WithContext(ctx).Omit("Teams").Save(updated).Error; err != nil {
		return fmt.Errorf("failed to update invite %s: %w", invite.ID, err)
	}

	invite.Teams = teams

	if len(teams) > 0 {
		if err := r.replaceTeams(ctx, invite.ID, teams); err != nil {
			return err
		}
	}

	return nil
}

func (r *inviteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&dbInvite{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete invite %s: %w", id, result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("delete invite %s: %w", id, gorm.ErrRecordNotFound)
	}

	if err := r.clearTeams(ctx, id); err != nil {
		return err
	}

	return nil
}

func (r *inviteRepository) List(
	ctx context.Context,
	params repository.InviteListParams,
	filters repository.InviteListFilters,
) ([]domain.Invite, int64, error) {
	var (
		dbRows []dbInvite
		total  int64
	)

	query := r.db.WithContext(ctx).Model(&dbInvite{})

	if filters.Query != nil {
		like := "%" + strings.TrimSpace(*filters.Query) + "%"
		query = query.Where("(email LIKE ? OR provided_name LIKE ?)", like, like)
	}

	if filters.ID != nil {
		query = query.Where("id = ?", *filters.ID)
	}

	if filters.Email != nil {
		query = query.Where("email = ?", *filters.Email)
	}

	if filters.Status != nil {
		query = query.Where("status = ?", string(*filters.Status))
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count invites: %w", err)
	}

	sorted, err := applySort(query, inviteSortColumns, params.Sort, params.Order)
	if err != nil {
		return nil, 0, err
	}

	if err := applyPagination(sorted, params).Preload("Teams").Find(&dbRows).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list invites: %w", err)
	}

	invites := make([]domain.Invite, 0, len(dbRows))
	for i := range dbRows {
		invites = append(invites, *toDomainInvite(&dbRows[i]))
	}

	return invites, total, nil
}

func (r *inviteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Invite, error) {
	var dbRow dbInvite

	err := r.db.WithContext(ctx).Preload("Teams").First(&dbRow, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	return toDomainInvite(&dbRow), nil
}

func (r *inviteRepository) GetByTokenHash(
	ctx context.Context,
	tokenHash string,
) (*domain.Invite, error) {
	var dbRow dbInvite

	err := r.db.WithContext(ctx).
		Preload("Teams").
		Where("token_hash = ?", tokenHash).
		First(&dbRow).Error
	if err != nil {
		return nil, err
	}

	return toDomainInvite(&dbRow), nil
}

func (r *inviteRepository) GetByEmail(ctx context.Context, email string) (*domain.Invite, error) {
	var dbRow dbInvite

	err := r.db.WithContext(ctx).
		Preload("Teams").
		Where("email = ?", email).
		First(&dbRow).Error
	if err != nil {
		return nil, err
	}

	return toDomainInvite(&dbRow), nil
}

// EmailExists prüft, ob die Adresse schon belegt ist - entweder im
// Crew-Verzeichnis oder in einer offenen Einladung.
func (r *inviteRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Model(&dbOrganizer{}).
		Where("email = ?", email).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check organizer email: %w", err)
	}

	if count > 0 {
		return true, nil
	}

	if err := r.db.WithContext(ctx).
		Model(&dbInvite{}).
		Where("email = ?", email).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check invite email: %w", err)
	}

	return count > 0, nil
}

func (r *inviteRepository) SubmitForReview(
	ctx context.Context,
	id uuid.UUID,
	providedName string,
	providedPhone string,
	notes string,
	cocAccepted bool,
) error {
	updates := map[string]any{
		"provided_name":  providedName,
		"provided_phone": providedPhone,
		"notes":          notes,
		"coc_accepted":   cocAccepted,
		"status":         string(domain.InviteStatusPendingReview),
	}

	if cocAccepted {
		updates["coc_accepted_at"] = time.Now().UTC()
	}

	result := r.db.WithContext(ctx).
		Model(&dbInvite{}).
		Where("id = ?", id).
		Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to submit invite %s: %w", id, result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("submit invite %s: %w", id, gorm.ErrRecordNotFound)
	}

	return nil
}

func (r *inviteRepository) Review(
	ctx context.Context,
	id uuid.UUID,
	status domain.InviteStatus,
	reviewerID uuid.UUID,
) error {
	result := r.db.WithContext(ctx).
		Model(&dbInvite{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":         string(status),
			"reviewed_by_id": reviewerID,
		})
	if result.Error != nil {
		return fmt.Errorf("failed to review invite %s: %w", id, result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("review invite %s: %w", id, gorm.ErrRecordNotFound)
	}

	return nil
}

func (r *inviteRepository) SetTeams(
	ctx context.Context,
	inviteID uuid.UUID,
	teams []domain.Team,
) error {
	var existing dbInvite
	if err := r.db.WithContext(ctx).First(&existing, "id = ?", inviteID).Error; err != nil {
		return fmt.Errorf("failed to load invite %s: %w", inviteID, err)
	}

	return r.replaceTeams(ctx, inviteID, teams)
}

func (r *inviteRepository) PurgeExpired(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at < ? AND status = ?", before, string(domain.InviteStatusSent)).
		Delete(&dbInvite{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to purge expired invites: %w", result.Error)
	}

	return result.RowsAffected, nil
}

func NewInviteRepository(db *gorm.DB) repository.InviteRepository {
	return &inviteRepository{db: db}
}

func NewInviteTokenRepository(db *gorm.DB) repository.InviteTokenRepository {
	return &inviteRepository{db: db}
}

func NewInviteWorkflowRepository(db *gorm.DB) repository.InviteWorkflowRepository {
	return &inviteRepository{db: db}
}

func NewInviteTeamRepository(db *gorm.DB) repository.InviteTeamRepository {
	return &inviteRepository{db: db}
}

func (r *inviteRepository) replaceTeams(
	ctx context.Context,
	inviteID uuid.UUID,
	teams []domain.Team,
) error {
	rows := make([]inviteTeams, 0, len(teams))
	for i := range teams {
		rows = append(rows, inviteTeams{
			InviteID: inviteID,
			TeamID:   teams[i].ID,
		})
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			"DELETE FROM invite_teams WHERE invite_id = ?",
			inviteID,
		).Error; err != nil {
			return fmt.Errorf("failed to clear invite teams: %w", err)
		}

		if len(rows) == 0 {
			return nil
		}

		if err := tx.CreateInBatches(&rows, joinBatchSize).Error; err != nil {
			return fmt.Errorf("failed to insert invite teams: %w", err)
		}

		return nil
	})
}

func (r *inviteRepository) clearTeams(ctx context.Context, inviteID uuid.UUID) error {
	err := r.db.WithContext(ctx).Exec(
		"DELETE FROM invite_teams WHERE invite_id = ?",
		inviteID,
	).Error
	if err != nil {
		return fmt.Errorf("failed to clear invite teams: %w", err)
	}

	return nil
}

// PurgeExpired entfernt Einladungen, die abgelaufen sind und noch nicht
// eingereicht wurden. Eingereichte bleiben zur Nachvollziehbarkeit erhalten.
