# ADR 001: Invite field visibility and editability by status

**Date:** 2026-11-07
**Status:** Accepted

## Context & Problem

The invite edit view grew organically and mixed fields from different
actors (inviter, invitee, admin) with the status-change buttons. It was
unclear which fields are editable when, and tracking checkboxes appeared
even before an invite was approved. We need a single, agreed matrix so
the UI stays consistent and reviewable.

## Decision

Field visibility/editability is driven by the invite status. Fields are
grouped by actor and rendered in this order:

1. Status banner (always, read-only)
2. Invitation data (from the inviter): `email`, `handle`, `demozoo_id`, `reason`
3. Submission (from the invitee, from `pending_review`): `notes`, `coc_accepted`
4. Review (admin, from `pending_review`): `review_note`
5. Status actions (admin, only `pending_approval`/`pending_review`)
6. Onboarding checklist (admin, only `approved`): `openid_account_created`, `slack_invited`

Matrix (s = shown, e = editable, – = hidden):

| Field | pending_approval | sent | pending_review | approved | rejected |
|---|---|---|---|---|---|
| email | s | s | s | s | s |
| handle | s | s | s | s | s |
| demozoo_id | s | s | s | s | s |
| reason (inviter) | s | s | s | s | s |
| notes (invitee) | – | – | s | s | s |
| coc_accepted | – | – | s | s | s |
| review_note (admin) | e | s | e | e | e |
| openid_account_created | – | – | – | e | – |
| slack_invited | – | – | – | e | – |

Rules:
- `email`, `handle`, `demozoo_id`, `reason` are never editable after creation.
- Tracking checkboxes only appear after approval.
- `Reject` is only offered while `pending_approval` or `pending_review`.
- `Approve & create organizer` also creates a confirmed participation for
  the active event year.

## Create view

Fields: `email` (required), `handle` (required), `demozoo_id` (optional),
`reason` (optional). Status is derived server-side from the caller's role
(admin → `sent`, non-admin → `pending_approval`).

## Consequences

- Positive: predictable, role- and status-aware UI; no invalid actions.
- Negative: more conditional rendering in `InviteEdit.tsx`.
```