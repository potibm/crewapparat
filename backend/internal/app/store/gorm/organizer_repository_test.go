package gorm

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/potibm/crewapparat/internal/app/domain"
	"github.com/potibm/crewapparat/internal/app/repository"
)

func ptrOf[T any](v T) *T { return &v }

func newOrganizerFixture(t *testing.T, mutate func(o *domain.Organizer)) *domain.Organizer {
	t.Helper()

	o := &domain.Organizer{
		Handle:          "handle",
		Realname:        "Real Name",
		Email:           fmtUnique("crew") + "@example.com",
		Phone:           "+49 123",
		SlackHandle:     "slack",
		SlackProfileURL: "https://slack.example/u/abc",
		AvatarURL:       "https://example.com/a.png",
		Pronouns:        "they/them",
		Languages:       []string{"de", "en"},
	}

	if mutate != nil {
		mutate(o)
	}

	return o
}

var uniqueCounter int

func fmtUnique(prefix string) string {
	uniqueCounter++

	return prefix + "-" + uuid.NewString()[:8]
}

func TestOrganizerRepository_CreateVergibtUUID(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewOrganizerRepository()
	ctx := context.Background()

	o := newOrganizerFixture(t, nil)
	o.ID = uuid.Nil

	require.NoError(t, repo.Create(ctx, o))
	require.NotEqual(t, uuid.Nil, o.ID, "Create muss eine UUID vergeben")
}

func TestOrganizerRepository_RoundtripErhaeltAlleFelder(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewOrganizerRepository()
	ctx := context.Background()

	sub := "rauthy-sub-abc"
	demozoo := uint(4711)
	original := newOrganizerFixture(t, func(o *domain.Organizer) {
		o.OidcSub = &sub
		o.DemozooID = &demozoo
	})

	require.NoError(t, repo.Create(ctx, original))

	loaded, err := repo.GetByID(ctx, original.ID)
	require.NoError(t, err)

	require.Equal(t, original.OidcSub, loaded.OidcSub, "OidcSub ging verloren")
	require.Equal(t, original.DemozooID, loaded.DemozooID)
	require.Equal(t, original.Handle, loaded.Handle)
	require.Equal(t, original.Realname, loaded.Realname)
	require.Equal(t, original.Email, loaded.Email)
	require.Equal(t, original.Phone, loaded.Phone)
	require.Equal(t, original.SlackHandle, loaded.SlackHandle)
	require.Equal(t, original.SlackProfileURL, loaded.SlackProfileURL)
	require.Equal(t, original.AvatarURL, loaded.AvatarURL)
	require.Equal(t, original.Pronouns, loaded.Pronouns)
	require.Equal(t, original.Languages, loaded.Languages)
}

func TestOrganizerRepository_GetByOidcSub(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewOrganizerRepository()
	ctx := context.Background()

	sub := "rauthy-sub-xyz"
	o := newOrganizerFixture(t, func(o *domain.Organizer) { o.OidcSub = &sub })
	require.NoError(t, repo.Create(ctx, o))

	found, err := repo.GetByOidcSub(ctx, sub)
	require.NoError(t, err, "Spaltenname muss oidc_sub sein, nicht oidcSub")
	require.Equal(t, o.ID, found.ID)
}

func TestOrganizerRepository_GetByOidcSubNichtGefunden(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewOrganizerRepository()

	_, err := repo.GetByOidcSub(context.Background(), "unbekannt")
	require.True(t, IsNotFound(err), "erwartet ErrRecordNotFound, bekam %v", err)
}

func TestOrganizerRepository_GetByEmail(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewOrganizerRepository()
	ctx := context.Background()

	o := newOrganizerFixture(t, nil)
	require.NoError(t, repo.Create(ctx, o))

	found, err := repo.GetByEmail(ctx, o.Email)
	require.NoError(t, err)
	require.Equal(t, o.ID, found.ID)
}

func TestOrganizerRepository_Update(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewOrganizerRepository()
	ctx := context.Background()

	o := newOrganizerFixture(t, nil)
	require.NoError(t, repo.Create(ctx, o))

	var createdAt string
	require.NoError(t, store.db.Raw(
		"SELECT created_at FROM organizers WHERE id = ?", o.ID,
	).Scan(&createdAt).Error)

	o.Realname = "Neuer Name"
	o.Languages = []string{"de", "fr"}
	require.NoError(t, repo.Update(ctx, o))

	loaded, err := repo.GetByID(ctx, o.ID)
	require.NoError(t, err)
	require.Equal(t, "Neuer Name", loaded.Realname)
	require.Equal(t, []string{"de", "fr"}, loaded.Languages)

	var createdAtAfter string
	require.NoError(t, store.db.Raw(
		"SELECT created_at FROM organizers WHERE id = ?", o.ID,
	).Scan(&createdAtAfter).Error)
	require.Equal(t, createdAt, createdAtAfter, "Update darf created_at nicht überschreiben")
}

func TestOrganizerRepository_UpdateUnbekannt(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewOrganizerRepository()

	o := newOrganizerFixture(t, nil)
	o.ID = uuid.New()

	err := repo.Update(context.Background(), o)
	require.True(t, IsNotFound(err), "Update darf keinen Datensatz erzeugen, bekam %v", err)
}

func TestOrganizerRepository_Delete(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewOrganizerRepository()
	ctx := context.Background()

	o := newOrganizerFixture(t, nil)
	require.NoError(t, repo.Create(ctx, o))

	require.NoError(t, repo.Delete(ctx, o.ID))

	_, err := repo.GetByID(ctx, o.ID)
	require.True(t, IsNotFound(err))
}

func TestOrganizerRepository_DeleteUnbekannt(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewOrganizerRepository()

	err := repo.Delete(context.Background(), uuid.New())
	require.True(t, IsNotFound(err), "erwartet ErrRecordNotFound, bekam %v", err)
}

func TestOrganizerRepository_ListMitFilterUndPaginierung(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewOrganizerRepository()
	ctx := context.Background()

	for _, handle := range []string{"alpha", "beta", "gamma"} {
		o := newOrganizerFixture(t, nil)
		o.Handle = handle
		o.Realname = "Real " + handle
		require.NoError(t, repo.Create(ctx, o))
	}

	list, total, err := repo.List(ctx,
		repository.OrganizerListParams{Offset: 0, Limit: 2, Sort: "handle", Order: "ASC"},
		repository.OrganizerListFilters{})
	require.NoError(t, err)
	require.Equal(t, int64(3), total, "total muss alle Treffer zählen, nicht die Seite")
	require.Len(t, list, 2)

	list, _, err = repo.List(ctx,
		repository.OrganizerListParams{Offset: 0, Limit: 10, Sort: "handle", Order: "ASC"},
		repository.OrganizerListFilters{Query: ptrOf("alpha")})
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "alpha", list[0].Handle)
}

// TestOrganizerRepository_FilterKlammernRegression fällt nur mit zwei
// Filtern gleichzeitig: ohne Klammern um das OR bindet das zweite AND stärker
// und der Filter liefert stillschweigend falsche Treffer.
func TestOrganizerRepository_FilterKlammernRegression(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewOrganizerRepository()
	ctx := context.Background()

	hit := newOrganizerFixture(t, func(o *domain.Organizer) {
		o.Handle = "needle"
		o.Realname = "Nadel"
		o.Email = "needle@example.com"
	})
	require.NoError(t, repo.Create(ctx, hit))

	other := newOrganizerFixture(t, func(o *domain.Organizer) {
		o.Handle = "other"
		o.Realname = "Other"
	})
	require.NoError(t, repo.Create(ctx, other))

	list, _, err := repo.List(ctx,
		repository.OrganizerListParams{Limit: 10, Sort: "handle", Order: "ASC"},
		repository.OrganizerListFilters{
			Query: ptrOf("needle"),
			ID:    &hit.ID,
		})
	require.NoError(t, err)
	require.Len(t, list, 1, "erwartet genau den Treffer mit Handle=needle")
	require.Equal(t, hit.ID, list[0].ID)
}

func TestOrganizerRepository_SortWhitelist(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewOrganizerRepository()
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, newOrganizerFixture(t, nil)))

	for _, allowed := range []string{"", "id", "handle", "realname", "email"} {
		_, _, err := repo.List(ctx,
			repository.OrganizerListParams{Limit: 10, Sort: allowed, Order: "ASC"},
			repository.OrganizerListFilters{})
		require.NoErrorf(t, err, "Spalte %q sollte erlaubt sein", allowed)
	}
}

func TestOrganizerRepository_SortWhitelistLehntInjektionAb(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewOrganizerRepository()
	ctx := context.Background()

	// Ungeprüft landet der Wert im ORDER BY, wo SQLite Unterabfragen erlaubt.
	for _, evil := range []string{
		"handle; DROP TABLE organizers--",
		"(SELECT password FROM organizers)",
		"token_hash",
	} {
		_, _, err := repo.List(ctx,
			repository.OrganizerListParams{Limit: 10, Sort: evil, Order: "ASC"},
			repository.OrganizerListFilters{})
		require.Errorf(t, err, "Spalte %q muss abgelehnt werden", evil)
	}

	require.True(t, store.db.Migrator().HasTable("organizers"), "Tabelle muss unversehrt sein")
}

func TestOrganizerRepository_UniqueEmail(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewOrganizerRepository()
	ctx := context.Background()

	first := newOrganizerFixture(t, nil)
	require.NoError(t, repo.Create(ctx, first))

	second := newOrganizerFixture(t, func(o *domain.Organizer) { o.Email = first.Email })
	require.Error(t, repo.Create(ctx, second), "doppelte E-Mail muss scheitern")
}

func TestOrganizerRepository_UniqueOidcSub(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewOrganizerRepository()
	ctx := context.Background()

	sub := "sub-doppelt"
	first := newOrganizerFixture(t, func(o *domain.Organizer) { o.OidcSub = &sub })
	require.NoError(t, repo.Create(ctx, first))

	second := newOrganizerFixture(t, func(o *domain.Organizer) { o.OidcSub = &sub })
	require.Error(t, repo.Create(ctx, second), "doppelter oidc_sub muss scheitern")
}
