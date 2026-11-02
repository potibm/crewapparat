package gorm

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// joinBatchSize is the chunk size for the many2many join tables.
const joinBatchSize = 100

// idGenerator is implemented by db models with a UUID primary key.
type idGenerator interface {
	PrimaryKey() uuid.UUID
	SetID(id uuid.UUID)
}

// RegisterUUIDCallbacks installs the BeforeCreate callback that assigns missing
// UUIDs.
//
// This replaces a `DEFAULT uuid_generate_v4()` in the column definition: that
// function only exists in Postgres, SQLite does not know it and aborts the
// migration with a syntax error. Assigning the ID in Go instead keeps it
// database independent.
func RegisterUUIDCallbacks(db *gorm.DB) error {
	err := db.Callback().Create().Before("gorm:create").Register("crewapparat:uuid", assignUUID)
	if err != nil {
		return fmt.Errorf("failed to register uuid callback: %w", err)
	}

	return nil
}

func assignUUID(tx *gorm.DB) {
	gen, ok := tx.Statement.Dest.(idGenerator)
	if !ok {
		return
	}

	if gen.PrimaryKey() == uuid.Nil {
		gen.SetID(uuid.New())
	}
}
