package mailer

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/potibm/crewapparat/templates"
)

const invitePendingApprovalSubject = "New invite requires approval"

type InvitePendingApprovalData struct {
	Email         string
	InvitedByName string
	Link          string
}

func (mailer *Mailer) SendInvitePendingApproval(to string, data InvitePendingApprovalData) error {
	tpl, err := template.ParseFS(
		templates.MailTemplateFiles,
		"mail/invite_pending_approval.txt",
		"mail/_footer.txt",
	)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	var body bytes.Buffer
	if err := tpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	return mailer.SendMail(to, invitePendingApprovalSubject, body.String())
}
