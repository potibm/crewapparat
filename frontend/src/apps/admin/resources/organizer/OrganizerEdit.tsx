import { useEffect, useState } from "react";
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
} from "react-admin";
import { Card, CardContent, CardHeader, Stack } from "@mui/material";
import { useAppConfig } from "@core/config/useConfig";
import { ParticipationSection } from "./ParticipationSection";

const OrganizerEditForm = () => {
  const { record, isPending } = useEditContext();
  const { data: identity } = useGetIdentity();
  const { permissions } = usePermissions();
  const redirect = useRedirect();
  const notify = useNotify();
  const config = useAppConfig();

  const { data: eventYears } = useGetList("event-years", {
    pagination: { page: 1, perPage: 100 },
  });
  const activeYear = eventYears?.find(
    (y: { is_active?: boolean; year: string }) => y.is_active,
  )?.year;

  const [extraPronouns, setExtraPronouns] = useState<string[]>([]);
  const [extraLanguages, setExtraLanguages] = useState<string[]>([]);

  const pronounChoices = Array.from(
    new Set([
      ...(config.pronoun_options ?? []),
      ...extraPronouns,
      ...(record?.pronouns ? [record.pronouns] : []),
    ]),
  ).map((value) => ({ id: value, name: value }));

  const languageChoices = Array.from(
    new Set([
      ...(config.language_options ?? []),
      ...extraLanguages,
      ...(record?.languages ?? []),
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
                onCreate={(filter) => {
                  const newValue = (filter ?? "").trim();
                  if (!newValue) return undefined;
                  setExtraPronouns((prev) =>
                    prev.includes(newValue) ? prev : [...prev, newValue],
                  );
                  return { id: newValue, name: newValue };
                }}
                autoSelect
              />
              <AutocompleteArrayInput
                source="languages"
                choices={languageChoices}
                onCreate={(filter) => {
                  const newValue = (filter ?? "").trim().toLowerCase();
                  if (newValue.length !== 2) return undefined;
                  setExtraLanguages((prev) =>
                    prev.includes(newValue) ? prev : [...prev, newValue],
                  );
                  return { id: newValue, name: newValue };
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

      {activeYear && record?.id && (
        <ParticipationSection organizerId={record.id} activeYear={activeYear} />
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
