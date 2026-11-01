import { useEffect, useRef, useState } from "react";
import {
  Edit,
  SimpleForm,
  TextInput,
  AutocompleteArrayInput,
  AutocompleteInput,
  NumberInput,
  ImageInput,
  ImageField,
  useEditContext,
  useGetIdentity,
  usePermissions,
  useRedirect,
  useNotify,
  useGetList,
  useDataProvider,
} from "react-admin";
import {
  Card,
  CardContent,
  CardHeader,
  Stack,
  Button,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  TextField,
  Autocomplete,
  Checkbox,
  FormControlLabel,
} from "@mui/material";
import { useAppConfig } from "@core/config/useConfig";

const OrganizerEditForm = () => {
  const { record, isPending } = useEditContext();
  const { data: identity } = useGetIdentity();
  const { permissions } = usePermissions();
  const redirect = useRedirect();
  const notify = useNotify();
  const dataProvider = useDataProvider();
  const config = useAppConfig();

  const { data: eventYears } = useGetList("event-years", {
    pagination: { page: 1, perPage: 100 },
  });
  const activeYear = eventYears?.find(
    (y: { is_active?: boolean; year: string }) => y.is_active,
  )?.year;

  const { data: participations, refetch: refetchParticipations } = useGetList(
    "participations",
    {
      pagination: { page: 1, perPage: 100 },
      filter: {
        organizer_id: record?.id,
        event_year_id: activeYear,
      },
    },
    { enabled: !!record?.id && !!activeYear },
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

  useEffect(() => {
    if (!participation) return;
    setPartStatus(participation.status ?? "pending");
    setPartTShirt(participation.t_shirt_size ?? "");
    setPartAdditionalPaid(!!participation.additional_paid_shirt);
    setPartConsent(!!participation.public_website_consent);
    setPartTeams((participation.teams ?? []).map((t: { id: string }) => t.id));
  }, [participation]);

  const pronounInputRef = useRef<HTMLInputElement>(null);
  const languageInputRef = useRef<HTMLInputElement>(null);
  const tshirtInputRef = useRef<HTMLInputElement>(null);

  const pronounChoices = Array.from(
    new Set([
      ...(config.pronoun_options ?? []),
      ...(record?.pronouns ? [record.pronouns] : []),
    ]),
  ).map((value) => ({ id: value, name: value }));

  const languageChoices = Array.from(
    new Set([...(config.language_options ?? []), ...(record?.languages ?? [])]),
  ).map((value) => ({ id: value, name: value }));

  const tshirtSizeChoices = Array.from(
    new Set([
      ...(config.tshirt_size_options ?? []),
      ...(partTShirt ? [partTShirt] : []),
    ]),
  ).map((value) => ({ id: value, name: value }));

  useEffect(() => {
    if (isPending || !record) return;
    const isAdmin = Array.isArray(permissions) && permissions.includes("admin");
    if (!isAdmin && record.id !== identity?.id) {
      notify("Access denied", { type: "error" });
      redirect("/organizers");
    }
  }, [record, identity, permissions, isPending, redirect, notify]);

  const handleSaveParticipation = async () => {
    if (!record?.id || !activeYear) return;
    setSavingPart(true);
    try {
      const payload = {
        organizer_id: record.id,
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

  return (
    <>
      <SimpleForm
        defaultValues={(rec: { avatar_url?: string }) => ({
          media_upload: rec?.avatar_url ? { src: rec.avatar_url } : undefined,
        })}
      >
        <Card sx={{ width: "100%", mb: 2 }}>
          <CardHeader title="Profile" />
          <CardContent>
            <Stack spacing={2}>
              <TextInput source="handle" />
              <TextInput source="realname" />
              <AutocompleteInput
                source="pronouns"
                choices={pronounChoices}
                TextFieldProps={{ inputRef: pronounInputRef }}
                onCreate={() => {
                  const newValue = pronounInputRef.current?.value?.trim() ?? "";
                  if (newValue) {
                    const newChoice = { id: newValue, name: newValue };
                    pronounChoices.push(newChoice);
                    return newChoice;
                  }
                  // eslint-disable-next-line no-alert
                  const prompted = prompt("Enter a new pronoun");
                  if (!prompted) return undefined;
                  const newChoice = { id: prompted, name: prompted };
                  pronounChoices.push(newChoice);
                  return newChoice;
                }}
                autoSelect
              />
              <AutocompleteArrayInput
                source="languages"
                choices={languageChoices}
                TextFieldProps={{ inputRef: languageInputRef }}
                onCreate={() => {
                  const raw = languageInputRef.current?.value?.trim() ?? "";
                  const newValue = raw.toLowerCase();
                  if (newValue && newValue.length === 2) {
                    const newChoice = { id: newValue, name: newValue };
                    languageChoices.push(newChoice);
                    return newChoice;
                  }
                  // eslint-disable-next-line no-alert
                  const prompted = prompt(
                    "Enter a new language (2-letter code)",
                  );
                  if (!prompted) return undefined;
                  const normalized = prompted.trim().toLowerCase();
                  if (normalized.length !== 2) return undefined;
                  const newChoice = { id: normalized, name: normalized };
                  languageChoices.push(newChoice);
                  return newChoice;
                }}
                translateChoice={false}
                autoSelect
              />
            </Stack>
          </CardContent>
        </Card>

        <Card sx={{ width: "100%", mb: 2 }}>
          <CardHeader title="Contact" />
          <CardContent>
            <Stack spacing={2}>
              <TextInput source="email" />
              <TextInput source="phone" />
              <TextInput source="slack_handle" />
              <TextInput
                source="slack_profile_url"
                helperText="Click on your profile, click the three dots button and choose 'copy profile link'."
              />
              <NumberInput source="demozoo_id" helperText="Demozoo scener ID" />
            </Stack>
          </CardContent>
        </Card>

        <Card sx={{ width: "100%", mb: 2 }}>
          <CardHeader title="Avatar" />
          <CardContent>
            <Stack spacing={2}>
              <ImageInput
                source="media_upload"
                accept={{ "image/*": [".png", ".jpg", ".jpeg"] }}
                label="Avatar"
                placeholder={<p>Drag image here or click to upload</p>}
              >
                <ImageField source="src" title="title" />
              </ImageInput>
            </Stack>
          </CardContent>
        </Card>
      </SimpleForm>

      {activeYear && (
        <Card sx={{ width: "100%", mb: 2 }}>
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
              <Select
                multiple
                label="Teams"
                value={partTeams}
                onChange={(e) => setPartTeams(e.target.value as string[])}
                sx={{ minWidth: 200 }}
              >
                {teamChoices.map((t) => (
                  <MenuItem key={t.id} value={t.id}>
                    {t.name}
                  </MenuItem>
                ))}
              </Select>
              <Button
                variant="contained"
                onClick={handleSaveParticipation}
                disabled={savingPart}
              >
                {savingPart ? "Saving…" : "Save participation"}
              </Button>
            </Stack>
          </CardContent>
        </Card>
      )}
    </>
  );
};

const OrganizerEdit = () => (
  <Edit redirect={false} mutationMode="pessimistic">
    <OrganizerEditForm />
  </Edit>
);

export default OrganizerEdit;
