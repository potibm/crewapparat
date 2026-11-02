import { Create, SimpleForm, TextInput, BooleanInput } from "react-admin";

const EventYearCreate = () => (
  <Create>
    <SimpleForm>
      <TextInput source="year" />
      <BooleanInput source="is_active" />
    </SimpleForm>
  </Create>
);

export default EventYearCreate;
