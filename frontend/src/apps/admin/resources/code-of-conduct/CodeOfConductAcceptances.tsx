import { useGetList, usePermissions, useRecordContext } from "react-admin";
import { Box, Card, CardContent, CardHeader, Typography } from "@mui/material";
import type {
  CodeOfConductAcceptanceRecord,
  CodeOfConductRecord,
} from "./code-of-conduct.types";

export const CodeOfConductAcceptances = () => {
  const record = useRecordContext<CodeOfConductRecord>();
  const { permissions } = usePermissions();
  const isAdmin = Array.isArray(permissions) && permissions.includes("admin");

  const { data, total, isPending } = useGetList<CodeOfConductAcceptanceRecord>(
    "code-of-conduct-acceptances",
    {
      filter: { code_of_conduct_id: record?.id },
      sort: { field: "accepted_at", order: "DESC" },
      pagination: { page: 1, perPage: 100 },
    },
    { enabled: isAdmin && Boolean(record?.id) },
  );

  if (!isAdmin) return null;

  return (
    <Box sx={{ p: 2 }}>
      <Card>
        <CardHeader title={`Accepted by ${total ?? 0} organizers`} />
        <CardContent>
          {isPending && <Typography>Loading…</Typography>}

          {!isPending && (!data || data.length === 0) && (
            <Typography color="text.secondary">
              Nobody accepted this version yet.
            </Typography>
          )}

          {!isPending && data && data.length > 0 && (
            <Box component="ul" sx={{ m: 0, pl: 2 }}>
              {data.map((acceptance) => (
                <li key={acceptance.id}>
                  <Typography variant="body2">
                    {acceptance.handle} ({acceptance.realname}) —{" "}
                    {new Date(acceptance.accepted_at).toLocaleString()}
                  </Typography>
                </li>
              ))}
            </Box>
          )}
        </CardContent>
      </Card>
    </Box>
  );
};
