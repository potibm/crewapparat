package gorm

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/potibm/crewapparat/internal/app/config"
)

// slowQueryThreshold is when GORM starts logging a query as slow.
const slowQueryThreshold = 200 * time.Millisecond

type Store struct {
	db *gorm.DB
}

// allModels is migrated into the database by AutoMigrate and removed again by
// PurgeAll. The GORM models of the domain logic belong here.
//
// joinTables are the many2many join tables. GORM creates them during
// AutoMigrate, but they are listed explicitly so PurgeAll removes them reliably.
var allModels = []interface{}{
	&dbEventYear{},
	&dbTeam{},
	&dbOrganizer{},
	&dbParticipation{},
	&dbInvite{},
	&dbCodeOfConduct{},
	&dbCodeOfConductAcceptance{},
	&dbTimetableEvent{},
}

var joinTables = []interface{}{
	&participationTeams{},
}

// participationTeams maps participation_teams. The column names are assigned
// here rather than left to GORM's generator because the repositories address
// them directly.
type participationTeams struct {
	ParticipationID uuid.UUID `gorm:"primaryKey;column:participation_id;type:text"`
	TeamID          uuid.UUID `gorm:"primaryKey;column:team_id;type:text"`
}

func (participationTeams) TableName() string { return "participation_teams" }

// dropOrder is the reverse dependency order: tables referencing others go first
// so DropTable does not fail on foreign keys.
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

// slogWriter routes GORM log output through slog, so database messages end up in
// the same stream and format as everything else.
type slogWriter struct{}

func (slogWriter) Printf(format string, args ...any) {
	slog.Warn(strings.TrimSpace(fmt.Sprintf(format, args...)))
}

// newLogger configures GORM's own logger.
//
// IgnoreRecordNotFoundError keeps "record not found" out of the log: a lookup
// that finds nothing is a normal outcome, and callers check for it explicitly
// via IsNotFound. ParameterizedQueries keeps actual values - emails, tokens -
// out of the log entirely.
func newLogger() logger.Interface {
	return logger.New(slogWriter{}, logger.Config{
		SlowThreshold:             slowQueryThreshold,
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: true,
		ParameterizedQueries:      true,
		Colorful:                  false,
	})
}

func newStore(dsn string) (*Store, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: newLogger()})
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

	// Partial unique index: empty slugs (not yet backfilled) must not collide.
	if err := db.Exec(
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_organizers_slug ON organizers(slug) WHERE slug <> ''",
	).Error; err != nil {
		return nil, fmt.Errorf("failed to create slug index: %w", err)
	}

	if err := RegisterUUIDCallbacks(db); err != nil {
		return nil, fmt.Errorf("failed to register uuid callbacks: %w", err)
	}

	err = RegisterAuditCallbacks(db)
	if err != nil {
		return nil, fmt.Errorf("failed to register audit callbacks: %w", err)
	}

	// Backfill immutable slugs for organizers that predate the column.
	if err := backfillOrganizerSlugs(db); err != nil {
		return nil, fmt.Errorf("failed to backfill organizer slugs: %w", err)
	}

	// Backfill versions for code of conduct records that predate the column.
	if err := backfillCodeOfConductVersions(db); err != nil {
		return nil, fmt.Errorf("failed to backfill code of conduct versions: %w", err)
	}

	return &Store{db: db}, nil
}
