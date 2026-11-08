package gorm

import (
	"time"

	"github.com/google/uuid"

	"github.com/potibm/crewapparat/internal/app/domain"
)

type dbShiftPlan struct {
	TimestampModel
	AuditModel

	ID          uuid.UUID              `gorm:"primaryKey;type:text;column:id"`
	EventYearID string                 `gorm:"column:event_year_id"`
	Version     string                 `gorm:"column:version"`
	Status      string                 `gorm:"column:status"`
	GeneratedAt *time.Time             `gorm:"column:generated_at"`
	Slots       []domain.ShiftPlanSlot `gorm:"serializer:json;column:slots"`
}

func (dbShiftPlan) TableName() string { return "shift_plans" }

func (c dbShiftPlan) PrimaryKey() uuid.UUID { return c.ID }

func (c *dbShiftPlan) SetID(id uuid.UUID) { c.ID = id }

func fromDomainShiftPlan(p *domain.ShiftPlan) *dbShiftPlan {
	return &dbShiftPlan{
		ID:          p.ID,
		EventYearID: p.EventYearID,
		Version:     p.Version,
		Status:      string(p.Status),
		GeneratedAt: p.GeneratedAt,
		Slots:       p.Slots,
	}
}

func toDomainShiftPlan(db *dbShiftPlan) *domain.ShiftPlan {
	return &domain.ShiftPlan{
		ID:          db.ID,
		EventYearID: db.EventYearID,
		Version:     db.Version,
		Status:      domain.ShiftPlanStatus(db.Status),
		GeneratedAt: db.GeneratedAt,
		Slots:       db.Slots,
	}
}
