package gorm

import (
	"time"

	"github.com/google/uuid"

	"github.com/potibm/crewapparat/internal/app/domain"
)

type dbCodeOfConduct struct {
	TimestampModel
	AuditModel

	ID          uuid.UUID  `gorm:"primaryKey;type:text;column:id"`
	Status      string     `gorm:"column:status"`
	ChangeNote  string     `gorm:"column:change_note"`
	Text        string     `gorm:"column:text"`
	PublishedAt *time.Time `gorm:"column:published_at"`
}

func (dbCodeOfConduct) TableName() string { return "code_of_conducts" }

func (c dbCodeOfConduct) PrimaryKey() uuid.UUID { return c.ID }

func (c *dbCodeOfConduct) SetID(id uuid.UUID) { c.ID = id }

func fromDomainCodeOfConduct(c *domain.CodeOfConduct) *dbCodeOfConduct {
	return &dbCodeOfConduct{
		ID:          c.ID,
		Status:      string(c.Status),
		ChangeNote:  c.ChangeNote,
		Text:        c.Text,
		PublishedAt: c.PublishedAt,
	}
}

func toDomainCodeOfConduct(db *dbCodeOfConduct) *domain.CodeOfConduct {
	return &domain.CodeOfConduct{
		ID:          db.ID,
		Status:      domain.CodeOfConductStatus(db.Status),
		ChangeNote:  db.ChangeNote,
		Text:        db.Text,
		PublishedAt: db.PublishedAt,
	}
}
