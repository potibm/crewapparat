import { useState } from "react";
import {
  List,
  DataTable,
  TextInput,
  SelectInput,
  EditButton,
  DeleteButton,
  useGetList,
} from "react-admin";
import { AvatarLightbox } from "./AvatarLightbox";
import { OrganizerAvatar } from "./OrganizerAvatar";
import { SlackIcon } from "./OrganizerIcons";
import type { OrganizerRecord } from "./organizer.types";
import { Box, Chip, IconButton, Tooltip, Typography } from "@mui/material";
import EmailIcon from "@mui/icons-material/Email";
import PhoneIcon from "@mui/icons-material/Phone";
import { TeamChips } from "./TeamChips";

const statusChoices = [
  { id: "confirmed", name: "Active" },
  { id: "all", name: "All" },
  { id: "pending", name: "Pending" },
  { id: "declined", name: "Declined" },
];

const OrganizerListContent = () => {
  const [lightboxSrc, setLightboxSrc] = useState<string | null>(null);

  const { data: eventYears } = useGetList("event-years", {
    pagination: { page: 1, perPage: 100 },
  });
  const activeYear = eventYears?.find(
    (y: { is_active?: boolean; year: string }) => y.is_active,
  )?.year;

  const { data: teams } = useGetList("teams", {
    pagination: { page: 1, perPage: 100 },
    sort: { field: "name", order: "ASC" },
  });

  const teamChoices = (teams ?? [])
    .filter(
      (t: { id: string; name: string; active?: boolean }) => t.active !== false,
    )
    .map((t: { id: string; name: string }) => ({
      id: t.id,
      name: t.name,
    }));

  const organizerFilters = [
    <TextInput key="q" source="q" label="Search" alwaysOn />,
    <SelectInput
      key="status"
      source="status"
      label="Status"
      choices={statusChoices}
      alwaysOn
    />,
    <SelectInput
      key="team"
      source="team"
      label="Team"
      choices={teamChoices}
      alwaysOn
    />,
  ];

  return (
    <>
      <List
        filter={{ event_year_id: activeYear }}
        filterDefaultValues={{ status: "confirmed" }}
        filters={organizerFilters}
      >
        <DataTable bulkActionButtons={false} rowClick="show">
          <DataTable.Col
            label="Avatar"
            render={(record: OrganizerRecord) => {
              return (
                <OrganizerAvatar
                  src={record.avatar_url}
                  handle={record.handle}
                  realname={record.realname}
                  size={40}
                  onClick={(event) => {
                    event.stopPropagation();
                    if (record.avatar_url) setLightboxSrc(record.avatar_url);
                  }}
                />
              );
            }}
          />
          <DataTable.Col source="handle" />
          <DataTable.Col source="realname" />
          <DataTable.Col
            label="Contact"
            disableSort
            render={(record: OrganizerRecord) => (
              <Box sx={{ display: "flex", gap: 0.5 }}>
                <Tooltip title={record.email || "No email"} arrow>
                  <span>
                    <IconButton
                      size="small"
                      component="a"
                      href={record.email ? `mailto:${record.email}` : undefined}
                      disabled={!record.email}
                      onClick={(event) => event.stopPropagation()}
                      sx={{
                        color: record.email ? "action" : "action.disabled",
                      }}
                    >
                      <EmailIcon fontSize="small" />
                    </IconButton>
                  </span>
                </Tooltip>
                <Tooltip title={record.phone || "No phone"} arrow>
                  <span>
                    <IconButton
                      size="small"
                      component="a"
                      href={record.phone ? `tel:${record.phone}` : undefined}
                      disabled={!record.phone}
                      onClick={(event) => event.stopPropagation()}
                      sx={{
                        color: record.phone ? "action" : "action.disabled",
                      }}
                    >
                      <PhoneIcon fontSize="small" />
                    </IconButton>
                  </span>
                </Tooltip>
                <Tooltip title={record.slack_handle || "No Slack"} arrow>
                  <span>
                    <IconButton
                      size="small"
                      component="a"
                      href={record.slack_profile_url || undefined}
                      target="_blank"
                      rel="noopener noreferrer"
                      disabled={
                        !record.slack_handle && !record.slack_profile_url
                      }
                      onClick={(event) => event.stopPropagation()}
                      sx={{
                        color:
                          record.slack_handle || record.slack_profile_url
                            ? "action"
                            : "action.disabled",
                      }}
                    >
                      <SlackIcon fontSize="small" />
                    </IconButton>
                  </span>
                </Tooltip>
              </Box>
            )}
          />
          <DataTable.Col
            label="Status"
            disableSort
            render={(record: OrganizerRecord) => {
              const status = record.current_participation_status;
              const colorMap: Record<string, "success" | "default" | "error"> =
                {
                  confirmed: "success",
                  pending: "default",
                  declined: "error",
                };
              const labelMap: Record<string, string> = {
                confirmed: "Active",
                pending: "Pending",
                declined: "Declined",
              };
              return status ? (
                <Chip
                  label={labelMap[status] || status}
                  size="small"
                  color={colorMap[status] || "default"}
                  variant={status === "pending" ? "outlined" : "filled"}
                />
              ) : (
                <Typography variant="body2" color="text.secondary">
                  —
                </Typography>
              );
            }}
          />
          <DataTable.Col
            label="Teams"
            disableSort
            render={(record: OrganizerRecord) =>
              record.current_participation_teams &&
              record.current_participation_teams.length > 0 ? (
                <TeamChips
                  teams={record.current_participation_teams.filter(
                    (t) => t.active !== false,
                  )}
                />
              ) : (
                <Typography variant="body2" color="text.secondary">
                  —
                </Typography>
              )
            }
          />
          <DataTable.Col>
            <EditButton />
            <DeleteButton
              confirmTitle="Delete organizer"
              confirmContent="Are you sure you want to delete this organizer?"
            />
          </DataTable.Col>
        </DataTable>
      </List>
      <AvatarLightbox
        src={lightboxSrc}
        open={lightboxSrc !== null}
        onClose={() => setLightboxSrc(null)}
      />
    </>
  );
};

const OrganizerList = () => <OrganizerListContent />;
export default OrganizerList;
