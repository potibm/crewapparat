import { useWatch } from "react-hook-form";
import {
  Edit,
  SimpleForm,
  TextInput,
  SelectInput,
  TextField,
  usePermissions,
  required,
} from "react-admin";
import { Alert } from "@mui/material";
import { DemozooScenerInput } from "../organizer/DemozooScenerInput";

const statusMessages: Record<
  string,
  { severity: "info" | "success" | "warning" | "error"; text: string }
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
  const status = useWatch({ name: "status" }) as string | undefined;
  const info = statusMessages[status ?? ""] ?? {
    severity: "info",
    text: "Status unknown.",
  };

  return (
    <Edit>
      <SimpleForm>
        <TextInput source="email" validate={required()} />
        <TextInput source="handle" validate={required()} />
        <DemozooScenerInput />
        <TextInput source="reason" multiline />
        {isAdmin ? (
          <SelectInput
            source="status"
            choices={[
              { id: "sent", name: "Sent" },
              { id: "pending_approval", name: "Pending Approval" },
              { id: "pending_review", name: "Pending Review" },
              { id: "approved", name: "Approved" },
              { id: "rejected", name: "Rejected" },
            ]}
          />
        ) : (
          <TextField source="status" />
        )}
        <Alert severity={info.severity} sx={{ mt: 1 }}>
          {info.text}
        </Alert>
      </SimpleForm>
    </Edit>
  );
};

export default InviteEdit;
