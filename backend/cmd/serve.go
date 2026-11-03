package cmd

import (
	"embed"
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
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
				Port:              Cfg.App.Port,
				StaticFiles:       staticFiles,
				Cfg:               Cfg,
				OrganizerRepo:     dbStore.NewOrganizerRepository(),
				EventYearRepo:     dbStore.NewEventYearRepository(),
				TeamRepo:          dbStore.NewTeamRepository(),
				ParticipationRepo: dbStore.NewParticipationRepository(),
				InviteRepo:        dbStore.NewInviteRepository(),
				CodeOfConductRepo: dbStore.NewCodeOfConductRepository(),
			})
			if err != nil {
				return fmt.Errorf("failed to initialize server: %w", err)
			}

			server.RegisterAdminRoutes = func(admin *gin.RouterGroup) {
				// Me
				admin.GET("/me", server.GetMe)
				admin.POST("/me/link", server.LinkMe)

				// Organizers
				admin.GET("/organizers", server.ListOrganizers)
				admin.POST("/organizers", server.CreateOrganizer)
				admin.GET("/organizers/by-slug/:slug", server.GetOrganizerBySlug)
				admin.GET("/organizers/:id", server.GetOrganizer)
				admin.PUT("/organizers/:id", server.UpdateOrganizer)
				admin.DELETE("/organizers/:id", server.DeleteOrganizer)

				// Event Years
				admin.GET("/event-years", server.ListEventYears)
				admin.POST("/event-years", server.CreateEventYear)
				admin.GET("/event-years/:year", server.GetEventYear)
				admin.PUT("/event-years/:year", server.UpdateEventYear)
				admin.DELETE("/event-years/:year", server.DeleteEventYear)

				// Teams
				admin.GET("/teams", server.ListTeams)
				admin.POST("/teams", server.CreateTeam)
				admin.GET("/teams/:id", server.GetTeam)
				admin.PUT("/teams/:id", server.UpdateTeam)
				admin.DELETE("/teams/:id", server.DeleteTeam)

				// Participations
				admin.GET("/participations", server.ListParticipations)
				admin.POST("/participations", server.CreateParticipation)
				admin.GET("/participations/:id", server.GetParticipation)
				admin.PUT("/participations/:id", server.UpdateParticipation)
				admin.DELETE("/participations/:id", server.DeleteParticipation)

				// Invites
				admin.GET("/invites", server.ListInvites)
				admin.POST("/invites", server.CreateInvite)
				admin.GET("/invites/:id", server.GetInvite)
				admin.PUT("/invites/:id", server.UpdateInvite)
				admin.DELETE("/invites/:id", server.DeleteInvite)

				// Code of Conduct
				admin.GET("/code-of-conducts", server.ListCodeOfConducts)
				admin.POST("/code-of-conducts", server.CreateCodeOfConduct)
				admin.GET("/code-of-conducts/:id", server.GetCodeOfConduct)
				admin.PUT("/code-of-conducts/:id", server.UpdateCodeOfConduct)
				admin.DELETE("/code-of-conducts/:id", server.DeleteCodeOfConduct)
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
