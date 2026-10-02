package gorm

import (
	"time"

	"gorm.io/gorm"
)

// TimestampModel liefert Zeitstempel und Soft Delete, aber bewusst keinen
// Primärschlüssel. Modelle mit UUID als PK deklarieren ihr eigenes `ID`-Feld,
// Modelle mit String-PK (z. B. EventYear.Year) brauchen gar keinen. Ein
// eingebettetes int64-`ID` würde dabei still überschrieben.
type TimestampModel struct {
	CreatedAt time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at;index"`
} // @name models.timestampModel
