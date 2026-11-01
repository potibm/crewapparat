import { Create, SimpleForm, TextInput, BooleanInput } from "react-admin";

const TeamCreate = () => (
  <Create>
    <SimpleForm>
      <TextInput source="name" />
      <BooleanInput source="has_shifts" label="Needs shift planning" />
    </SimpleForm>
  </Create>
);

export default TeamCreate;
