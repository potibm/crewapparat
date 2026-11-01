import {
  Create,
  SimpleForm,
  TextInput,
  FileInput,
  ImageField,
} from "react-admin";
import { PronounsInput } from "./PronounsInput";

const OrganizerCreate = () => (
  <Create>
    <SimpleForm>
      <TextInput source="handle" />
      <TextInput source="realname" />
      <TextInput source="email" />
      <TextInput source="phone" />
      <TextInput source="slack_handle" />
      <PronounsInput source="pronouns" />
      <FileInput
        source="media_upload"
        accept={{ "image/*": [".png", ".jpg", ".jpeg"] }}
        label="Avatar"
        placeholder={<p>Drag image here or click to upload</p>}
      >
        <ImageField source="src" title="title" />
      </FileInput>
    </SimpleForm>
  </Create>
);

export default OrganizerCreate;
