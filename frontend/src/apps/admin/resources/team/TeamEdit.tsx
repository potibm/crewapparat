import { Edit, SimpleForm, TextInput } from "react-admin";

const TeamEdit = () => (
  <Edit>
    <SimpleForm>
      <TextInput source="name" />
    </SimpleForm>
  </Edit>
);

export default TeamEdit;
