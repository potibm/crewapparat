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
	"id":         "id",
	"email":      "email",
	"handle":     "handle",
	"status":     "status",
	"expires_at": "expires_at",
	"created_at": "created_at",
}

func (s *Store) NewInviteRepository() repository.InviteRepository {
	return NewInviteRepository(s.db)
}

func NewInviteRepository(db *gorm.DB) repository.InviteRepository {
	return &inviteRepository{db: db}
}

func (s *Store) NewInviteWorkflowRepository() repository.InviteWorkflowRepository {
	return NewInviteWorkflowRepository(s.db)
}

func NewInviteWorkflowRepository(db *gorm.DB) repository.InviteWorkflowRepository {
	return &inviteRepository{db: db}
}

func (r *inviteRepository) Create(ctx context.Context, invite *domain.Invite) error {
	dbObj := fromDomainInvite(invite)

	if err := r.db.WithContext(ctx).Create(dbObj).Error; err != nil {
		return fmt.Errorf("failed to create invite: %w", err)
	}

	invite.ID = dbObj.ID

	return nil
}

func (r *inviteRepository) Update(ctx context.Context, invite *domain.Invite) error {
	var existing dbInvite
	if err := r.db.WithContext(ctx).First(&existing, whereID, invite.ID).Error; err != nil {
		return fmt.Errorf("failed to load invite %s: %w", invite.ID, err)
	}

	updated := fromDomainInvite(invite)
	updated.CreatedAt = existing.CreatedAt
	updated.DeletedAt = existing.DeletedAt

	if err := r.db.WithContext(ctx).Save(updated).Error; err != nil {
		return fmt.Errorf("failed to update invite %s: %w", invite.ID, err)
	}

	return nil
}

func (r *inviteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&dbInvite{}, whereID, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete invite %s: %w", id, result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("delete invite %s: %w", id, gorm.ErrRecordNotFound)
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
		// The parentheses are required: without them a second filter added via
		// .Where() binds tighter than the OR and silently returns wrong rows.
		query = query.Where("(email LIKE ? OR provided_name LIKE ?)", like, like)
	}

	if filters.ID != nil {
		query = query.Where(whereID, *filters.ID)
	}

	if filters.Email != nil {
		query = query.Where(whereEmail, *filters.Email)
	}

	if filters.Status != nil {
		query = query.Where("status = ?", string(*filters.Status))
	}

	if filters.InvitedByID != nil {
		query = query.Where("invited_by_id = ?", *filters.InvitedByID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count invites: %w", err)
	}

	sorted, err := applySort(query, inviteSortColumns, params.Sort, params.Order)
	if err != nil {
		return nil, 0, err
	}

	if err := applyPagination(sorted, params).Find(&dbRows).Error; err != nil {
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

	if err := r.db.WithContext(ctx).First(&dbRow, whereID, id).Error; err != nil {
		return nil, err
	}

	return toDomainInvite(&dbRow), nil
}

func (r *inviteRepository) GetByEmail(ctx context.Context, email string) (*domain.Invite, error) {
	var dbRow dbInvite

	if err := r.db.WithContext(ctx).Where(whereEmail, email).First(&dbRow).Error; err != nil {
		return nil, err
	}

	return toDomainInvite(&dbRow), nil
}

func (r *inviteRepository) GetByToken(ctx context.Context, token string) (*domain.Invite, error) {
	var dbRow dbInvite

	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&dbRow).Error; err != nil {
		return nil, err
	}

	return toDomainInvite(&dbRow), nil
}

// EmailExists reports whether the address is already taken, either in the crew
// directory or in an open invitation.
func (r *inviteRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Model(&dbOrganizer{}).
		Where(whereEmail, email).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check organizer email: %w", err)
	}

	if count > 0 {
		return true, nil
	}

	if err := r.db.WithContext(ctx).
		Model(&dbInvite{}).
		Where(whereEmail, email).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check invite email: %w", err)
	}

	return count > 0, nil
}

func (r *inviteRepository) SubmitForReview(
	ctx context.Context,
	id uuid.UUID,
	notes string,
	cocAccepted bool,
	cocVersion string,
) error {
	updates := map[string]any{
		"notes":                 notes,
		"coc_accepted":          cocAccepted,
		"status":                string(domain.InviteStatusPendingReview),
		"coc_accepted_version":  cocVersion,
	}

	if cocAccepted {
		updates["coc_accepted_at"] = time.Now().UTC()
	}

	result := r.db.WithContext(ctx).
		Model(&dbInvite{}).
		Where(whereID, id).
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
		Where(whereID, id).
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

// PurgeExpired removes invitations that expired without ever being submitted.
// Submitted ones are kept for the audit trail.
func (r *inviteRepository) PurgeExpired(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at < ? AND status = ?", before, string(domain.InviteStatusSent)).
		Delete(&dbInvite{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to purge expired invites: %w", result.Error)
	}

	return result.RowsAffected, nil
}
