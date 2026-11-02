import { List, DataTable, TextInput } from "react-admin";
import { Chip } from "@mui/material";

const eventYearFilters = [
  <TextInput key="q" source="q" label="Search" alwaysOn />,
];

const EventYearList = () => (
  <List filters={eventYearFilters}>
    <DataTable bulkActionButtons={false} rowClick={false}>
      <DataTable.Col source="year" />
      <DataTable.Col
        source="is_active"
        label="Active"
        render={(record: { is_active?: boolean }) =>
          record.is_active ? (
            <Chip label="Active" size="small" color="success" />
          ) : (
            <Chip label="Inactive" size="small" variant="outlined" />
          )
        }
      />
    </DataTable>
  </List>
);

export default EventYearList;
