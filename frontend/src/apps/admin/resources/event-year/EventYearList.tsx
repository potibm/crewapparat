import { List, Datagrid, TextField, TextInput } from "react-admin";

const eventYearFilters = [
  <TextInput key="q" source="q" label="Search" alwaysOn />,
];

const EventYearList = () => (
  <List filters={eventYearFilters}>
    <Datagrid rowClick="edit" bulkActionButtons={false}>
      <TextField source="year" />
      <TextField source="is_active" />
    </Datagrid>
  </List>
);

export default EventYearList;
