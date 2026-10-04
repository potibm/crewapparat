import { List, Datagrid, TextField, TextInput } from "react-admin";

const organizerFilters = [
  <TextInput key="q" source="q" label="Search" alwaysOn />,
];

const OrganizerList = () => (
  <List filters={organizerFilters}>
    <Datagrid rowClick="edit" bulkActionButtons={false}>
      <TextField source="handle" />
      <TextField source="realname" />
      <TextField source="email" />
      <TextField source="phone" />
    </Datagrid>
  </List>
);

export default OrganizerList;
