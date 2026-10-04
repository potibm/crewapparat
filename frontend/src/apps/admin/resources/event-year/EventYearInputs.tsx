import { TextInput, BooleanInput } from "react-admin";

export const EventYearInputs = () => (
  <>
    <TextInput source="year" />
    <BooleanInput source="is_active" />
  </>
);
