import { List, DataTable, TextInput, SelectInput } from "react-admin";
import { Chip } from "@mui/material";
import type { ShiftPlanRecord } from "./shift-plan.types";

const filters = [
  <TextInput key="year" source="event_year_id" label="Year" />,
  <SelectInput
    key="status"
    source="status"
    choices={[
      { id: "pending", name: "Pending" },
      { id: "published", name: "Published" },
    ]}
  />,
];

const ShiftPlanList = () => (
  <List filters={filters} sort={{ field: "created_at", order: "DESC" }}>
    <DataTable bulkActionButtons={false} rowClick="show">
      <DataTable.Col source="version" label="Version" />
      <DataTable.Col source="event_year_id" label="Year" />
      <DataTable.Col
        source="status"
        render={(record: ShiftPlanRecord) => (
          <Chip
            label={record.status}
            size="small"
            color={record.status === "published" ? "success" : "default"}
          />
        )}
      />
      <DataTable.Col source="generated_at" label="Generated" />
    </DataTable>
  </List>
);

export default ShiftPlanList;