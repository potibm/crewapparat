import { useState } from "react";
import {
  Show,
  useRecordContext,
  useNotify,
  usePermissions,
  DataTable,
  useGetList,
} from "react-admin";
import {
  Box,
  Card,
  CardContent,
  CardHeader,
  Chip,
  IconButton,
  Link,
  Stack,
  Tooltip,
  Typography,
} from "@mui/material";
import ContentCopyIcon from "@mui/icons-material/ContentCopy";
import EmailIcon from "@mui/icons-material/Email";
import OpenInNewIcon from "@mui/icons-material/OpenInNew";
import PhoneIcon from "@mui/icons-material/Phone";
import { DemozooIcon, SlackIcon } from "./OrganizerIcons";
import { OrganizerRecord } from "./organizer.types";
import { AvatarLightbox } from "./AvatarLightbox";
import { OrganizerAvatar } from "./OrganizerAvatar";

const OrganizerShowContent = () => {
  const record = useRecordContext<OrganizerRecord>();
  const notify = useNotify();
  const { permissions } = usePermissions();
  const [lightboxSrc, setLightboxSrc] = useState<string | null>(null);

  const { data: eventYears } = useGetList("event-years", {
    pagination: { page: 1, perPage: 100 },
  });
  const activeYear = eventYears?.find(
    (y: { is_active?: boolean; year: string }) => y.is_active,
  )?.year;

  if (!record) return null;

  const isAdmin = Array.isArray(permissions) && permissions.includes("admin");

  const sortedParticipations = record.participations
    ? [...record.participations].sort((a, b) =>
        (b.event_year_id ?? "").localeCompare(a.event_year_id ?? ""),
      )
    : [];

  const publicUrl = `${window.location.origin}/u/${record.slug}`;
  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(publicUrl);
      notify("Copied to clipboard", { type: "success" });
    } catch {
      notify("Copy failed", { type: "error" });
    }
  };

  return (
    <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
      <Card>
        <CardContent
          sx={{
            display: "flex",
            alignItems: "center",
            gap: 3,
            flexWrap: "wrap",
          }}
        >
          <OrganizerAvatar
            src={record.avatar_url}
            handle={record.handle}
            realname={record.realname}
            size={96}
            onClick={() => {
              if (record.avatar_url) setLightboxSrc(record.avatar_url);
            }}
          />
          <Box sx={{ flex: 1, minWidth: 240 }}>
            <Typography variant="h4">{record.handle}</Typography>
            <Typography variant="subtitle1" color="text.secondary">
              {record.realname}
            </Typography>
            <Stack
              direction="row"
              spacing={1}
              sx={{ mt: 1, flexWrap: "wrap", gap: 1 }}
            >
              {record.pronouns && (
                <Chip label={record.pronouns} size="small" color="secondary" />
              )}
              {record.languages?.map((lang) => (
                <Chip key={lang} label={lang} size="small" variant="outlined" />
              ))}
            </Stack>
            <Box
              sx={{
                mt: 2,
                display: "flex",
                alignItems: "center",
                gap: 1,
                flexWrap: "wrap",
              }}
            >
              <Typography variant="body2" color="text.secondary">
                {publicUrl}
              </Typography>
              <Tooltip title="Copy public URL">
                <IconButton size="small" onClick={handleCopy}>
                  <ContentCopyIcon fontSize="small" />
                </IconButton>
              </Tooltip>
              <Tooltip title="Open public page">
                <IconButton
                  size="small"
                  component="a"
                  href={publicUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  <OpenInNewIcon fontSize="small" />
                </IconButton>
              </Tooltip>
            </Box>
          </Box>
        </CardContent>
      </Card>

      <Box
        sx={{
          display: "grid",
          gridTemplateColumns: { xs: "1fr", md: "1fr 1fr" },
          gap: 2,
        }}
      >
        <Card>
          <CardHeader title="Profile" />
          <CardContent>
            <Stack spacing={1.5}>
              <Box>
                <Typography variant="caption" color="text.secondary">
                  Real name
                </Typography>
                <Typography variant="body1">
                  {record.realname || "—"}
                </Typography>
              </Box>
              <Box>
                <Typography variant="caption" color="text.secondary">
                  Pronouns
                </Typography>
                <Typography variant="body1">
                  {record.pronouns || "—"}
                </Typography>
              </Box>
              <Box>
                <Typography variant="caption" color="text.secondary">
                  Languages
                </Typography>
                <Stack
                  direction="row"
                  spacing={1}
                  sx={{ mt: 0.5, flexWrap: "wrap" }}
                >
                  {record.languages && record.languages.length > 0 ? (
                    record.languages.map((lang) => (
                      <Chip
                        key={lang}
                        label={lang}
                        size="small"
                        variant="outlined"
                      />
                    ))
                  ) : (
                    <Typography variant="body1">—</Typography>
                  )}
                </Stack>
              </Box>
            </Stack>
          </CardContent>
        </Card>

        <Card>
          <CardHeader title="Contact" />
          <CardContent>
            <Stack spacing={1.5}>
              <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
                <EmailIcon fontSize="small" color="action" />
                {record.email ? (
                  <Link href={`mailto:${record.email}`}>{record.email}</Link>
                ) : (
                  <Typography variant="body1">—</Typography>
                )}
              </Box>
              <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
                <PhoneIcon fontSize="small" color="action" />
                {record.phone ? (
                  <Link href={`tel:${record.phone}`}>{record.phone}</Link>
                ) : (
                  <Typography variant="body1">—</Typography>
                )}
              </Box>
              <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
                <SlackIcon fontSize="small" color="action" />
                {record.slack_handle || record.slack_profile_url ? (
                  record.slack_profile_url ? (
                    <Link
                      href={record.slack_profile_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      sx={{
                        display: "inline-flex",
                        alignItems: "center",
                        gap: 0.5,
                      }}
                    >
                      {record.slack_handle || "Slack profile"}
                      <OpenInNewIcon fontSize="inherit" />
                    </Link>
                  ) : (
                    <Typography variant="body1">
                      {record.slack_handle}
                    </Typography>
                  )
                ) : (
                  <Typography variant="body1">—</Typography>
                )}
              </Box>
            </Stack>
          </CardContent>
        </Card>

        <Card>
          <CardHeader title="Links" />
          <CardContent>
            <Stack spacing={1.5}>
              <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
                <DemozooIcon fontSize="small" color="action" />
                {record.demozoo_id ? (
                  <Link
                    href={`https://demozoo.org/sceners/${record.demozoo_id}/`}
                    target="_blank"
                    rel="noopener noreferrer"
                    sx={{
                      display: "inline-flex",
                      alignItems: "center",
                      gap: 0.5,
                    }}
                  >
                    sceners/{record.demozoo_id}
                    <OpenInNewIcon fontSize="inherit" />
                  </Link>
                ) : (
                  <Typography variant="body1">—</Typography>
                )}
              </Box>
              {isAdmin && record.oidc_sub && (
                <Box>
                  <Typography variant="caption" color="text.secondary">
                    OIDC subject
                  </Typography>
                  <Typography variant="body2" sx={{ wordBreak: "break-all" }}>
                    {record.oidc_sub}
                  </Typography>
                </Box>
              )}
            </Stack>
          </CardContent>
        </Card>

        <Card sx={{ gridColumn: { xs: "auto", md: "1 / -1" } }}>
          <CardHeader title="Participations" />
          <CardContent>
            {sortedParticipations.length > 0 ? (
              <DataTable
                data={sortedParticipations}
                bulkActionButtons={false}
                rowClick={false}
                rowSx={(rec: Record<string, unknown>) =>
                  rec.event_year_id === activeYear
                    ? { bgcolor: "action.selected" }
                    : {}
                }
              >
                {" "}
                <DataTable.Col source="event_year_id" />
                <DataTable.Col source="status" />
                <DataTable.Col source="t_shirt_size" />
                <DataTable.Col
                  label="Teams"
                  render={(participation) =>
                    participation.teams && participation.teams.length > 0
                      ? participation.teams
                          .map(
                            (team: { id: string; name: string }) => team.name,
                          )
                          .join(", ")
                      : "—"
                  }
                />
              </DataTable>
            ) : (
              <Typography color="text.secondary">
                No participations yet.
              </Typography>
            )}
          </CardContent>
        </Card>
      </Box>
      <AvatarLightbox
        src={lightboxSrc}
        open={lightboxSrc !== null}
        onClose={() => setLightboxSrc(null)}
      />
    </Box>
  );
};

const OrganizerShow = () => (
  <Show>
    <OrganizerShowContent />
  </Show>
);

export default OrganizerShow;
