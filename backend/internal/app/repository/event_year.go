package repository

import (
	"context"

	"github.com/potibm/crewapparat/internal/app/domain"
)

type EventYearListFilters struct {
	Query *string
}

type EventYearListParams = ListParams

type EventYearRepository interface {
	Create(ctx context.Context, year *domain.EventYear) error
	Update(ctx context.Context, year *domain.EventYear) error
	Delete(ctx context.Context, year string) error
	List(
		ctx context.Context,
		params EventYearListParams,
		filters EventYearListFilters,
	) ([]domain.EventYear, int64, error)
	GetByYear(ctx context.Context, year string) (*domain.EventYear, error)
	GetActive(ctx context.Context) (*domain.EventYear, error)
	// ActivateYear setzt genau ein Jahr aktiv und deaktiviert alle anderen.
	// SQLite unterstützt keine partiellen Unique-Indizes, deshalb wird die
	// Eindeutigkeit hier in einer Transaktion sichergestellt.
	ActivateYear(ctx context.Context, year string) error
}
