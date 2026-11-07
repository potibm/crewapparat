import {
  Edit,
  SimpleForm,
  TextInput,
  BooleanInput,
  usePermissions,
  FormDataConsumer,
  useUpdate,
  useNotify,
  useRedirect,
  useGetRecordId,
} from "react-admin";
import { Alert, Button, Stack, Typography } from "@mui/material";

const statusMessages: Record<
  string,
  { severity: "info" | "success" | "warning"; text: string }
> = {
  sent: {
    severity: "success",
    text: "This invite has been approved and was sent to the invitee.",
  },
  pending_approval: {
    severity: "info",
    text: "This invite is pending approval by an admin.",
  },
  pending_review: {
    severity: "info",
    text: "The invitee has submitted their details; an admin is reviewing them.",
  },
  approved: { severity: "success", text: "This invite has been approved." },
  rejected: { severity: "warning", text: "This invite has been rejected." },
};

const InviteEdit = () => {
  const { permissions } = usePermissions();
  const isAdmin = Array.isArray(permissions) && permissions.includes("admin");
  const recordId = useGetRecordId();
  const [update] = useUpdate();
  const notify = useNotify();
  const redirect = useRedirect();

  const changeStatus = (status: string, formData: Record<string, unknown>) => {
    if (!recordId) return;
    update(
      "invites",
      {
        id: recordId,
        data: {
          email: formData.email,
          handle: formData.handle,
          demozoo_id: formData.demozoo_id,
          reason: formData.reason,
          status,
          review_note: formData.review_note,
          openid_account_created: formData.openid_account_created,
          slack_invited: formData.slack_invited,
        },
      },
      {
        onSuccess: () => {
          notify("Invite updated", { type: "success" });
          redirect("list", "invites");
        },
        onError: (error: unknown) => {
          notify(error instanceof Error ? error.message : "Update failed", {
            type: "error",
          });
        },
      },
    );
  };

  return (
    <Edit>
      <SimpleForm toolbar={false}>
        <FormDataConsumer>
          {({ formData }) => {
            const status = formData?.status ?? "";
            const showSubmission =
              status === "pending_review" ||
              status === "approved" ||
              status === "rejected";
            const info = statusMessages[status] ?? {
              severity: "info",
              text: "Status unknown.",
            };
            return (
              <>
                <Alert
                  severity={info.severity}
                  sx={{
                    width: "100%",
                    border: "1px solid",
                    borderColor: "info.light",
                    borderRadius: 1,
                  }}
                >
                  {info.text}
                </Alert>

                <Typography variant="subtitle2" sx={{ mt: 2 }}>
                  Invitation data (from the inviter)
                </Typography>
                <TextInput source="email" disabled />
                <TextInput source="handle" disabled />
                <TextInput source="demozoo_id" disabled />
                <TextInput
                  source="reason"
                  disabled
                  multiline
                  label="Reason (from the inviter)"
                />

                {showSubmission && (
                  <>
                    <Typography variant="subtitle2" sx={{ mt: 2 }}>
                      Submission (from the invitee)
                    </Typography>
                    <TextInput
                      source="notes"
                      disabled
                      multiline
                      label="Notes (from the invitee)"
                    />
                    <Typography variant="body2">
                      Code of Conduct accepted:{" "}
                      {formData?.coc_accepted ? "Yes" : "No"}
                    </Typography>
                    {formData?.coc_accepted_version && (
                      <Typography variant="body2">
                        Code of Conduct version: {formData.coc_accepted_version}
                      </Typography>
                    )}
                  </>
                )}

                {(status === "pending_approval" ||
                  status === "sent" ||
                  showSubmission) && (
                  <>
                    <Typography variant="subtitle2" sx={{ mt: 2 }}>
                      Review (admin)
                    </Typography>
                    <TextInput
                      source="review_note"
                      multiline
                      label="Review note (from the admin)"
                      disabled={!isAdmin || status === "sent"}
                    />
                  </>
                )}

                {isAdmin && (
                  <FormDataConsumer>
                    {({ formData: inner }) => (
                      <Stack direction="row" spacing={1} sx={{ mt: 2, mb: 2 }}>
                        {inner?.status === "pending_approval" && (
                          <Button
                            variant="contained"
                            color="success"
                            onClick={() => changeStatus("sent", inner ?? {})}
                          >
                            Approve &amp; Send
                          </Button>
                        )}
                        {inner?.status === "pending_review" && (
                          <Button
                            variant="contained"
                            color="success"
                            onClick={() =>
                              changeStatus("approved", inner ?? {})
                            }
                          >
                            Approve &amp; create organizer
                          </Button>
                        )}
                        {(inner?.status === "pending_approval" ||
                          inner?.status === "pending_review") && (
                          <Button
                            variant="outlined"
                            color="error"
                            onClick={() =>
                              changeStatus("rejected", inner ?? {})
                            }
                          >
                            Reject
                          </Button>
                        )}
                      </Stack>
                    )}
                  </FormDataConsumer>
                )}

                {isAdmin && status === "approved" && (
                  <>
                    <Typography variant="subtitle2" sx={{ mt: 2 }}>
                      Onboarding checklist (manual)
                    </Typography>
                    <Typography
                      variant="caption"
                      color="text.secondary"
                      sx={{ mb: 1, display: "block" }}
                    >
                      These steps do not change the invite status.
                    </Typography>
                    <BooleanInput
                      source="openid_account_created"
                      label="OpenID account created (id.evoke.eu)"
                    />
                    <BooleanInput
                      source="slack_invited"
                      label="Invited to Slack (evoke-eu.slack.com/admin)"
                    />
                    <Button
                      variant="contained"
                      sx={{ mt: 1 }}
                      onClick={() => changeStatus(status, formData ?? {})}
                    >
                      Save changes
                    </Button>
                  </>
                )}
              </>
            );
          }}
        </FormDataConsumer>
      </SimpleForm>
    </Edit>
  );
};

export default InviteEdit;
