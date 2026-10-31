import { useEffect } from "react";
import {
  Edit,
  SimpleForm,
  TextInput,
  SelectInput,
  SelectArrayInput,
  NumberInput,
  useEditContext,
  useGetIdentity,
  usePermissions,
  useRedirect,
  useNotify,
} from "react-admin";
import { Card, CardContent, CardHeader, Stack } from "@mui/material";

const pronounChoices = [
  { id: "", name: "—" },
  { id: "she/her", name: "she/her" },
  { id: "he/him", name: "he/him" },
  { id: "they/them", name: "they/them" },
];

const languageChoices = [
  { id: "de", name: "German (de)" },
  { id: "en", name: "English (en)" },
  { id: "da", name: "Danish (da)" },
  { id: "nl", name: "Dutch (nl)" },
  { id: "fr", name: "French (fr)" },
  { id: "sv", name: "Swedish (sv)" },
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
          <Stack spacing={1}>
            <TextInput source="handle" />
            <TextInput source="realname" />
            <SelectInput source="pronouns" choices={pronounChoices} />
            <SelectArrayInput source="languages" choices={languageChoices} />
          </Stack>
        </CardContent>
      </Card>

      <Card sx={{ width: "100%", mb: 2 }}>
        <CardHeader title="Contact" />
        <CardContent>
          <Stack spacing={1}>
            <TextInput source="email" />
            <TextInput source="phone" />
            <TextInput source="slack_handle" />
            <TextInput source="slack_profile_url" />
          </Stack>
        </CardContent>
      </Card>

      <Card sx={{ width: "100%", mb: 2 }}>
        <CardHeader title="Links & Meta" />
        <CardContent>
          <Stack spacing={1}>
            <TextInput source="avatar_url" />
            <NumberInput source="demozoo_id" />
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
