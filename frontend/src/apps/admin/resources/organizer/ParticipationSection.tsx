import { useEffect, useMemo, useRef, useState } from "react";
import { useDataProvider, useGetList, useNotify } from "react-admin";
import {
  Autocomplete,
  Box,
  Button,
  Card,
  CardContent,
  CardHeader,
  Checkbox,
  Divider,
  FormControl,
  FormControlLabel,
  InputLabel,
  MenuItem,
  Select,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import SaveIcon from "@mui/icons-material/Save";
import { useAppConfig } from "@core/config/useConfig";
import type { ParticipationRecord } from "../participation/participation.types";
import { ShiftSchedule } from "./ShiftSchedule";

interface ParticipationSectionProps {
  organizerId?: string;
  activeYear?: string;
}

interface TeamChoice {
  id: string;
  name: string;
  has_shifts?: boolean;
  active?: boolean;
}

interface OrganizerChoice {
  id: string;
  handle: string;
  realname?: string;
}

const PAGE_SIZE = 500;

export const ParticipationSection = ({
  organizerId,
  activeYear,
}: ParticipationSectionProps) => {
  const dataProvider = useDataProvider();
  const notify = useNotify();
  const config = useAppConfig();
  const tshirtDeadline = config.tshirt_deadline;
  const [tshirtPastDeadline] = useState(
    () =>
      !!tshirtDeadline &&
      !Number.isNaN(Date.parse(tshirtDeadline)) &&
      Date.now() > Date.parse(tshirtDeadline),
  );

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

  const participation = participations?.[0] as ParticipationRecord | undefined;

  const { data: teams } = useGetList("teams", {
    pagination: { page: 1, perPage: 100 },
  });

  const teamChoices = ((teams ?? []) as TeamChoice[])
    .filter((t) => t.active !== false)
    .map((t) => ({
      id: t.id,
      name: t.name,
    }));

  const [partStatus, setPartStatus] = useState("pending");
  const [partTShirt, setPartTShirt] = useState("");
  const [partAdditionalPaid, setPartAdditionalPaid] = useState(false);
  const [partConsent, setPartConsent] = useState(false);
  const [partTeams, setPartTeams] = useState<string[]>([]);
  const [preferredPartners, setPreferredPartners] = useState<string[]>([]);
  const [shiftPreferredTimes, setShiftPreferredTimes] = useState<number[]>([]);
  const [shiftUnavailableTimes, setShiftUnavailableTimes] = useState<number[]>(
    [],
  );
  const [savingPart, setSavingPart] = useState(false);

  const tshirtInputRef = useRef<HTMLInputElement>(null);

  const tshirtSizeChoices = Array.from(
    new Set([
      ...(config.tshirt_size_options ?? []),
      ...(partTShirt ? [partTShirt] : []),
    ]),
  ).map((value) => ({ id: value, name: value }));

  const shiftTeamIds = useMemo(
    () =>
      new Set(
        ((teams ?? []) as TeamChoice[])
          .filter((t) => t.has_shifts && partTeams.includes(t.id))
          .map((t) => t.id),
      ),
    [teams, partTeams],
  );

  const showShifts = shiftTeamIds.size > 0;

  // Organizers who share a shift team this year, used as partner choices.
  const { data: yearParticipations } = useGetList(
    "participations",
    {
      pagination: { page: 1, perPage: PAGE_SIZE },
      filter: { event_year_id: activeYear },
    },
    { enabled: showShifts && !!activeYear },
  );

  const { data: allOrganizers } = useGetList(
    "organizers",
    {
      pagination: { page: 1, perPage: PAGE_SIZE },
      sort: { field: "handle", order: "ASC" },
    },
    { enabled: showShifts },
  );

  const partnerChoices = useMemo(() => {
    const organizersById = new Map(
      ((allOrganizers ?? []) as OrganizerChoice[]).map((o) => [o.id, o]),
    );

    const partnerIds = new Set(
      ((yearParticipations ?? []) as ParticipationRecord[])
        .filter(
          (p) =>
            p.organizer_id !== organizerId &&
            (p.teams ?? []).some((t) => shiftTeamIds.has(t.id)),
        )
        .map((p) => p.organizer_id),
    );

    return [...partnerIds]
      .map((id) => organizersById.get(id))
      .filter((o): o is OrganizerChoice => o !== undefined)
      .map((o) => ({ id: o.id, name: o.handle }));
  }, [allOrganizers, yearParticipations, organizerId, shiftTeamIds]);

  // The form state mirrors the participation of the selected organizer/year.
  // Resetting it from an effect is intentional here: the object returned by
  // useGetList is the single source of truth and every switch between
  // organizers or years has to replace the local draft.
  /* eslint-disable @eslint-react/set-state-in-effect */
  useEffect(() => {
    if (!participation) {
      setPartStatus("pending");
      setPartTShirt("");
      setPartAdditionalPaid(false);
      setPartConsent(false);
      setPartTeams([]);
      setPreferredPartners([]);
      setShiftPreferredTimes([]);
      setShiftUnavailableTimes([]);

      return;
    }

    setPartStatus(participation.status ?? "pending");
    setPartTShirt(participation.t_shirt_size ?? "");
    setPartAdditionalPaid(!!participation.additional_paid_shirt);
    setPartConsent(!!participation.public_website_consent);
    setPartTeams((participation.teams ?? []).map((t) => t.id));
    setPreferredPartners(participation.shift_preferred_partner ?? []);
    setShiftPreferredTimes(participation.shift_preferred_times ?? []);
    setShiftUnavailableTimes(participation.shift_unavailable_times ?? []);
  }, [participation]);
  /* eslint-enable @eslint-react/set-state-in-effect */

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
        // Always send the shift data so unselecting the team only hides it
        // instead of deleting it.
        shift_preferred_times: shiftPreferredTimes,
        shift_unavailable_times: shiftUnavailableTimes,
        shift_preferred_partner: preferredPartners,
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
              {(participation?.status === "confirmed" ||
                participation?.status === "declined") === false && (
                <MenuItem value="pending">Pending</MenuItem>
              )}
              <MenuItem value="confirmed">Confirmed</MenuItem>
              <MenuItem value="declined">Declined</MenuItem>
            </Select>
          </FormControl>
          {partStatus !== "declined" && (
            <>
              <Autocomplete
                size="small"
                fullWidth
                disabled={tshirtPastDeadline}
                options={tshirtSizeChoices.map((c) => c.id)}
                value={partTShirt || null}
                onChange={(_event, newValue) => setPartTShirt(newValue ?? "")}
                freeSolo
                renderInput={(params) => (
                  <TextField
                    {...params}
                    label="T-Shirt size"
                    helperText={
                      tshirtDeadline
                        ? `Changes allowed until ${new Date(tshirtDeadline).toLocaleString()}`
                        : undefined
                    }
                    inputRef={tshirtInputRef}
                  />
                )}
              />
              <FormControlLabel
                control={
                  <Checkbox
                    checked={partAdditionalPaid}
                    disabled={tshirtPastDeadline}
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

              {showShifts && (
                <>
                  <Divider />
                  <Typography variant="h6">Shift planning</Typography>
                  {config.shift.start && config.shift.block_count > 0 ? (
                    <ShiftSchedule
                      start={config.shift.start}
                      blockCount={config.shift.block_count}
                      preferred={shiftPreferredTimes}
                      unavailable={shiftUnavailableTimes}
                      onChange={(preferred, unavailable) => {
                        setShiftPreferredTimes(preferred);
                        setShiftUnavailableTimes(unavailable);
                      }}
                    />
                  ) : (
                    <Typography variant="body2" color="text.secondary">
                      Shift planning is not configured.
                    </Typography>
                  )}
                  <Autocomplete
                    multiple
                    size="small"
                    fullWidth
                    options={partnerChoices.map((c) => c.id)}
                    getOptionLabel={(id) =>
                      partnerChoices.find((c) => c.id === id)?.name ?? id
                    }
                    value={preferredPartners}
                    onChange={(_event, value) => setPreferredPartners(value)}
                    renderInput={(params) => (
                      <TextField {...params} label="Preferred shift partners" />
                    )}
                  />
                </>
              )}
            </>
          )}
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
