package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/potibm/crewapparat/internal/app/domain"
)

// ListParams und ListFilters bilden das React-Admin-Protokoll ab: die
// Endpunkte lesen `?_start&_end&_sort&_order` und `?q=`.
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
	// GetByOidcSub findet den Organizer über den OIDC-`sub`. Liefert
	// gorm.ErrRecordNotFound, solange das Konto noch nicht verknüpft ist.
	GetByOidcSub(ctx context.Context, oidcSub string) (*domain.Organizer, error)
	// GetByEmail findet einen Organizer über die E-Mail-Adresse.
	GetByEmail(ctx context.Context, email string) (*domain.Organizer, error)
}
