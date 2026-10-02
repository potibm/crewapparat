package gorm

import (
	"github.com/potibm/crewapparat/internal/app/domain"
)

// dbEventYear hat bewusst kein ID-Feld: der Primärschlüssel ist die Jahrzahl
// als String.
type dbEventYear struct {
	TimestampModel
	AuditModel

	Year     string `gorm:"primaryKey;type:text;column:year"`
	IsActive bool   `gorm:"index;column:is_active"`
}

func (dbEventYear) TableName() string { return "event_years" }

func fromDomainEventYear(y *domain.EventYear) *dbEventYear {
	return &dbEventYear{
		Year:     y.Year,
		IsActive: y.IsActive,
	}
}

func toDomainEventYear(db *dbEventYear) *domain.EventYear {
	return &domain.EventYear{
		Year:     db.Year,
		IsActive: db.IsActive,
	}
}
