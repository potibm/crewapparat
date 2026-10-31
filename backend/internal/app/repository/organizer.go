package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/potibm/crewapparat/internal/app/domain"
)

// ListParams and ListFilters mirror the React Admin protocol: endpoints read
// `?_start&_end&_sort&_order` and `?q=`.
type ListParams struct {
	Offset int
	Limit  int
	Sort   string
	Order  string
}

type OrganizerListParams = ListParams

type OrganizerListFilters struct {
	Query *string
	ID    *uuid.UUID
	Slug  *string
}

type OrganizerRepository interface {
	Create(ctx context.Context, organizer *domain.Organizer) error
	Update(ctx context.Context, organizer *domain.Organizer) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(
		ctx context.Context,
		params OrganizerListParams,
		filters OrganizerListFilters,
	) ([]domain.Organizer, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Organizer, error)
	// GetByOidcSub finds the organizer by OIDC `sub`. Returns
	// gorm.ErrRecordNotFound while the account is not linked yet.
	GetByOidcSub(ctx context.Context, oidcSub string) (*domain.Organizer, error)
	// GetByEmail finds an organizer by email address.
	GetByEmail(ctx context.Context, email string) (*domain.Organizer, error)
	// GetBySlug finds an organizer by its immutable slug.
	GetBySlug(ctx context.Context, slug string) (*domain.Organizer, error)
}
