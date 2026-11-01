import { useRef } from "react";
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

  const pronounInputRef = useRef<HTMLInputElement>(null);
  const languageInputRef = useRef<HTMLInputElement>(null);

  const pronounChoices = Array.from(new Set(config.pronoun_options ?? [])).map(
    (value) => ({ id: value, name: value }),
  );

  const languageChoices = Array.from(
    new Set(config.language_options ?? []),
  ).map((value) => ({ id: value, name: value }));

  return (
    <Create>
      <SimpleForm>
        <TextInput source="handle" />
        <TextInput source="realname" />
        <TextInput source="email" />
        <TextInput source="phone" />
        <TextInput source="slack_handle" />
        <AutocompleteInput
          source="pronouns"
          choices={pronounChoices}
          TextFieldProps={{ inputRef: pronounInputRef }}
          onCreate={() => {
            const newValue = pronounInputRef.current?.value?.trim() ?? "";
            if (newValue) {
              const newChoice = { id: newValue, name: newValue };
              pronounChoices.push(newChoice);
              return newChoice;
            }
            // eslint-disable-next-line no-alert
            const prompted = prompt("Enter a new pronoun");
            if (!prompted) return undefined;
            const newChoice = { id: prompted, name: prompted };
            pronounChoices.push(newChoice);
            return newChoice;
          }}
          autoSelect
        />
        <AutocompleteArrayInput
          source="languages"
          choices={languageChoices}
          TextFieldProps={{ inputRef: languageInputRef }}
          onCreate={() => {
            const raw = languageInputRef.current?.value?.trim() ?? "";
            const newValue = raw.toLowerCase();
            if (newValue && newValue.length === 2) {
              const newChoice = { id: newValue, name: newValue };
              languageChoices.push(newChoice);
              return newChoice;
            }
            // eslint-disable-next-line no-alert
            const prompted = prompt("Enter a new language (2-letter code)");
            if (!prompted) return undefined;
            const normalized = prompted.trim().toLowerCase();
            if (normalized.length !== 2) return undefined;
            const newChoice = { id: normalized, name: normalized };
            languageChoices.push(newChoice);
            return newChoice;
          }}
          translateChoice={false}
          autoSelect
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
