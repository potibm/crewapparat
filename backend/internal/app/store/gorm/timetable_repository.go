package gorm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/potibm/crewapparat/internal/app/domain"
	"github.com/potibm/crewapparat/internal/app/repository"
	"gorm.io/gorm"
)

type timetableRepository struct {
	db *gorm.DB
}

func (s *Store) NewTimetableRepository() repository.TimetableRepository {
	return NewTimetableRepository(s.db)
}

func NewTimetableRepository(db *gorm.DB) repository.TimetableRepository {
	return &timetableRepository{db: db}
}

func (r *timetableRepository) ListByRange(ctx context.Context, start, end time.Time) ([]domain.TimetableEvent, error) {
	var rows []dbTimetableEvent
	if err := r.db.WithContext(ctx).
		Where("start_time < ? AND end_time > ?", end, start).
		Order("start_time ASC").
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to list timetable events: %w", err)
	}

	result := make([]domain.TimetableEvent, 0, len(rows))
	for i := range rows {
		result = append(result, *toDomainTimetableEvent(&rows[i]))
	}

	return result, nil
}

func (r *timetableRepository) Save(ctx context.Context, event *domain.TimetableEvent) error {
	var existing dbTimetableEvent

	err := r.db.WithContext(ctx).
		Where("source = ? AND external_id = ?", event.Source, event.ExternalID).
		First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to load timetable event: %w", err)
	}

	dbObj := fromDomainTimetableEvent(event)
	if err == nil {
		dbObj.ID = existing.ID
		dbObj.CreatedAt = existing.CreatedAt

		dbObj.DeletedAt = existing.DeletedAt
		if err := r.db.WithContext(ctx).Save(dbObj).Error; err != nil {
			return fmt.Errorf("failed to update timetable event: %w", err)
		}

		event.ID = dbObj.ID

		return nil
	}

	if err := r.db.WithContext(ctx).Create(dbObj).Error; err != nil {
		return fmt.Errorf("failed to create timetable event: %w", err)
	}

	event.ID = dbObj.ID

	return nil
}

func (r *timetableRepository) Sync(
	ctx context.Context,
	source string,
	events []domain.TimetableEvent,
) (created, updated, deleted int, err error) {
	var existing []dbTimetableEvent
	if err := r.db.WithContext(ctx).Where("source = ?", source).Find(&existing).Error; err != nil {
		return 0, 0, 0, fmt.Errorf("failed to list existing timetable events: %w", err)
	}

	incoming := make(map[string]domain.TimetableEvent, len(events))
	for _, e := range events {
		incoming[e.ExternalID] = e
	}

	for _, e := range events {
		ev := e

		var row dbTimetableEvent

		err = r.db.WithContext(ctx).Where("source = ? AND external_id = ?", source, e.ExternalID).First(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := r.Save(ctx, &ev); err != nil {
				return created, updated, 0, err
			}

			created++

			continue
		}

		if err != nil {
			return created, updated, 0, fmt.Errorf("failed to load timetable event: %w", err)
		}

		if err := r.Save(ctx, &ev); err != nil {
			return created, updated, 0, err
		}

		updated++
	}

	for _, row := range existing {
		if _, ok := incoming[row.ExternalID]; ok {
			continue
		}

		if err := r.db.WithContext(ctx).
			Delete(&dbTimetableEvent{}, "source = ? AND external_id = ?", source, row.ExternalID).
			Error; err != nil {
			return created, updated, deleted, fmt.Errorf("failed to delete timetable event: %w", err)
		}

		deleted++
	}

	return created, updated, deleted, nil
}

func (r *timetableRepository) Delete(ctx context.Context, source, externalID string) error {
	result := r.db.WithContext(ctx).Delete(&dbTimetableEvent{}, "source = ? AND external_id = ?", source, externalID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete timetable event: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("delete timetable event %s/%s: %w", source, externalID, gorm.ErrRecordNotFound)
	}

	return nil
}
