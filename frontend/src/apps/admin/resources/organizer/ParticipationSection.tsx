import { useEffect, useRef, useState } from "react";
import { useDataProvider, useGetList, useNotify } from "react-admin";
import {
  Autocomplete,
  Box,
  Button,
  Card,
  CardContent,
  CardHeader,
  Checkbox,
  FormControl,
  FormControlLabel,
  InputLabel,
  MenuItem,
  Select,
  Stack,
  TextField,
} from "@mui/material";
import SaveIcon from "@mui/icons-material/Save";
import { useAppConfig } from "@core/config/useConfig";

interface ParticipationSectionProps {
  organizerId?: string;
  activeYear?: string;
}

export const ParticipationSection = ({
  organizerId,
  activeYear,
}: ParticipationSectionProps) => {
  const dataProvider = useDataProvider();
  const notify = useNotify();
  const config = useAppConfig();

  const { data: participations, refetch: refetchParticipations } = useGetList(
    "participations",
    {
      pagination: { page: 1, perPage: 100 },
      filter: {
        organizer_id: organizerId,
        event_year_id: activeYear,
      },
    },
    { enabled: !!organizerId && !!activeYear },
  );

  const participation = participations?.[0];

  const { data: teams } = useGetList("teams", {
    pagination: { page: 1, perPage: 100 },
  });

  const teamChoices = (teams ?? []).map((t: { id: string; name: string }) => ({
    id: t.id,
    name: t.name,
  }));

  const [partStatus, setPartStatus] = useState("pending");
  const [partTShirt, setPartTShirt] = useState("");
  const [partAdditionalPaid, setPartAdditionalPaid] = useState(false);
  const [partConsent, setPartConsent] = useState(false);
  const [partTeams, setPartTeams] = useState<string[]>([]);
  const [savingPart, setSavingPart] = useState(false);

  const tshirtInputRef = useRef<HTMLInputElement>(null);

  const tshirtSizeChoices = Array.from(
    new Set([
      ...(config.tshirt_size_options ?? []),
      ...(partTShirt ? [partTShirt] : []),
    ]),
  ).map((value) => ({ id: value, name: value }));

  useEffect(() => {
    if (!participation) return;
    setPartStatus(participation.status ?? "pending");
    setPartTShirt(participation.t_shirt_size ?? "");
    setPartAdditionalPaid(!!participation.additional_paid_shirt);
    setPartConsent(!!participation.public_website_consent);
    setPartTeams((participation.teams ?? []).map((t: { id: string }) => t.id));
  }, [participation]);

  const handleSaveParticipation = async () => {
    if (!organizerId || !activeYear) return;
    setSavingPart(true);
    try {
      const payload = {
        organizer_id: organizerId,
        event_year_id: activeYear,
        status: partStatus,
        t_shirt_size: tshirtInputRef.current?.value?.trim() || partTShirt,
        additional_paid_shirt: partAdditionalPaid,
        public_website_consent: partConsent,
        teams: teamChoices.filter((t) => partTeams.includes(t.id)),
      };
      if (participation?.id) {
        await dataProvider.update("participations", {
          id: participation.id,
          data: payload,
          previousData: participation,
        });
      } else {
        await dataProvider.create("participations", { data: payload });
      }
      await refetchParticipations();
      notify("Participation saved", { type: "success" });
    } catch {
      notify("Failed to save participation", { type: "error" });
    } finally {
      setSavingPart(false);
    }
  };

  if (!activeYear) return null;

  return (
    <Card sx={{ width: "100%", mt: 3, mb: 2 }}>
      <CardHeader title={`Participation (${activeYear})`} />
      <CardContent>
        <Stack spacing={2}>
          <FormControl fullWidth size="small">
            <InputLabel>Status</InputLabel>
            <Select
              label="Status"
              value={partStatus}
              onChange={(e) => setPartStatus(e.target.value)}
            >
              <MenuItem value="pending">Pending</MenuItem>
              <MenuItem value="confirmed">Confirmed</MenuItem>
              <MenuItem value="declined">Declined</MenuItem>
            </Select>
          </FormControl>
          <Autocomplete
            size="small"
            fullWidth
            options={tshirtSizeChoices.map((c) => c.id)}
            value={partTShirt || null}
            onChange={(_event, newValue) => setPartTShirt(newValue ?? "")}
            freeSolo
            renderInput={(params) => (
              <TextField
                {...params}
                label="T-Shirt size"
                inputRef={tshirtInputRef}
              />
            )}
          />
          <FormControlLabel
            control={
              <Checkbox
                checked={partAdditionalPaid}
                onChange={(e) => setPartAdditionalPaid(e.target.checked)}
              />
            }
            label="Additional paid shirt"
          />
          <FormControlLabel
            control={
              <Checkbox
                checked={partConsent}
                onChange={(e) => setPartConsent(e.target.checked)}
              />
            }
            label="Public website consent"
          />
          <FormControl fullWidth size="small">
            <InputLabel>Teams</InputLabel>
            <Select
              multiple
              label="Teams"
              value={partTeams}
              onChange={(e) => setPartTeams(e.target.value as string[])}
            >
              {teamChoices.map((t) => (
                <MenuItem key={t.id} value={t.id}>
                  {t.name}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        </Stack>
      </CardContent>
      <Box
        sx={{
          display: "flex",
          alignItems: "center",
          px: 2,
          py: 1.5,
          bgcolor: (theme) =>
            theme.palette.mode === "dark"
              ? theme.palette.grey[800]
              : theme.palette.grey[300],
        }}
      >
        <Button
          variant="contained"
          startIcon={<SaveIcon />}
          onClick={handleSaveParticipation}
          disabled={savingPart}
        >
          {savingPart ? "Saving…" : "Save participation"}
        </Button>
      </Box>
    </Card>
  );
};
