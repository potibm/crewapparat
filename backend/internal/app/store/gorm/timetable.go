package gorm

import (
	"time"

	"github.com/potibm/crewapparat/internal/app/domain"
)

type dbTimetableEvent struct {
	TimestampModel
	AuditModel

	ID         int64  `gorm:"primaryKey;column:id"`
	Source     string `gorm:"uniqueIndex:tte_idx_ext;column:source"`
	ExternalID string `gorm:"uniqueIndex:tte_idx_ext;column:external_id"`

	Title       string    `gorm:"column:title"`
	Description string    `gorm:"column:description"`
	ExternalURL string    `gorm:"column:external_url"`
	StartTime   time.Time `gorm:"column:start_time"`
	EndTime     time.Time `gorm:"column:end_time"`
	IsHidden    bool      `gorm:"column:is_hidden"`

	CategoryName  string `gorm:"column:category_name"`
	CategoryColor string `gorm:"column:category_color"`

	LocationName    string `gorm:"column:location_name"`
	LocationAddress string `gorm:"column:location_address"`
}

func (dbTimetableEvent) TableName() string { return "timetable_events" }

func fromDomainTimetableEvent(e *domain.TimetableEvent) *dbTimetableEvent {
	db := &dbTimetableEvent{
		ID:          e.ID,
		Source:      e.Source,
		ExternalID:  e.ExternalID,
		Title:       e.Title,
		Description: e.Description,
		ExternalURL: e.ExternalURL,
		StartTime:   e.StartTime,
		EndTime:     e.EndTime,
		IsHidden:    e.IsHidden,
	}
	if e.Category != nil {
		db.CategoryName = e.Category.Name
		db.CategoryColor = e.Category.Color
	}

	if e.Location != nil {
		db.LocationName = e.Location.Name
		db.LocationAddress = e.Location.Address
	}

	return db
}

func toDomainTimetableEvent(db *dbTimetableEvent) *domain.TimetableEvent {
	result := &domain.TimetableEvent{
		ID:          db.ID,
		Source:      db.Source,
		ExternalID:  db.ExternalID,
		Title:       db.Title,
		Description: db.Description,
		ExternalURL: db.ExternalURL,
		StartTime:   db.StartTime,
		EndTime:     db.EndTime,
		IsHidden:    db.IsHidden,
	}
	if db.CategoryName != "" || db.CategoryColor != "" {
		result.Category = &domain.Category{Name: db.CategoryName, Color: db.CategoryColor}
	}

	if db.LocationName != "" || db.LocationAddress != "" {
		result.Location = &domain.Location{Name: db.LocationName, Address: db.LocationAddress}
	}

	return result
}
