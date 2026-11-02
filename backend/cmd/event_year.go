package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/potibm/crewapparat/internal/app/domain"
	"github.com/potibm/crewapparat/internal/app/repository"
	store "github.com/potibm/crewapparat/internal/app/store/gorm"
)

func NewEventYearCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "event-year",
		Short: "Manage event years (create, delete, activate)",
		Annotations: map[string]string{
			skipConfigValidationAnnotation: "true",
		},
	}

	cmd.AddCommand(
		NewEventYearCreateCmd(),
		NewEventYearDeleteCmd(),
		NewEventYearActivateCmd(),
	)

	return cmd
}

func NewEventYearCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <year>",
		Short: "Create a new event year and pending participations for all organizers",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ensureAppInfrastructure()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			year := args[0]

			dbStore, err := store.NewSqliteStore(Cfg.App.DbFilename)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer func() {
				if closeErr := dbStore.Close(); closeErr != nil {
					slog.Error("failed to close database connection", "error", closeErr)
				}
			}()

			eventYearRepo := dbStore.NewEventYearRepository()
			organizerRepo := dbStore.NewOrganizerRepository()
			participationRepo := dbStore.NewParticipationRepository()

			// Check if year already exists.
			if _, err := eventYearRepo.GetByYear(ctx, year); err == nil {
				return fmt.Errorf("event year %s already exists", year)
			}

			eventYear := &domain.EventYear{Year: year, IsActive: false}
			if err := eventYearRepo.Create(ctx, eventYear); err != nil {
				return fmt.Errorf("failed to create event year %s: %w", year, err)
			}

			// Create a pending participation for every organizer.
			organizers, _, err := organizerRepo.List(ctx, repository.ListParams{}, repository.OrganizerListFilters{})
			if err != nil {
				return fmt.Errorf("failed to list organizers: %w", err)
			}

			created := 0

			for i := range organizers {
				participation := &domain.Participation{
					OrganizerID: organizers[i].ID,
					EventYearID: year,
					Status:      domain.ParticipationStatusPending,
				}
				if err := participationRepo.Create(ctx, participation); err != nil {
					slog.Warn("failed to create participation",
						"organizer_id", organizers[i].ID,
						"error", err,
					)

					continue
				}

				created++
			}

			slog.Info("Event year created successfully",
				"year", year,
				"organizers", len(organizers),
				"participations_created", created,
			)

			return nil
		},
	}

	return cmd
}

func NewEventYearDeleteCmd() *cobra.Command {
	force := false

	cmd := &cobra.Command{
		Use:   "delete <year>",
		Short: "Delete an event year and all its participations",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ensureAppInfrastructure()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			year := args[0]

			if !force {
				if !confirm(
					fmt.Sprintf(
						"This will delete event year %s and ALL its participations. This cannot be undone.",
						year,
					),
				) {
					slog.Info("Aborted.")

					return nil
				}
			}

			dbStore, err := store.NewSqliteStore(Cfg.App.DbFilename)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer func() {
				if closeErr := dbStore.Close(); closeErr != nil {
					slog.Error("failed to close database connection", "error", closeErr)
				}
			}()

			eventYearRepo := dbStore.NewEventYearRepository()
			participationRepo := dbStore.NewParticipationRepository()

			// Delete all participations for this year first.
			participations, _, err := participationRepo.List(
				ctx,
				repository.ParticipationListParams{},
				repository.ParticipationListFilters{EventYearID: &year},
			)
			if err != nil {
				return fmt.Errorf("failed to list participations: %w", err)
			}

			deleted := 0

			for i := range participations {
				if delErr := participationRepo.Delete(ctx, participations[i].ID); delErr != nil {
					slog.Warn("failed to delete participation",
						"participation_id", participations[i].ID,
						"error", delErr,
					)

					continue
				}

				deleted++
			}

			if err := eventYearRepo.Delete(ctx, year); err != nil {
				return fmt.Errorf("failed to delete event year %s: %w", year, err)
			}

			slog.Info("Event year deleted successfully",
				"year", year,
				"participations_deleted", deleted,
			)

			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")

	return cmd
}

func NewEventYearActivateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate <year>",
		Short: "Activate an event year (deactivates all others)",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ensureAppInfrastructure()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			year := args[0]

			dbStore, err := store.NewSqliteStore(Cfg.App.DbFilename)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer func() {
				if closeErr := dbStore.Close(); closeErr != nil {
					slog.Error("failed to close database connection", "error", closeErr)
				}
			}()

			eventYearRepo := dbStore.NewEventYearRepository()

			if _, err := eventYearRepo.GetByYear(ctx, year); err != nil {
				return fmt.Errorf("event year %s not found", year)
			}

			if err := eventYearRepo.ActivateYear(ctx, year); err != nil {
				return fmt.Errorf("failed to activate event year %s: %w", year, err)
			}

			slog.Info("Event year activated successfully", "year", year)

			return nil
		},
	}

	return cmd
}
