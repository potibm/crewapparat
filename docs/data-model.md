# Data Model

Reference for the GORM/SQLite data model. The canonical definitions live in
`backend/internal/app/domain/`; the DB mappings in `backend/internal/app/store/gorm/`.

## Master data

### Organizer

```go
type Organizer struct {
    ID                   uuid.UUID
    OidcSub              *string   // set on first login against the IdP
    Handle               string
    Slug                 string    // immutable, slugified handle, unique
    Realname             string
    Email                string
    Phone                string
    SlackHandle          string
    SlackProfileURL      string
    MediaWikiProfileURL  string
    AvatarURL            string
    DemozooID            *uint
    FirstYearAsOrganizer *int
    Pronouns             string
    Languages            []string  // JSON array in SQLite
}
```

## Per-event data

### EventYear

```go
type EventYear struct {
    Year     string // primary key, e.g. "2026"
    IsActive bool
}
```

### Team

```go
type Team struct {
    ID        uuid.UUID
    Name      string // unique
    HasShifts bool
}
```

### Participation

```go
type Participation struct {
    ID                    uuid.UUID
    OrganizerID           uuid.UUID
    EventYearID           string
    Status                string // "pending", "confirmed", "declined"
    TShirtSize            string
    AdditionalPaidShirt   bool
    PublicWebsiteConsent  bool
    ShiftPreferredTimes   []int    // JSON array
    ShiftUnavailableTimes []int    // JSON array
    ShiftPreferredPartner []string // UUIDs, JSON array
    Teams                 []Team   // many-to-many
}
```

## Invite & onboarding data

### Invite

```go
type Invite struct {
    ID                   uuid.UUID
    Email                string // unique
    Status               string // "sent", "pending_approval", "pending_review", "approved", "rejected"
    Handle               string
    DemozooID            *uint
    Reason               string
    ReviewNote           string
    OpenIDAccountCreated bool
    SlackInvited         bool
    Token                string // magic-link credential (never serialized)
    ExpiresAt            time.Time
    Notes                string
    CoCAccepted          bool
    CoCAcceptedAt        *time.Time
    CoCAcceptedVersion   string
    InvitedByID          uuid.UUID
    ReviewedByID         *uuid.UUID
}
```

## Notes

- UUID primary keys are filled by the `BeforeCreate` hook in
  `store/gorm/uuid.go` (SQLite has no `uuid_generate_v4()`).
- Base models embed `TimestampModel` + `AuditModel`; audit columns are filled
  from the request context key `domain.UserIDKey`.
- Field parity between domain and DB structs is enforced by
  `store/gorm/mapper_parity_test.go`.
```