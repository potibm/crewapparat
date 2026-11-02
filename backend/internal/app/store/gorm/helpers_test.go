package gorm

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// newTestStore opens a store on a fresh in-memory database. Every test gets
// its own so they cannot influence each other - a shared cache store would leak
// data between them.
func newTestStore(t *testing.T) *Store {
	t.Helper()

	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	dsn := fmt.Sprintf(
		"file:%s_%d?mode=memory&cache=shared&_busy_timeout=5000",
		name,
		time.Now().UnixNano(),
	)

	store, err := newStore(dsn)
	require.NoError(t, err, "store should open")

	t.Cleanup(func() {
		_ = store.Close()
	})

	require.NoError(t, store.PurgeAll(), "store should migrate cleanly")

	return store
}
