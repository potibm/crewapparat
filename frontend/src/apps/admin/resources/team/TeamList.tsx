import { List, DataTable, TextInput, BooleanField } from "react-admin";

const teamFilters = [<TextInput key="q" source="q" label="Search" alwaysOn />];

const TeamList = () => (
  <List filters={teamFilters}>
    <DataTable rowClick="edit" bulkActionButtons={false}>
      <DataTable.Col source="name" />
      <DataTable.Col source="has_shifts">
        <BooleanField source="has_shifts" />
      </DataTable.Col>
    </DataTable>
  </List>
);

export default TeamList;
