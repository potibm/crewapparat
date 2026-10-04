import { List, Datagrid, TextField, TextInput } from "react-admin";

const inviteFilters = [
  <TextInput key="q" source="q" label="Search" alwaysOn />,
];

const InviteList = () => (
  <List filters={inviteFilters}>
    <Datagrid rowClick="edit" bulkActionButtons={false}>
      <TextField source="email" />
      <TextField source="status" />
      <TextField source="expires_at" />
    </Datagrid>
  </List>
);

export default InviteList;
