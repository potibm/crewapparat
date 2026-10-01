package cmd

import (
	"embed"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/potibm/crewapparat/internal/app/config"
	"github.com/potibm/crewapparat/internal/app/hub"
	"github.com/potibm/crewapparat/internal/app/initializer"
	store "github.com/potibm/crewapparat/internal/app/store/gorm"
)

//go:embed assets
var staticFiles embed.FS

const (
	otelEndpointFlagName = "otel-endpoint"
	portFlagName         = "port"
)

var (
	port         int
	otelEndpoint string
)

func NewServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Runs the HTTP server for the Crewapparat application",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ensureAppInfrastructure()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := slog.Default()

			// =========================================================================
			// PHASE 1: Observability & Telemetry
			// =========================================================================
			shutdownFn, err := initializer.InitTelemetry(ctx, Cfg.App.OtelEndpoint, Cfg.App.Version)
			if err != nil {
				return fmt.Errorf("failed to initialize telemetry: %w", err)
			}

			if shutdownFn != nil {
				defer shutdownFn()
			}

			initializer.InitializeSentry(Cfg.Sentry)

			// =========================================================================
			// PHASE 2: Core Infrastructure (DB & Redis)
			// =========================================================================
			dbStore, err := store.NewSqliteStore(Cfg.App.DbFilename)
			if err != nil {
				return fmt.Errorf("database error: %w", err)
			}

			defer func() {
				if err := dbStore.Close(); err != nil {
					logger.Error("failed to close database", "error", err)
				}
			}()

			initializer.InitializeRedis(Cfg.App.RedisURL)

			// =========================================================================
			// PHASE 3: API Server
			// =========================================================================

			server, err := hub.NewServer(hub.Config{
				Port:        Cfg.App.Port,
				StaticFiles: staticFiles,
				Cfg:         Cfg,
			})
			if err != nil {
				return fmt.Errorf("failed to initialize server: %w", err)
			}

			return server.Run(ctx)
		},
	}

	cmd.Flags().IntVarP(&port, portFlagName, "p", config.DefaultPort, "Set the port number for the server to listen on")
	_ = viper.BindPFlag("app.port", cmd.Flags().Lookup(portFlagName))

	cmd.Flags().
		StringVar(&otelEndpoint, otelEndpointFlagName, "", "Set the OpenTelemetry endpoint (e.g., localhost:4317)")
	_ = viper.BindPFlag("app.otel_endpoint", cmd.Flags().Lookup(otelEndpointFlagName))

	return cmd
}
