import { useEffect } from "react";
import {
  Edit,
  SimpleForm,
  TextInput,
  AutocompleteArrayInput,
  NumberInput,
  FileInput,
  ImageField,
  BooleanInput,
  useEditContext,
  useGetIdentity,
  usePermissions,
  useRedirect,
  useNotify,
} from "react-admin";
import { Card, CardContent, CardHeader, Stack } from "@mui/material";
import { PronounsInput } from "./PronounsInput";
import { OrganizerAvatar } from "./OrganizerAvatar";

const languageChoices = [
  { id: "de", name: "German (de)" },
  { id: "en", name: "English (en)" },
  { id: "fr", name: "French (fr)" },
  { id: "hu", name: "Hungarian (hu)" },
  { id: "no", name: "Norwegian (no)" },
  { id: "nl", name: "Dutch (nl)" },
  { id: "ru", name: "Russian (ru)" },
  { id: "pl", name: "Polish (pl)" },
];

const OrganizerEditForm = () => {
  const { record, isPending } = useEditContext();
  const { data: identity } = useGetIdentity();
  const { permissions } = usePermissions();
  const redirect = useRedirect();
  const notify = useNotify();

  useEffect(() => {
    if (isPending || !record) return;
    const isAdmin = Array.isArray(permissions) && permissions.includes("admin");
    if (!isAdmin && record.id !== identity?.id) {
      notify("Access denied", { type: "error" });
      redirect("/organizers");
    }
  }, [record, identity, permissions, isPending, redirect, notify]);

  return (
    <SimpleForm>
      <Card sx={{ width: "100%", mb: 2 }}>
        <CardHeader title="Profile" />
        <CardContent>
          <Stack spacing={2}>
            <TextInput source="handle" />
            <TextInput source="realname" />
            <PronounsInput source="pronouns" />
            <AutocompleteArrayInput
              source="languages"
              choices={languageChoices}
              onCreate={(filter?: string) =>
                filter ? { id: filter, name: filter } : undefined
              }
              translateChoice={false}
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
            {record?.avatar_url && (
              <OrganizerAvatar
                src={record.avatar_url}
                handle={record.handle}
                realname={record.realname}
                size={96}
              />
            )}
            <FileInput
              source="media_upload"
              accept={{ "image/*": [".png", ".jpg", ".jpeg", ".webp"] }}
              label="Avatar"
              placeholder={<p>Drag image here or click to upload</p>}
            >
              <ImageField source="src" title="title" />
            </FileInput>
            {record?.avatar_url && (
              <BooleanInput source="remove_avatar" label="Remove avatar" />
            )}
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
