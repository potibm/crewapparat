import { Create, SimpleForm, TextInput, required } from "react-admin";
import { DemozooScenerInput } from "../organizer/DemozooScenerInput";

const InviteCreate = () => (
  <Create>
    <SimpleForm>
      <TextInput source="email" validate={required()} />
      <TextInput source="handle" validate={required()} />
      <DemozooScenerInput />
      <TextInput source="reason" multiline />
    </SimpleForm>
  </Create>
);

export default InviteCreate;
