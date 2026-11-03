package gorm

// Common WHERE clauses, shared by the repositories so the query strings stay
// consistent.
const (
	whereID    = "id = ?"
	whereYear  = "year = ?"
	whereEmail = "email = ?"
)
