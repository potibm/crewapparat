import {
  Create,
  SimpleForm,
  TextInput,
  NumberInput,
  ImageInput,
  ImageField,
  AutocompleteInput,
  AutocompleteArrayInput,
} from "react-admin";
import { useAppConfig } from "@core/config/useConfig";

const OrganizerCreate = () => {
  const config = useAppConfig();

  const pronounChoices = (config.pronoun_options ?? []).map((value) => ({
    id: value,
    name: value,
  }));
  const languageChoices = (config.language_options ?? []).map((value) => ({
    id: value,
    name: value,
  }));

  return (
    <Create>
      <SimpleForm defaultValues={{ media_upload: undefined }}>
        <TextInput source="handle" />
        <TextInput source="realname" />
        <TextInput source="email" />
        <TextInput source="phone" />
        <TextInput source="slack_handle" />
        <AutocompleteInput
          source="pronouns"
          choices={pronounChoices}
          onCreate={(filter?: string) =>
            filter ? { id: filter, name: filter } : undefined
          }
        />
        <AutocompleteArrayInput
          source="languages"
          choices={languageChoices}
          onCreate={(filter?: string) =>
            filter ? { id: filter, name: filter } : undefined
          }
          translateChoice={false}
        />
        <ImageInput
          source="media_upload"
          accept={{ "image/*": [".png", ".jpg", ".jpeg"] }}
          label="Avatar"
          placeholder={<p>Drag image here or click to upload</p>}
        >
          <ImageField source="src" title="title" />
        </ImageInput>
        <NumberInput source="demozoo_id" helperText="Demozoo scener ID" />
      </SimpleForm>
    </Create>
  );
};

export default OrganizerCreate;
