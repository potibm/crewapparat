package repository

import (
	"context"
	"time"

	"github.com/potibm/crewapparat/internal/app/domain"
)

type TimetableRepository interface {
	ListByRange(ctx context.Context, start, end time.Time) ([]domain.TimetableEvent, error)
	Save(ctx context.Context, event *domain.TimetableEvent) error
	Sync(
		ctx context.Context,
		source string,
		events []domain.TimetableEvent,
	) (created int, updated int, deleted int, err error)
	Delete(ctx context.Context, source, externalID string) error
}
