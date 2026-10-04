import { Edit, SimpleForm, TextInput } from "react-admin";

const OrganizerEdit = () => (
  <Edit>
    <SimpleForm>
      <TextInput source="handle" />
      <TextInput source="realname" />
      <TextInput source="email" />
      <TextInput source="phone" />
      <TextInput source="slack_handle" />
      <TextInput source="pronouns" />
    </SimpleForm>
  </Edit>
);

export default OrganizerEdit;
