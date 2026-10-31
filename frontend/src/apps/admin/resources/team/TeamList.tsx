import { List, DataTable, TextInput } from "react-admin";

const teamFilters = [<TextInput key="q" source="q" label="Search" alwaysOn />];

const TeamList = () => (
  <List filters={teamFilters}>
    <DataTable rowClick="edit" bulkActionButtons={false}>
      <DataTable.Col source="name" />
    </DataTable>
  </List>
);

export default TeamList;
