import { List, Datagrid, TextField, TextInput } from "react-admin";

const participationFilters = [
  <TextInput key="q" source="q" label="Search" alwaysOn />,
];

const ParticipationList = () => (
  <List filters={participationFilters}>
    <Datagrid rowClick="edit" bulkActionButtons={false}>
      <TextField source="organizer_id" />
      <TextField source="event_year_id" />
      <TextField source="status" />
      <TextField source="t_shirt_size" />
    </Datagrid>
  </List>
);

export default ParticipationList;
