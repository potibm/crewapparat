import { List, DataTable, TextInput } from "react-admin";

const participationFilters = [
  <TextInput key="q" source="q" label="Search" alwaysOn />,
];

const ParticipationList = () => (
  <List filters={participationFilters}>
    <DataTable rowClick="edit" bulkActionButtons={false}>
      <DataTable.Col source="organizer_id" />
      <DataTable.Col source="event_year_id" />
      <DataTable.Col source="status" />
      <DataTable.Col source="t_shirt_size" />
    </DataTable>
  </List>
);

export default ParticipationList;
