package mailer

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/potibm/crewapparat/templates"
)

const inviteLinkSubject = "You are invited"

type InviteLinkData struct {
	EventName string
	Link      string
	ExpiresAt time.Time
}

func (mailer *Mailer) SendInviteLink(to string, data InviteLinkData) error {
	tpl, err := template.ParseFS(
		templates.MailTemplateFiles,
		"mail/invite_link.txt",
		"mail/_footer.txt",
	)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	var body bytes.Buffer
	if err := tpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	return mailer.SendMail(to, inviteLinkSubject, body.String())
}

const inviteSubmittedSubject = "We received your submission"

type InviteSubmittedData struct {
	EventName string
}

func (mailer *Mailer) SendInviteSubmitted(to string, data InviteSubmittedData) error {
	tpl, err := template.ParseFS(
		templates.MailTemplateFiles,
		"mail/invite_submitted.txt",
		"mail/_footer.txt",
	)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	var body bytes.Buffer
	if err := tpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	return mailer.SendMail(to, inviteSubmittedSubject, body.String())
}
