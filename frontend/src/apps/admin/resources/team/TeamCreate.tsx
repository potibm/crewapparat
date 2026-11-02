import { Create, SimpleForm, TextInput, BooleanInput } from "react-admin";

const TeamCreate = () => (
  <Create>
    <SimpleForm>
      <TextInput source="name" />
      <TextInput
        source="color"
        type="color"
        helperText="Pastel hex color for the team chip, e.g. #FFD6A5"
      />
      <BooleanInput source="has_shifts" label="Needs shift planning" />
      <BooleanInput source="active" label="Active (can be selected by users)" />
    </SimpleForm>
  </Create>
);

export default TeamCreate;
