import {
  Show,
  SimpleShowLayout,
  TextField,
  useRecordContext,
} from "react-admin";

const OrganizerTitle = () => {
  const record = useRecordContext();
  return <span>Organizer {record?.handle}</span>;
};

const OrganizerShow = () => (
  <Show title={<OrganizerTitle />}>
    <SimpleShowLayout>
      <TextField source="handle" />
      <TextField source="slug" />
      <TextField source="realname" />
      <TextField source="email" />
      <TextField source="phone" />
      <TextField source="slack_handle" />
      <TextField source="pronouns" />
    </SimpleShowLayout>
  </Show>
);

export default OrganizerShow;
