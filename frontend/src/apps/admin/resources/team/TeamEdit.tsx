import { Edit, SimpleForm, TextInput, BooleanInput } from "react-admin";

const TeamEdit = () => (
  <Edit>
    <SimpleForm>
      <TextInput source="name" />
      <TextInput
        source="color"
        type="color"
        helperText="Pastel hex color for the team chip, e.g. #FFD6A5"
      />
      <BooleanInput source="has_shifts" label="Needs shift planning" />
    </SimpleForm>
  </Edit>
);

export default TeamEdit;
