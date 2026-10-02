package domain

// EventYear ist ein Editionsjahr, z. B. "2026". Genau ein Jahr ist aktiv; die
// Eindeutigkeit wird in der Anwendung sichergestellt (siehe
// EventYearRepository.ActivateYear), weil SQLite keine partiellen
// Unique-Indizes unterstützt.
type EventYear struct {
	Year     string `json:"year"`
	IsActive bool   `json:"is_active"`
}
