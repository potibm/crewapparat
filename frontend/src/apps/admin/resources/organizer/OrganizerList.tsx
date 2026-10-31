import {
  List,
  Datagrid,
  TextField,
  TextInput,
  EditButton,
  DeleteButton,
} from "react-admin";

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
      <EditButton />
      <DeleteButton
        confirmTitle="Delete organizer"
        confirmContent="Are you sure you want to delete this organizer?"
      />
    </Datagrid>
  </List>
);

export default OrganizerList;
