package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/potibm/crewapparat/internal/app/domain"
)

type InviteListFilters struct {
	Query  *string
	ID     *uuid.UUID
	Email  *string
	Status *domain.InviteStatus
}

type InviteListParams = ListParams

// InviteRepository ist das Lesen und Schreiben von Einladungen im Admin.
type InviteRepository interface {
	Create(ctx context.Context, invite *domain.Invite) error
	Update(ctx context.Context, invite *domain.Invite) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, params InviteListParams, filters InviteListFilters) ([]domain.Invite, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Invite, error)
	GetByEmail(ctx context.Context, email string) (*domain.Invite, error)
}

// InviteTokenRepository kapselt den öffentlichen Magic-Link-Zugriff. Der Token
// wird nie im Klartext gespeichert, gesucht wird ausschließlich über dessen
// SHA-256-Hash.
type InviteTokenRepository interface {
	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Invite, error)
}

// InviteWorkflowRepository ist der Onboarding-Ablauf: Einreichen durch den
// Eingeladenen, Review durch einen Hauptorganizer, Ablauf-Räumung. Bewusst vom
// reinen CRUD getrennt, weil dieser Pfad unauthentifiziert erreicht wird.
type InviteWorkflowRepository interface {
	SubmitForReview(
		ctx context.Context,
		id uuid.UUID,
		providedName, providedPhone, notes string,
		cocAccepted bool,
	) error
	Review(ctx context.Context, id uuid.UUID, status domain.InviteStatus, reviewerID uuid.UUID) error
	EmailExists(ctx context.Context, email string) (bool, error)
	PurgeExpired(ctx context.Context, before time.Time) (int64, error)
}

// InviteTeamRepository ist die Zuordnung der beantragten Teams.
type InviteTeamRepository interface {
	SetTeams(ctx context.Context, inviteID uuid.UUID, teams []domain.Team) error
}
