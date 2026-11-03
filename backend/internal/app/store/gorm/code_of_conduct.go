package gorm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/potibm/crewapparat/internal/app/domain"
)

type dbCodeOfConduct struct {
	TimestampModel
	AuditModel

	ID          uuid.UUID  `gorm:"primaryKey;type:text;column:id"`
	Version     string     `gorm:"column:version"`
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
		Version:     c.Version,
		Status:      string(c.Status),
		ChangeNote:  c.ChangeNote,
		Text:        c.Text,
		PublishedAt: c.PublishedAt,
	}
}

func toDomainCodeOfConduct(db *dbCodeOfConduct) *domain.CodeOfConduct {
	return &domain.CodeOfConduct{
		ID:          db.ID,
		Version:     db.Version,
		Status:      domain.CodeOfConductStatus(db.Status),
		ChangeNote:  db.ChangeNote,
		Text:        db.Text,
		PublishedAt: db.PublishedAt,
	}
}

// backfillCodeOfConductVersions assigns a version to records that predate the
// version column, so existing installations have a usable overview.
func backfillCodeOfConductVersions(db *gorm.DB) error {
	return db.Model(&dbCodeOfConduct{}).
		Where("version = '' OR version IS NULL").
		Update("version", "1.0.0").Error
}
