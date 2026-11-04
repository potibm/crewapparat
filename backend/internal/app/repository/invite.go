package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/potibm/crewapparat/internal/app/domain"
)

type InviteListFilters struct {
	Query       *string
	ID          *uuid.UUID
	Email       *string
	Status      *domain.InviteStatus
	InvitedByID *uuid.UUID
}

type InviteListParams = ListParams

// InviteRepository covers reading and writing invitations in the admin.
type InviteRepository interface {
	Create(ctx context.Context, invite *domain.Invite) error
	Update(ctx context.Context, invite *domain.Invite) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, params InviteListParams, filters InviteListFilters) ([]domain.Invite, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Invite, error)
	GetByEmail(ctx context.Context, email string) (*domain.Invite, error)
	// GetByToken resolves a magic link back to its invitation.
	GetByToken(ctx context.Context, token string) (*domain.Invite, error)
}

// InviteWorkflowRepository is the onboarding flow: submission by the invitee,
// review by a lead organizer, expiry cleanup. Kept apart from plain CRUD
// because this path is reached without authentication.
type InviteWorkflowRepository interface {
	SubmitForReview(
		ctx context.Context,
		id uuid.UUID,
		providedName, providedPhone, notes string,
		cocAccepted bool,
	) error
	Review(ctx context.Context, id uuid.UUID, status domain.InviteStatus, reviewerID uuid.UUID) error
	EmailExists(ctx context.Context, email string) (bool, error)
	PurgeExpired(ctx context.Context, before time.Time) (int64, error)
}
