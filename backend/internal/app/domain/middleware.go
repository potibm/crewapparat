package domain

type ContextKey string

const (
	UserIDKey ContextKey = "userID"

	UserEmailKey ContextKey = "userEmail"

	// UserEmailVerifiedKey holds a *bool: nil means the IdP did not send the
	// email_verified claim, a non-nil value carries the claim's boolean.
	UserEmailVerifiedKey ContextKey = "userEmailVerified"
)
