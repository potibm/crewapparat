import { List, DataTable, TextInput } from "react-admin";

const inviteFilters = [
  <TextInput key="q" source="q" label="Search" alwaysOn />,
];

const InviteList = () => (
  <List filters={inviteFilters}>
    <DataTable rowClick="edit" bulkActionButtons={false}>
      <DataTable.Col source="email" />
      <DataTable.Col source="status" />
      <DataTable.Col source="expires_at" />
    </DataTable>
  </List>
);

export default InviteList;
