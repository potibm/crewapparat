import {
  Create,
  SimpleForm,
  TextInput,
  BooleanInput,
  required,
} from "react-admin";
import { DemozooScenerInput } from "../organizer/DemozooScenerInput";

const InviteCreate = () => (
  <Create>
    <SimpleForm>
      <TextInput source="email" validate={required()} />
      <TextInput source="handle" validate={required()} />
      <DemozooScenerInput />
      <TextInput source="reason" multiline />
      <TextInput source="provided_name" />
      <TextInput source="provided_phone" />
      <TextInput source="notes" />
      <BooleanInput source="coc_accepted" />
    </SimpleForm>
  </Create>
);

export default InviteCreate;
