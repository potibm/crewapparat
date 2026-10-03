package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/potibm/crewapparat/internal/app/seeder"
	store "github.com/potibm/crewapparat/internal/app/store/gorm"
)

var (
	seedOrganizerCount int
	seedRNGValue       int64
)

func NewDatabaseSeedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seed",
		Short: "Fill the database with initial data",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ensureAppInfrastructure()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			slog.Info("Running database seed...",
				"file", Cfg.App.DbFilename,
				"organizers", seedOrganizerCount,
				"rng_seed", seedRNGValue,
			)

			dbStore, err := store.NewSqliteStore(Cfg.App.DbFilename)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}

			defer func() {
				if closeErr := dbStore.Close(); closeErr != nil {
					slog.Error("failed to close database connection", "error", closeErr)
				}
			}()

			seed := seeder.New(
				seeder.Repositories{
					Organizers:     dbStore.NewOrganizerRepository(),
					Teams:          dbStore.NewTeamRepository(),
					EventYears:     dbStore.NewEventYearRepository(),
					Participations: dbStore.NewParticipationRepository(),
				},
				seeder.WithOrganizerCount(seedOrganizerCount),
				seeder.WithSeed(seedRNGValue),
			)

			if err := seed.Run(ctx); err != nil {
				return fmt.Errorf("database seed failed: %w", err)
			}

			slog.Info("Database seed completed successfully!")

			return nil
		},
	}

	cmd.Flags().IntVar(&seedOrganizerCount, "organizers", seeder.DefaultOrganizerCount,
		"Number of random organizers to generate on top of the two Dex accounts")
	cmd.Flags().Int64Var(&seedRNGValue, "rng-seed", 1,
		"RNG seed for reproducible data")

	return cmd
}
