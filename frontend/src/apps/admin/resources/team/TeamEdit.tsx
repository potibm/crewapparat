import { Edit, SimpleForm, TextInput, BooleanInput } from "react-admin";

const TeamEdit = () => (
  <Edit>
    <SimpleForm>
      <TextInput source="name" />
      <BooleanInput source="has_shifts" label="Needs shift planning" />
    </SimpleForm>
  </Edit>
);

export default TeamEdit;
