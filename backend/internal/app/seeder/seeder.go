package seeder

import (
	"context"
	"fmt"
	"hash/fnv"
	"log/slog"
	"slices"
	"strings"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"

	"github.com/potibm/crewapparat/internal/app/domain"
	"github.com/potibm/crewapparat/internal/app/repository"
	"github.com/potibm/crewapparat/internal/app/store/gorm"
)

// Seed sizing. These are hand-tuned rather than derived, so the seeded dataset
// looks plausible in the admin panel.
const (
	// DefaultOrganizerCount is how many random organizers are generated on top
	// of the two Dex accounts.
	DefaultOrganizerCount = 25

	// partyDays and shiftBlocksPerDay describe the shift grid. Block 0 is 10:00
	// on the first day, so block 32 is 10:00 on the second day - that is what
	// makes the example [0, 1, 2, 32, 33] from CONCEPT.md readable.
	partyDays         = 3
	shiftBlocksPerDay = 32

	// minShiftBlocks and maxShiftBlocks bound how many blocks an organizer marks
	// as available or unavailable.
	minShiftBlocks = 2
	maxShiftBlocks = 10

	// maxTeamsPerParticipation keeps most people in one or two teams.
	maxTeamsPerParticipation = 3

	// A past participation exists for roughly this share of organizers, so the
	// annual check-in has a previous year to prefill from.
	previousYearRatio = 0.6

	// Status distribution, in percent. Most crew members have already checked
	// in; a minority is still pending or declined the year.
	statusConfirmedPercent = 60
	statusPendingPercent   = 30
)

// StatusDeclinedPercent is implied by the other two and exists so the split adds
// up to 100.
const StatusDeclinedPercent = 100 - statusConfirmedPercent - statusPendingPercent

var (
	// teamNames are placeholders until the real crew structure is known.
	teamNames = []string{
		"Infodesk",
		"Compol",
		"Halles",
		"Kasse",
		"Netzwerk",
		"Druck",
		"Shuttle",
		"Cleanup",
		"Social",
		"Beamer",
	}

	shirtSizes = []string{"XXS", "XS", "S", "M", "L", "XL", "XXL", "3XL"}

	pronounsOptions = []string{"", "she/her", "he/him", "they/them"}

	languagePool = []string{"de", "en", "da", "nl", "fr", "sv"}
)

// Repositories bundles everything the seeder writes to.
type Repositories struct {
	Organizers     repository.OrganizerRepository
	Teams          repository.TeamRepository
	EventYears     repository.EventYearRepository
	Participations repository.ParticipationRepository
}

// Seeder fills the database with a plausible crew.
//
// Run is idempotent: organizers are updated by email, teams and event years are
// only created when missing, and participations only where none exists yet.
// Running it twice must not fail on a unique constraint.
type Seeder struct {
	repos          Repositories
	organizerCount int
	rngSeed        int64
	handleSuffix   int
}

// Option customises a Seeder.
type Option func(*Seeder)

// WithOrganizerCount overrides how many random organizers are generated.
func WithOrganizerCount(count int) Option {
	return func(s *Seeder) {
		if count >= 0 {
			s.organizerCount = count
		}
	}
}

// WithSeed fixes the RNG seed so two runs produce identical data.
func WithSeed(seed int64) Option {
	return func(s *Seeder) {
		s.rngSeed = seed
	}
}

// New builds a Seeder.
func New(repos Repositories, opts ...Option) *Seeder {
	s := &Seeder{
		repos:          repos,
		organizerCount: DefaultOrganizerCount,
		rngSeed:        1,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Run generates the seed data.
//
// Every step aborts on the first error. That matters: an organizer whose Create
// failed has an empty ID, and every participation seeded after it would point at
// a record that does not exist.
func (s *Seeder) Run(ctx context.Context) error {
	slog.Info("Starting database seed...")

	if err := gofakeit.Seed(s.rngSeed); err != nil {
		return fmt.Errorf("failed to seed rng: %w", err)
	}

	teams, err := s.seedTeams(ctx)
	if err != nil {
		return err
	}

	if err := s.seedEventYears(ctx); err != nil {
		return err
	}

	organizers, err := s.seedOrganizers(ctx)
	if err != nil {
		return err
	}

	participations, err := s.seedParticipations(ctx, organizers, teams)
	if err != nil {
		return err
	}

	// Report what is in the database, not what this run happened to touch.
	_, organizerTotal, err := s.repos.Organizers.List(ctx, listAll(), repository.OrganizerListFilters{})
	if err != nil {
		return fmt.Errorf("failed to count organizers: %w", err)
	}

	_, participationTotal, err := s.repos.Participations.List(
		ctx, listAll(), repository.ParticipationListFilters{},
	)
	if err != nil {
		return fmt.Errorf("failed to count participations: %w", err)
	}

	slog.Info("Seeding finished successfully",
		"teams", len(teams),
		"organizers", organizerTotal,
		"created_this_run", participations,
		"participations", participationTotal,
		"active_year", DefaultActiveYear,
	)

	return nil
}

func (s *Seeder) seedTeams(ctx context.Context) ([]domain.Team, error) {
	teams := make([]domain.Team, 0, len(teamNames))

	for _, name := range teamNames {
		wantShifts := teamNeedsShifts(name)

		existing, err := s.repos.Teams.GetByName(ctx, name)
		if err == nil {
			// Keep the flag in sync so databases seeded before it existed get it.
			if existing.HasShifts != wantShifts {
				existing.HasShifts = wantShifts

				if updateErr := s.repos.Teams.Update(ctx, existing); updateErr != nil {
					return nil, fmt.Errorf("failed to update team %s: %w", name, updateErr)
				}
			}

			teams = append(teams, *existing)

			continue
		}

		if !gorm.IsNotFound(err) {
			return nil, fmt.Errorf("failed to look up team %s: %w", name, err)
		}

		team := &domain.Team{Name: name, HasShifts: wantShifts}
		if err := s.repos.Teams.Create(ctx, team); err != nil {
			return nil, fmt.Errorf("failed to create team %s: %w", name, err)
		}

		slog.Debug("Seeded team", "name", name, "id", team.ID, "has_shifts", team.HasShifts)
		teams = append(teams, *team)
	}

	return teams, nil
}

// teamNeedsShifts marks the seeded teams that require shift planning.
func teamNeedsShifts(name string) bool {
	return name == "Infodesk"
}

func (s *Seeder) seedEventYears(ctx context.Context) error {
	for _, year := range previousYears {
		if err := s.ensureEventYear(ctx, year); err != nil {
			return err
		}
	}

	if err := s.ensureEventYear(ctx, DefaultActiveYear); err != nil {
		return err
	}

	// ActivateYear deactivates every other year, which is what makes the active
	// year unique.
	if err := s.repos.EventYears.ActivateYear(ctx, DefaultActiveYear); err != nil {
		return fmt.Errorf("failed to activate event year %s: %w", DefaultActiveYear, err)
	}

	return nil
}

func (s *Seeder) ensureEventYear(ctx context.Context, year string) error {
	_, err := s.repos.EventYears.GetByYear(ctx, year)
	if err == nil {
		return nil
	}

	if !gorm.IsNotFound(err) {
		return fmt.Errorf("failed to look up event year %s: %w", year, err)
	}

	if err := s.repos.EventYears.Create(ctx, &domain.EventYear{Year: year}); err != nil {
		return fmt.Errorf("failed to create event year %s: %w", year, err)
	}

	slog.Debug("Seeded event year", "year", year)

	return nil
}

// seedOrganizers writes the two Dex accounts first, then the random ones. The
// Dex accounts are refreshed on every run so their sub and data stay correct.
func (s *Seeder) seedOrganizers(ctx context.Context) ([]domain.Organizer, error) {
	dexUsers := DexUsers()
	organizers := make([]domain.Organizer, 0, len(dexUsers)+s.organizerCount)

	for _, def := range dexUsers {
		organizer, err := s.upsertOrganizer(ctx, def.ToOrganizer())
		if err != nil {
			return nil, err
		}

		organizers = append(organizers, *organizer)
	}

	// organizerCount is a target size, not a growth factor. Without this, a run
	// with a different --rng-seed would generate different emails and pile up
	// another full set of organizers on top of the existing ones.
	missing, err := s.missingOrganizerCount(ctx, len(dexUsers))
	if err != nil {
		return nil, err
	}

	for range missing {
		organizer, err := s.upsertOrganizer(ctx, s.buildRandomOrganizer())
		if err != nil {
			return nil, err
		}

		organizers = append(organizers, *organizer)
	}

	return organizers, nil
}

// missingOrganizerCount reports how many random organizers still have to be
// generated to reach the target size.
func (s *Seeder) missingOrganizerCount(ctx context.Context, dexCount int) (int, error) {
	_, total, err := s.repos.Organizers.List(ctx, listAll(), repository.OrganizerListFilters{})
	if err != nil {
		return 0, fmt.Errorf("failed to count organizers: %w", err)
	}

	random := int(total) - dexCount
	if random < 0 {
		random = 0
	}

	return max(s.organizerCount-random, 0), nil
}

// upsertOrganizer updates the record when the email already exists and creates it
// otherwise. Keying on email is what makes repeated runs safe.
//
// A sub that a real login produced wins over the candidate, so the link between
// user and organizer is never broken. Records that still carry a NULL sub (e.g.
// from an older seed) are repaired with the candidate's sub instead.
func (s *Seeder) upsertOrganizer(ctx context.Context, candidate *domain.Organizer) (*domain.Organizer, error) {
	existing, err := s.repos.Organizers.GetByEmail(ctx, candidate.Email)
	if err != nil && !gorm.IsNotFound(err) {
		return nil, fmt.Errorf("failed to look up organizer %s: %w", candidate.Email, err)
	}

	if err == nil {
		candidate.ID = existing.ID
		// Preserve a sub that a real login produced, but repair records that
		// were seeded/created before a sub was available (e.g. NULL oidc_sub).
		if existing.OidcSub != nil {
			candidate.OidcSub = existing.OidcSub
		}

		// Slug is immutable: keep the stored value, or repair an empty one.
		if existing.Slug != "" {
			candidate.Slug = existing.Slug
		} else if candidate.Slug == "" {
			candidate.Slug = domain.Slugify(candidate.Handle)
		}

		candidate.Participations = nil

		if updateErr := s.repos.Organizers.Update(ctx, candidate); updateErr != nil {
			return nil, fmt.Errorf("failed to update organizer %s: %w", candidate.Email, updateErr)
		}

		return candidate, nil
	}

	if createErr := s.repos.Organizers.Create(ctx, candidate); createErr != nil {
		return nil, fmt.Errorf("failed to create organizer %s: %w", candidate.Email, createErr)
	}

	return candidate, nil
}

func (s *Seeder) seedParticipations(
	ctx context.Context,
	organizers []domain.Organizer,
	teams []domain.Team,
) (int, error) {
	created := 0

	for i := range organizers {
		organizer := &organizers[i]

		// Dex organizers keep the teams defined for them; everyone else gets a
		// random subset.
		preferred, isDex := dexTeamsFor(organizer.Email)
		teamsFor := pickTeams(teams, preferred, !isDex)

		participation, wasCreated, err := s.ensureParticipation(
			ctx, organizer, DefaultActiveYear, teamsFor,
		)
		if err != nil {
			return 0, err
		}

		if wasCreated {
			created++
		}

		// Give roughly 60% a previous year so the check-in has history to
		// prefill from.
		previous := previousYears[len(previousYears)-1]

		if !hasPreviousYear(organizer.Email) {
			continue
		}

		if _, _, err := s.ensureParticipation(
			ctx, organizer, previous, pickTeams(teams, nil, true),
		); err != nil {
			return 0, err
		}

		slog.Debug("Seeded participation", "organizer", organizer.Handle, "id", participation.ID)
	}

	return created, nil
}

// hasPreviousYear decides from a hash of the email rather than from the shared
// RNG stream.
//
// The stream position shifts between runs: on a second run the participations
// that already exist are skipped and therefore consume no randomness, so every
// later draw lands elsewhere. Deciding per organizer instead keeps the answer
// stable, which is what makes repeated runs add nothing.
func hasPreviousYear(email string) bool {
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(email))

	return float64(hash.Sum32()%1000)/1000 < previousYearRatio
}

// ensureParticipation creates a participation unless one exists for that
// organizer and year already.
func (s *Seeder) ensureParticipation(
	ctx context.Context,
	organizer *domain.Organizer,
	year string,
	teams []domain.Team,
) (*domain.Participation, bool, error) {
	existing, err := s.repos.Participations.GetByOrganizerAndYear(ctx, organizer.ID, year)
	if err == nil {
		return existing, false, nil
	}

	if !gorm.IsNotFound(err) {
		return nil, false, fmt.Errorf("failed to look up participation: %w", err)
	}

	participation := buildRandomParticipation(organizer, year, teams)
	if err := s.repos.Participations.Create(ctx, participation); err != nil {
		return nil, false, fmt.Errorf("failed to create participation: %w", err)
	}

	return participation, true, nil
}

func (s *Seeder) buildRandomOrganizer() *domain.Organizer {
	demozooID := uint(gofakeit.Number(1, 60000))

	organizer := &domain.Organizer{
		Handle:    s.uniqueHandle(),
		Realname:  gofakeit.Name(),
		Email:     gofakeit.Email(),
		Phone:     gofakeit.Phone(),
		Pronouns:  pick(pronounsOptions),
		Languages: pickLanguages(),
		DemozooID: &demozooID,
	}
	organizer.Slug = domain.Slugify(organizer.Handle)

	if gofakeit.Bool() {
		organizer.SlackHandle = strings.ToLower(gofakeit.Username())
	}

	if gofakeit.Bool() {
		organizer.AvatarURL = fmt.Sprintf(
			"https://api.dicebear.com/7.x/identicon/svg?seed=%s", organizer.Handle,
		)
	}

	return organizer
}

// uniqueHandle appends a counter so handles stay distinct across a seed run.
func (s *Seeder) uniqueHandle() string {
	s.handleSuffix++

	return fmt.Sprintf("%s%d", strings.ToLower(gofakeit.Username()), s.handleSuffix)
}

func buildRandomParticipation(
	organizer *domain.Organizer,
	year string,
	teams []domain.Team,
) *domain.Participation {
	return &domain.Participation{
		OrganizerID:           organizer.ID,
		EventYearID:           year,
		Status:                pickStatus(),
		TShirtSize:            pick(shirtSizes),
		AdditionalPaidShirt:   gofakeit.Bool(),
		PublicWebsiteConsent:  gofakeit.Bool(),
		ShiftPreferredTimes:   buildShiftTimes(),
		ShiftUnavailableTimes: buildShiftTimes(),
		Teams:                 teams,
	}
}

// buildShiftTimes returns distinct half-hour blocks inside the party window.
func buildShiftTimes() []int {
	total := partyDays * shiftBlocksPerDay
	count := gofakeit.Number(minShiftBlocks, maxShiftBlocks)

	blocks := make([]int, 0, count)

	for range count {
		blocks = append(blocks, gofakeit.Number(0, total-1))
	}

	slices.Sort(blocks)

	return slices.Compact(blocks)
}

func pickStatus() domain.ParticipationStatus {
	roll := gofakeit.Number(0, 99)

	switch {
	case roll < statusConfirmedPercent:
		return domain.ParticipationStatusConfirmed
	case roll < statusConfirmedPercent+statusPendingPercent:
		return domain.ParticipationStatusPending
	default:
		return domain.ParticipationStatusDeclined
	}
}

func pickLanguages() []string {
	languages := []string{"de"}

	if gofakeit.Bool() {
		languages = append(languages, "en")
	}

	if gofakeit.Bool() {
		if extra := pick(languagePool); !slices.Contains(languages, extra) {
			languages = append(languages, extra)
		}
	}

	return languages
}

// pickTeams returns the named teams when names are given, otherwise a random
// subset. Random is false for the Dex organizers, whose teams are fixed.
func pickTeams(teams []domain.Team, names []string, random bool) []domain.Team {
	if len(names) > 0 {
		selected := make([]domain.Team, 0, len(names))

		for _, name := range names {
			idx := slices.IndexFunc(teams, func(t domain.Team) bool { return t.Name == name })
			if idx >= 0 {
				selected = append(selected, teams[idx])
			}
		}

		return selected
	}

	if !random || len(teams) == 0 {
		return nil
	}

	count := gofakeit.Number(1, maxTeamsPerParticipation)

	// A set, not slices.CompactFunc: that only drops *adjacent* duplicates, so a
	// pick order of [A, B, A] would insert the join row for A twice.
	seen := make(map[uuid.UUID]bool, count)
	selected := make([]domain.Team, 0, count)

	for range count {
		team := teams[gofakeit.Number(0, len(teams)-1)]
		if seen[team.ID] {
			continue
		}

		seen[team.ID] = true

		selected = append(selected, team)
	}

	return selected
}

// dexTeamsFor returns the fixed team names of a Dex organizer, if the email
// belongs to one.
func dexTeamsFor(email string) ([]string, bool) {
	for _, def := range DexUsers() {
		if def.Email == email {
			return def.Teams, true
		}
	}

	return nil, false
}

func pick[T any](options []T) T {
	return options[gofakeit.Number(0, len(options)-1)]
}

// listAll disables sorting surprises by using the default column and no limit.
func listAll() repository.ListParams {
	return repository.ListParams{Offset: 0, Limit: 0, Sort: "id", Order: "ASC"}
}
