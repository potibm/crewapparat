import { List, DataTable, TextInput } from "react-admin";

const eventYearFilters = [
  <TextInput key="q" source="q" label="Search" alwaysOn />,
];

const EventYearList = () => (
  <List filters={eventYearFilters}>
    <DataTable rowClick="edit" bulkActionButtons={false}>
      <DataTable.Col source="year" />
      <DataTable.Col source="is_active" />
    </DataTable>
  </List>
);

export default EventYearList;
