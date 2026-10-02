# Concept: crewapparat

## 1. Overview & Objectives

The `crewapparat` is the central service for managing the Evoke event organizers (crew). It solves the problem of scattered contact lists, stale data (e.g., outdated T-shirt sizes), and manual permission management.

The primary goals are a frictionless annual "check-in" process for returning crew members, a secure onboarding workflow for new organizers, and seamless integration into the existing Identity & Access Management (Rauthy OIDC). It also serves as a central "contact book" during the event.

## 2. Core Features

- **Master Data Management (CRUD):** Central directory of all organizers featuring demoscene handles, real names, contact info, and pronouns.
- **Event-Specific Tracking:** Strict separation between static master data and annual, event-specific data (e.g., T-shirt sizes, team assignments, shift preferences).
- **Onboarding Workflow (Invites):** A secure, magic-link-based public form for new organizers to enter their details and accept the Evoke Organizer Code of Conduct prior to account creation.
- **Annual Check-In:** An automated workflow asking existing organizers if they are joining the current year, pre-filling preferences based on their last participation.
- **Rauthy Integration (IAM):** Automatic assignment of checked-in organizers to the current year's Rauthy group (e.g., `organizer-2026`), governing access to all other tools (`kasseapparat`, `billedapparat`, etc.).

## 3. Data Model (GORM / SQLite)

The data model uses GORM's built-in JSON serializer to store arrays natively in SQLite, providing a clean JSON API for React Admin without manual parsing.

### 3.1 Master Data

```go
type Organizer struct {
    ID          uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
    OidcSub     *string   `gorm:"uniqueIndex"` // Tied to Rauthy upon first login
    Handle      string    // Demoscene Handle
    Realname    string    // Legal/Real name
    Email       string
    Phone       string
    SlackHandle string
    AvatarURL   string
    DemozooID   *uint
    Pronouns    string

    // Serialized as JSON array in SQLite -> API Output: ["de", "en"]
    Languages   []string  `gorm:"serializer:json"`

    Participations []Participation
}
```

### 3.2 Per-Event Data

```go
type EventYear struct {
    Year     string `gorm:"primaryKey"` // e.g., "2026"
    IsActive bool
}

type Team struct {
    ID   uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
    Name string    `gorm:"unique"` // e.g., "Infodesk", "Network"
}

type Participation struct {
    ID                   uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
    OrganizerID          uuid.UUID
    EventYearID          string
    Status               string    // "pending", "confirmed", "declined"

    // Merch & Public Info
    TShirtSize           string
    AdditionalPaidShirt  bool      `gorm:"default:false"`
    PublicWebsiteConsent bool      `gorm:"default:false"`

    // Shift Planning (IDs representing half-hour blocks)
    // API Output: [0, 1, 2, 32, 33]
    ShiftAvailableTimes   []int    `gorm:"serializer:json"`
    ShiftUnavailableTimes []int    `gorm:"serializer:json"`

    // UUIDs of preferred partners. API Output: ["uuid-1", "uuid-2"]
    ShiftPreferredPartner []string `gorm:"serializer:json"`

    // Many-to-Many: Which teams the organizer joined in THIS specific year
    Teams []Team `gorm:"many2many:participation_teams;"`
}
```

### 3.3 Invite & Onboarding Data

```go
type Invite struct {
    ID            uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
    Email         string    `gorm:"uniqueIndex"`
    Token         string    `gorm:"uniqueIndex"` // UUIDv4 Magic Link Token
    Status        string    // "sent", "pending_review", "approved", "rejected"

    // Form submission data
    ProvidedName  string
    ProvidedPhone string
    Notes         string
    CoCAccepted   bool
    CoCAcceptedAt *time.Time

    InvitedByID   uuid.UUID
    ReviewedByID  *uuid.UUID
}
```

## 4. Core Workflows

### 4.1 The Invite Workflow (New Organizers)

1. **Creation:** A Main Organizer enters an email address in React Admin. The backend creates an `Invite` (`status: sent`) and sends a magic link via AWS SES v2.
2. **Data Entry (Public Frontend):** The invitee clicks the link, lands on a public `CustomRoute` (no OIDC required), and fills in their basic details and the mandatory check of the _Organizer Code of Conduct_. Submit changes status to `pending_review`. Team assignment happens at approval, not here.
3. **Approval:** A Main Organizer reviews the submitted data in React Admin's "Pending Invites" view.
4. **Provisioning:**
   - Main Organizer manually creates the Slack and Rauthy accounts.
   - Main Organizer clicks "Approve" in React Admin. The backend converts the `Invite` record into a real `Organizer` record (matching the email).
5. **Account Linking:** Upon the user's first Rauthy login, the OIDC middleware detects an empty `oidc_sub` for that email in the database and automatically saves the ID, closing the loop securely.

### 4.2 The Annual Check-In (Returning Organizers)

1. **Trigger:** Admin starts the check-in for the upcoming event year (e.g., 2026).
2. **Notification:** Magic links or login prompts are sent to past organizers.
3. **Action:** The organizer logs in. The frontend fetches their `Participation` record from the previous year to pre-fill the form:
   - _“Are you joining in 2026?”_
   - _“Are your pronouns and languages still up to date?”_
   - _“Last year your shirt size was L. Still L? Do you want an extra paid shirt?”_
   - _“Are you in the Infodesk team again?”_
   - _“Select your shift availability and preferred partners.”_
4. **Sync:** Upon submission, the backend saves the new `Participation` record and calls the Rauthy API to add the user to the `organizer-2026` group, granting them access to current event tools.

## 5. Technical Stack

- **Backend:** Go, Gin, GORM (SQLite), AWS SES v2 (for emails).
- **API Design:** Standard REST JSON endpoints tailored to React Admin's `ra-data-simple-rest` data provider (handling sorting, pagination, and `X-Total-Count` headers).
- **Frontend:** React Admin (using Vite), Tailwind CSS / Flowbite (for the public onboarding route), Zod (schema validation for the public form).
- **IAM:** Rauthy (OIDC) acting as the single source of truth for authentication and group authorization.
