package gorm

import (
	"github.com/google/uuid"

	"github.com/potibm/crewapparat/internal/app/domain"
)

type dbTeam struct {
	TimestampModel
	AuditModel

	ID        uuid.UUID `gorm:"primaryKey;type:text;column:id"`
	Name      string    `gorm:"uniqueIndex;column:name"`
	Color     string    `gorm:"column:color"`
	HasShifts bool      `gorm:"default:false;column:has_shifts"`
}

func (dbTeam) TableName() string { return "teams" }

func (t dbTeam) PrimaryKey() uuid.UUID { return t.ID }

func (t *dbTeam) SetID(id uuid.UUID) { t.ID = id }

func fromDomainTeam(t *domain.Team) *dbTeam {
	return &dbTeam{
		ID:        t.ID,
		Name:      t.Name,
		Color:     t.Color,
		HasShifts: t.HasShifts,
	}
}

func toDomainTeam(db *dbTeam) *domain.Team {
	return &domain.Team{
		ID:        db.ID,
		Name:      db.Name,
		Color:     db.Color,
		HasShifts: db.HasShifts,
	}
}
