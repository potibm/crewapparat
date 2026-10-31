import {
  List,
  DataTable,
  TextInput,
  EditButton,
  DeleteButton,
} from "react-admin";

const organizerFilters = [
  <TextInput key="q" source="q" label="Search" alwaysOn />,
];

const OrganizerList = () => (
  <List filters={organizerFilters}>
    <DataTable bulkActionButtons={false} rowClick="show">
       <DataTable.Col source="handle" />
       <DataTable.Col source="realname" />

       <DataTable.Col source="email" />
       <DataTable.Col source="phone" />
          <DataTable.Col>
                <EditButton />
      <DeleteButton
        confirmTitle="Delete organizer"
        confirmContent="Are you sure you want to delete this organizer?"
      />
            </DataTable.Col>
    </DataTable>
  </List>
);

export default OrganizerList;
