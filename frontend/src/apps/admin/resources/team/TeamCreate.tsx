import { Create, SimpleForm, TextInput } from "react-admin";

const TeamCreate = () => (
  <Create>
    <SimpleForm>
      <TextInput source="name" />
    </SimpleForm>
  </Create>
);

export default TeamCreate;
