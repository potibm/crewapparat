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
  required,
} from "react-admin";
import { Alert, Button, Stack } from "@mui/material";
import { DemozooScenerInput } from "../organizer/DemozooScenerInput";

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
        <TextInput source="email" validate={required()} />
        <TextInput source="handle" validate={required()} />
        <DemozooScenerInput />
        <TextInput source="reason" multiline />
        {isAdmin ? (
          <>
            <TextField source="status" />
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
            <TextField source="status" />
            <TextField source="review_note" />
          </>
        )}
        <FormDataConsumer>
          {({ formData }) => {
            const info = statusMessages[formData?.status ?? ""] ?? {
              severity: "info",
              text: "Status unknown.",
            };
            return <Alert severity={info.severity}>{info.text}</Alert>;
          }}
        </FormDataConsumer>
      </SimpleForm>
    </Edit>
  );
};

export default InviteEdit;
