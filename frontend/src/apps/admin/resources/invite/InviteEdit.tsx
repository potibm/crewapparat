import {
  Edit,
  SimpleForm,
  TextInput,
  SelectInput,
  TextField,
  usePermissions,
  required,
} from "react-admin";
import { DemozooScenerInput } from "../organizer/DemozooScenerInput";

const InviteEdit = () => {
  const { permissions } = usePermissions();
  const isAdmin = Array.isArray(permissions) && permissions.includes("admin");

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
      </SimpleForm>
    </Edit>
  );
};

export default InviteEdit;
