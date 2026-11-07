# Core Workflows

## Invite workflow (new organizers)

1. **Creation.** An organizer (admin or non-admin) creates an invite in
   React Admin. The backend creates an `Invite`:
   - admin → status `sent`
   - non-admin → status `pending_approval` (admins are notified by email)
2. **Approval to send (non-admin only).** An admin reviews the pending
   invite and clicks "Approve & Send", which sets the status to `sent` and
   sends the magic link to the invitee.
3. **Data entry (public).** The invitee opens the magic link, reads the
   Code of Conduct, and submits notes + acceptance. Status becomes
   `pending_review`.
4. **Review.** An admin reviews the submitted data in React Admin.
5. **Approval.** The admin clicks "Approve & create organizer". The
   backend converts the invite into an `Organizer` record (plus a
   confirmed `Participation` for the active event year) and sets the
   status to `approved`.
6. **Provisioning (manual).** The admin creates the Slack and Rauthy
   accounts and ticks the onboarding checklist.
7. **Account linking.** On the invitee's first login, `LinkMe` matches
   the organizer by email and stores the OIDC `sub`.

Expired, unsubmitted invites can be purged with
`InviteWorkflowRepository.PurgeExpired`.

## Annual check-in (returning organizers)

Not yet implemented. See `CONCEPT.md` §4.2 for the intended flow.
```