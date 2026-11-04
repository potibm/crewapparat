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
import { DemozooScenerInput } from "./DemozooScenerInput";

const validateSlackProfileUrl = (value?: string) =>
  value && !/^https:\/\/[a-z0-9-]+\.slack\.com\/team\/[A-Za-z0-9]+$/.test(value)
    ? "Must be https://<workspace>.slack.com/team/<id>"
    : undefined;

const validateHttpUrl = (value?: string) => {
  if (!value) return undefined;
  try {
    const u = new URL(value);
    if (u.protocol !== "http:" && u.protocol !== "https:")
      return "Must be a valid http(s) URL";
  } catch {
    return "Must be a valid URL";
  }
  return undefined;
};

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
            if (newValue?.length === 2) {
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
        <TextInput
          source="slack_profile_url"
          validate={validateSlackProfileUrl}
        />
        <TextInput source="mediawiki_profile_url" validate={validateHttpUrl} />
        <NumberInput
          source="first_year_as_organizer"
          helperText="Year of first participation as organizer for this event"
          validate={(value) => {
            if (value == null || value === "") return undefined;
            const n = Number(value);
            const max = new Date().getFullYear() + 1;
            if (n < 1980 || n > max) return `Must be between 1980 and ${max}`;
            return undefined;
          }}
        />
        <DemozooScenerInput />
      </SimpleForm>
    </Create>
  );
};

export default OrganizerCreate;
