package gorm

import (
	"fmt"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/potibm/crewapparat/internal/app/config"
)

type Store struct {
	db *gorm.DB
}

// allModels wird per AutoMigrate in die Datenbank übertragen und von PurgeAll
// wieder entfernt. Hier gehören die GORM-Modelle der Fachlogik rein.
//
// joinTables sind die many2many-Verknüpfungen. GORM legt sie beim
// AutoMigrate automatisch an, sie stehen hier aber explizit, damit PurgeAll sie
// zuverlässig wieder entfernt.
var allModels = []interface{}{
	&dbEventYear{},
	&dbTeam{},
	&dbOrganizer{},
	&dbParticipation{},
	&dbInvite{},
}

var joinTables = []interface{}{
	&participationTeams{},
	&inviteTeams{},
}

// participationTeams bildet participation_teams ab. Die Spalten heißen
// bewusst nicht participation_id/team_id aus GORM-Generierung, sondern werden
// hier fest vergeben, weil die Repositorys die Namen direkt ansprechen.
type participationTeams struct {
	ParticipationID uuid.UUID `gorm:"primaryKey;column:participation_id;type:text"`
	TeamID          uuid.UUID `gorm:"primaryKey;column:team_id;type:text"`
}

func (participationTeams) TableName() string { return "participation_teams" }

type inviteTeams struct {
	InviteID uuid.UUID `gorm:"primaryKey;column:invite_id;type:text"`
	TeamID   uuid.UUID `gorm:"primaryKey;column:team_id;type:text"`
}

func (inviteTeams) TableName() string { return "invite_teams" }

// dropOrder ist die umgekehrte Abhängigkeitsreihenfolge: erst die Tabellen, die
// auf andere verweisen, damit DropTable nicht an Fremdschlüsseln scheitert.
func dropOrder() []interface{} {
	all := append(append([]interface{}{}, allModels...), joinTables...)

	reverse := make([]interface{}, 0, len(all))
	for i := len(all) - 1; i >= 0; i-- {
		reverse = append(reverse, all[i])
	}

	return reverse
}

func NewSqliteStore(filename string) (*Store, error) {
	if filename == "" {
		filename = config.DefaultDBFilename
	}

	dbPath := filepath.Join(config.DataDirname, filename+".db")

	dsn := fmt.Sprintf("%s?_busy_timeout=5000", dbPath)

	return newStore(dsn)
}

func NewSqliteInMemoryStore() (*Store, error) {
	dsn := "file::memory:?cache=shared"

	return newStore(dsn)
}

func (s *Store) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying database connection: %w", err)
	}

	return sqlDB.Close()
}

func (s *Store) PurgeAll() error {
	if err := s.db.Migrator().DropTable(dropOrder()...); err != nil {
		return fmt.Errorf("failed to drop tables: %w", err)
	}

	migrateTargets := append(append([]interface{}{}, allModels...), joinTables...)
	if err := s.db.AutoMigrate(migrateTargets...); err != nil {
		return fmt.Errorf("failed to recreate tables: %w", err)
	}

	return nil
}

func newStore(dsn string) (*Store, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err == nil {
		_, err = sqlDB.Exec("PRAGMA journal_mode = WAL;")
		if err != nil {
			return nil, fmt.Errorf("failed to set journal mode: %w", err)
		}

		_, err = sqlDB.Exec("PRAGMA foreign_keys = ON;")
		if err != nil {
			return nil, fmt.Errorf("failed to set foreign keys: %w", err)
		}
	}

	migrateTargets := append(append([]interface{}{}, allModels...), joinTables...)
	if err := db.AutoMigrate(migrateTargets...); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	if err := RegisterUUIDCallbacks(db); err != nil {
		return nil, fmt.Errorf("failed to register uuid callbacks: %w", err)
	}

	err = RegisterAuditCallbacks(db)
	if err != nil {
		return nil, fmt.Errorf("failed to register audit callbacks: %w", err)
	}

	return &Store{db: db}, nil
}
