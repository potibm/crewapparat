import {
  Edit,
  SimpleForm,
  TextInput,
  TextField,
  usePermissions,
  FormDataConsumer,
  useUpdate,
  useNotify,
  useRedirect,
  useGetRecordId,
  BooleanInput,
} from "react-admin";
import { Alert, Button, Stack } from "@mui/material";

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
      <SimpleForm>
        <FormDataConsumer>
          {({ formData }) => {
            const info = statusMessages[formData?.status ?? ""] ?? {
              severity: "info",
              text: "Status unknown.",
            };
            return <Alert severity={info.severity}>{info.text}</Alert>;
          }}
        </FormDataConsumer>
        <TextField source="email" />
        <TextField source="handle" />
        <TextField source="demozoo_id" />
        <TextField source="reason" />
        <TextField source="notes" />
        {isAdmin ? (
          <>
            <BooleanInput
              source="openid_account_created"
              label="OpenID account created (id.evoke.eu)"
            />
            <BooleanInput
              source="slack_invited"
              label="Invited to Slack (evoke-eu.slack.com/admin)"
            />
            <TextInput source="review_note" multiline />
            <FormDataConsumer>
              {({ formData }) => (
                <Stack direction="row" spacing={1} sx={{ mt: 1 }}>
                  <Button
                    variant="contained"
                    color="success"
                    onClick={() => changeStatus("sent", formData ?? {})}
                  >
                    Approve &amp; Send
                  </Button>
                  <Button
                    variant="outlined"
                    color="error"
                    onClick={() => changeStatus("rejected", formData ?? {})}
                  >
                    Reject
                  </Button>
                </Stack>
              )}
            </FormDataConsumer>
          </>
        ) : (
          <>
            <TextField source="openid_account_created" />
            <TextField source="slack_invited" />
            <TextField source="review_note" />
          </>
        )}
      </SimpleForm>
    </Edit>
  );
};

export default InviteEdit;
