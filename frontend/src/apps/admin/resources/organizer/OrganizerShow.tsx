import { Show, SimpleShowLayout, TextField } from "react-admin";

const OrganizerShow = () => (
  <Show>
    <SimpleShowLayout>
      <TextField source="handle" />
      <TextField source="realname" />
      <TextField source="email" />
      <TextField source="phone" />
      <TextField source="slack_handle" />
      <TextField source="pronouns" />
    </SimpleShowLayout>
  </Show>
);

export default OrganizerShow;
