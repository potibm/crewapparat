import { TextInput } from "react-admin";

export const OrganizerInputs = () => (
  <>
    <TextInput source="handle" />
    <TextInput source="realname" />
    <TextInput source="email" />
    <TextInput source="phone" />
    <TextInput source="slack_handle" />
    <TextInput source="pronouns" />
  </>
);
