import { useEffect, useRef } from "react";
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
} from "react-admin";
import { Card, CardContent, CardHeader, Stack } from "@mui/material";
import { useAppConfig } from "@core/config/useConfig";

const OrganizerEditForm = () => {
  const { record, isPending } = useEditContext();
  const { data: identity } = useGetIdentity();
  const { permissions } = usePermissions();
  const redirect = useRedirect();
  const notify = useNotify();
  const config = useAppConfig();

  const pronounInputRef = useRef<HTMLInputElement>(null);
  const languageInputRef = useRef<HTMLInputElement>(null);

  const pronounChoices = Array.from(
    new Set([
      ...(config.pronoun_options ?? []),
      ...(record?.pronouns ? [record.pronouns] : []),
    ]),
  ).map((value) => ({ id: value, name: value }));

  const languageChoices = Array.from(
    new Set([...(config.language_options ?? []), ...(record?.languages ?? [])]),
  ).map((value) => ({ id: value, name: value }));

  useEffect(() => {
    if (isPending || !record) return;
    const isAdmin = Array.isArray(permissions) && permissions.includes("admin");
    if (!isAdmin && record.id !== identity?.id) {
      notify("Access denied", { type: "error" });
      redirect("/organizers");
    }
  }, [record, identity, permissions, isPending, redirect, notify]);

  return (
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
                const prompted = prompt("Enter a new language (2-letter code)");
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
  );
};

const OrganizerEdit = () => (
  <Edit>
    <OrganizerEditForm />
  </Edit>
);

export default OrganizerEdit;
