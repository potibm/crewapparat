import { BooleanInput, TextInput } from "react-admin";

export const TeamInputs = () => (
  <>
    <TextInput source="name" />
    <BooleanInput source="has_shifts" label="Needs shift planning" />
  </>
);
