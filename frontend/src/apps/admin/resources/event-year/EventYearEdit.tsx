import { Edit, SimpleForm, TextInput, BooleanInput } from "react-admin";

const EventYearEdit = () => (
  <Edit>
    <SimpleForm>
      <TextInput source="year" />
      <BooleanInput source="is_active" />
    </SimpleForm>
  </Edit>
);

export default EventYearEdit;
