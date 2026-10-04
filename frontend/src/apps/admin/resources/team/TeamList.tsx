import { List, Datagrid, TextField, TextInput } from "react-admin";

const teamFilters = [<TextInput key="q" source="q" label="Search" alwaysOn />];

const TeamList = () => (
  <List filters={teamFilters}>
    <Datagrid rowClick="edit" bulkActionButtons={false}>
      <TextField source="name" />
    </Datagrid>
  </List>
);

export default TeamList;
