package cmd

import (
	"fmt"
	"log/slog"

	store "github.com/potibm/crewapparat/internal/app/store/gorm"
	"github.com/spf13/cobra"
)

var resetForce bool

func NewDatabaseResetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset the database, dropping and recreating all tables",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ensureAppInfrastructure()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !resetForce && !confirmReset(Cfg.App.DbFilename) {
				slog.Info("Reset aborted by user")

				return nil
			}

			if err := performDatabaseReset(Cfg.App.DbFilename); err != nil {
				slog.Error("Database reset failed", "error", err)

				return fmt.Errorf("database reset failed: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&resetForce, "force", "f", false, "Skips the confirmation prompt (for CI/CD)")

	return cmd
}

func confirmReset(dbName string) bool {
	return confirm(fmt.Sprintf("WARNING: The database '%s' and all its data will be COMPLETELY deleted!", dbName))
}

func performDatabaseReset(dbName string) error {
	slog.Info("Performing database reset...")

	dbStore, err := store.NewSqliteStore(dbName)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	defer func() {
		if err := dbStore.Close(); err != nil {
			slog.Error("failed to close database", "error", err)
		}
	}()

	if err := dbStore.PurgeAll(); err != nil {
		return fmt.Errorf("failed to purge database: %w", err)
	}

	slog.Info("Database reset completed successfully!")

	return nil
}
