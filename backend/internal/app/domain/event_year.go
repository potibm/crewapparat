package domain

// EventYear is one edition year, e.g. "2026". Exactly one year is active;
// uniqueness is enforced in the application (see
// EventYearRepository.ActivateYear) because SQLite has no partial unique
// indexes.
type EventYear struct {
	Year     string `json:"year"`
	IsActive bool   `json:"is_active"`
}
