import {
  List,
  DataTable,
  TextInput,
  EditButton,
  DeleteButton,
  usePermissions,
  TopToolbar,
  ExportButton,
} from "react-admin";
import { BooleanField } from "react-admin";
import { Box, Typography } from "@mui/material";
import type { TeamRecord } from "./team.types";

const teamFilters = [<TextInput key="q" source="q" label="Search" alwaysOn />];

const TeamListContent = () => {
  const { permissions } = usePermissions();
  const isAdmin = Array.isArray(permissions) && permissions.includes("admin");

  return (
    <List
      filters={teamFilters}
      actions={
        isAdmin ? undefined : (
          <TopToolbar>
            <ExportButton />
          </TopToolbar>
        )
      }
    >
      <DataTable rowClick="edit" bulkActionButtons={false}>
        <DataTable.Col source="name" />
        <DataTable.Col
          label="Color"
          disableSort
          render={(record: TeamRecord) => (
            <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
              <Box
                sx={{
                  width: 16,
                  height: 16,
                  borderRadius: "50%",
                  bgcolor: record.color || "#E5E7EB",
                  border: 1,
                  borderColor: "divider",
                }}
              />
              <Typography variant="body2">{record.color || "—"}</Typography>
            </Box>
          )}
        />
        <DataTable.Col source="has_shifts">
          <BooleanField source="has_shifts" />
        </DataTable.Col>
        <DataTable.Col>
          <EditButton />
          {isAdmin && (
            <DeleteButton
              confirmTitle="Delete team"
              confirmContent="Are you sure you want to delete this team?"
            />
          )}
        </DataTable.Col>
      </DataTable>
    </List>
  );
};

const TeamList = () => <TeamListContent />;
export default TeamList;
