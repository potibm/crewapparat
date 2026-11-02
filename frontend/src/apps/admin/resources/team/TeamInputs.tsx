import { BooleanInput, TextInput } from "react-admin";

export const TeamInputs = () => (
  <>
    <TextInput source="name" />
    <TextInput
      source="color"
      type="color"
      helperText="Pastel hex color for the team chip, e.g. #FFD6A5"
    />
    <BooleanInput source="has_shifts" label="Needs shift planning" />
    <BooleanInput source="active" label="Active (can be selected by users)" />
  </>
);
