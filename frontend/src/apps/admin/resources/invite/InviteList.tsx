import {
  List,
  DataTable,
  TextInput,
  SelectInput,
  FunctionField,
} from "react-admin";

const statusChoices = [
  { id: "pending_approval", name: "Pending approval" },
  { id: "pending_review", name: "Submitted, awaiting review" },
  { id: "sent", name: "Sent (magic link)" },
  { id: "approved", name: "Approved" },
  { id: "rejected", name: "Rejected" },
];

const statusLabel = (status: string) =>
  statusChoices.find((c) => c.id === status)?.name ?? status;

const inviteFilters = [
  <TextInput key="q" source="q" label="Search" alwaysOn />,
  <SelectInput key="status" source="status" choices={statusChoices} />,
];

const InviteList = () => (
  <List sort={{ field: "created_at", order: "DESC" }} filters={inviteFilters}>
    <DataTable rowClick="edit" bulkActionButtons={false}>
      <DataTable.Col source="email" />
      <DataTable.Col source="handle" />
      <DataTable.Col source="invited_by_name" label="Invited by" />
      <DataTable.Col source="status" label="Status">
        <FunctionField
          render={(record: { status?: string }) =>
            statusLabel(record?.status ?? "")
          }
        />
      </DataTable.Col>
      <DataTable.Col source="expires_at" label="Expires">
        <FunctionField
          render={(record: { status?: string; expires_at?: string }) => {
            if (
              record?.status !== "sent" &&
              record?.status !== "pending_review"
            ) {
              return "—";
            }
            return record?.expires_at
              ? new Date(record.expires_at).toLocaleString()
              : "—";
          }}
        />
      </DataTable.Col>
    </DataTable>
  </List>
);

export default InviteList;
