package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/potibm/crewapparat/internal/app/config"
	"github.com/potibm/crewapparat/internal/app/domain"
	"github.com/potibm/crewapparat/internal/app/repository"
	store "github.com/potibm/crewapparat/internal/app/store/gorm"
)

func NewMediaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "media",
		Short: "Media file commands",
		Annotations: map[string]string{
			skipConfigValidationAnnotation: "true",
		},
	}
	cmd.AddCommand(NewMediaCleanupCmd())

	return cmd
}

func NewMediaCleanupCmd() *cobra.Command {
	var (
		mediaPath string
		grace     string
		dryRun    bool
	)

	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Remove media files that are no longer referenced by any organizer",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ensureAppInfrastructure()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			graceDuration, err := time.ParseDuration(grace)
			if err != nil {
				return fmt.Errorf("invalid grace duration %q: %w", grace, err)
			}

			if mediaPath == "" {
				mediaPath = filepath.Join(config.DataDirname, "media")
			}

			return runMediaCleanup(cmd, mediaPath, graceDuration, dryRun)
		},
	}

	cmd.Flags().StringVar(&mediaPath, "path", "", "Media directory to clean up")
	cmd.Flags().StringVar(&grace, "grace", "24h", "Only delete files older than this duration")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be deleted without deleting")

	return cmd
}

func runMediaCleanup(cmd *cobra.Command, mediaPath string, graceDuration time.Duration, dryRun bool) error {
	dbStore, err := store.NewSqliteStore(Cfg.App.DbFilename)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() {
		if closeErr := dbStore.Close(); closeErr != nil {
			slog.Error("failed to close database connection", "error", closeErr)
		}
	}()

	organizers, _, err := dbStore.NewOrganizerRepository().List(
		cmd.Context(),
		repository.OrganizerListParams{},
		repository.OrganizerListFilters{},
	)
	if err != nil {
		return fmt.Errorf("failed to list organizers: %w", err)
	}

	referenced := collectReferencedMedia(organizers)

	entries, err := os.ReadDir(mediaPath)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("media directory does not exist, nothing to do", "path", mediaPath)

			return nil
		}

		return fmt.Errorf("failed to read media directory %s: %w", mediaPath, err)
	}

	cutoff := time.Now().Add(-graceDuration)
	deleted := 0
	skipped := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			slog.Warn("failed to stat media file", "file", entry.Name(), "error", err)

			continue
		}

		if info.ModTime().After(cutoff) {
			skipped++

			continue
		}

		if _, ok := referenced[entry.Name()]; ok {
			continue
		}

		fullPath := filepath.Join(mediaPath, entry.Name())
		if dryRun {
			slog.Info("would delete media file", "file", fullPath)

			deleted++

			continue
		}

		if err := os.Remove(fullPath); err != nil {
			slog.Warn("failed to delete media file", "file", fullPath, "error", err)

			continue
		}

		slog.Info("deleted media file", "file", fullPath)

		deleted++
	}

	slog.Info("media cleanup completed",
		"deleted", deleted,
		"skipped_recent", skipped,
		"dry_run", dryRun,
	)

	return nil
}

func localMediaFilename(avatarURL string) string {
	avatarURL = strings.TrimSpace(avatarURL)
	if !strings.HasPrefix(avatarURL, "/media/") {
		return ""
	}

	return filepath.Base(avatarURL)
}

func collectReferencedMedia(organizers []domain.Organizer) map[string]struct{} {
	referenced := map[string]struct{}{}

	for i := range organizers {
		name := localMediaFilename(organizers[i].AvatarURL)
		if name != "" {
			referenced[name] = struct{}{}
		}
	}

	return referenced
}
