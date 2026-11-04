import {
  Create,
  SimpleForm,
  TextInput,
  SelectInput,
  BooleanInput,
} from "react-admin";

const InviteCreate = () => (
  <Create>
    <SimpleForm>
      <TextInput source="email" />
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
      <TextInput source="provided_name" />
      <TextInput source="provided_phone" />
      <TextInput source="notes" />
      <BooleanInput source="coc_accepted" />
    </SimpleForm>
  </Create>
);

export default InviteCreate;
