package gorm

import (
	"time"

	"gorm.io/gorm"
)

// TimestampModel provides timestamps and soft delete, but deliberately no
// primary key. Models keyed by UUID declare their own `ID` field; models keyed
// by string (EventYear.Year) need none at all. An embedded int64 `ID` would be
// shadowed silently.
type TimestampModel struct {
	CreatedAt time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at;index"`
} // @name models.timestampModel
