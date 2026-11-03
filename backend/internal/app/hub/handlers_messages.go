package hub

// Shared problem details used by the admin handlers. Keeping them in one place
// avoids drifting wording between the individual endpoints.
const (
	eventYearNotFound     = "Event year not found"
	inviteNotFound        = "Invite not found"
	organizerNotFound     = "Organizer not found"
	participationNotFound = "Participation not found"
	teamNotFound          = "Team not found"

	organizerGetErrorPrefix = "Failed to get organizer: "
)
