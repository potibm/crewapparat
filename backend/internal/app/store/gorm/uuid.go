package gorm

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// joinBatchSize ist die Chunk-Größe für die many2many-Join-Tabellen.
const joinBatchSize = 100

// idGenerator wird von db-Modellen mit UUID-Primärschlüssel implementiert.
type idGenerator interface {
	PrimaryKey() uuid.UUID
	SetID(id uuid.UUID)
}

// RegisterUUIDCallbacks installiert den BeforeCreate-Callback, der fehlende
// UUIDs vergibt.
//
// Das ersetzt ein `DEFAULT uuid_generate_v4()` in der Spaltendefinition:
// diese Funktion existiert nur in Postgres, SQLite kennt sie nicht und
// bricht die Migration mit einem Syntaxfehler ab. Stattdessen wird die ID
// beim Anlegen einer Zeile in Go gesetzt, was zudem Datenbank-unabhängig bleibt.
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
