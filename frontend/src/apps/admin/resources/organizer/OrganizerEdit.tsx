import { useEffect } from "react";
import {
  Edit,
  SimpleForm,
  TextInput,
  AutocompleteArrayInput,
  AutocompleteInput,
  NumberInput,
  ImageInput,
  ImageField,
  BooleanInput,
  useEditContext,
  useGetIdentity,
  usePermissions,
  useRedirect,
  useNotify,
} from "react-admin";
import { Card, CardContent, CardHeader, Stack } from "@mui/material";
import { useAppConfig } from "@core/config/useConfig";
import { OrganizerAvatar } from "./OrganizerAvatar";

const OrganizerEditForm = () => {
  const { record, isPending } = useEditContext();
  const { data: identity } = useGetIdentity();
  const { permissions } = usePermissions();
  const redirect = useRedirect();
  const notify = useNotify();
  const config = useAppConfig();

  const pronounChoices = (config.pronoun_options ?? []).map((value) => ({
    id: value,
    name: value,
  }));
  const languageChoices = (config.language_options ?? []).map((value) => ({
    id: value,
    name: value,
  }));

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
              onCreate={(filter?: string) =>
                filter ? { id: filter, name: filter } : undefined
              }
            />
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
            <ImageInput
              source="media_upload"
              accept={{ "image/*": [".png", ".jpg", ".jpeg"] }}
              label="Avatar"
              placeholder={<p>Drag image here or click to upload</p>}
            >
              <ImageField source="src" title="title" />
            </ImageInput>
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
