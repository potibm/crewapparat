import { Create, SimpleForm, TextInput } from "react-admin";

const OrganizerCreate = () => (
  <Create>
    <SimpleForm>
      <TextInput source="handle" />
      <TextInput source="realname" />
      <TextInput source="email" />
      <TextInput source="phone" />
      <TextInput source="slack_handle" />
      <TextInput source="pronouns" />
    </SimpleForm>
  </Create>
);

export default OrganizerCreate;
