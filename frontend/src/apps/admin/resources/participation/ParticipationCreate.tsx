import {
  Create,
  SimpleForm,
  TextInput,
  SelectInput,
  BooleanInput,
} from "react-admin";

const ParticipationCreate = () => (
  <Create>
    <SimpleForm>
      <TextInput source="organizer_id" />
      <TextInput source="event_year_id" />
      <SelectInput
        source="status"
        choices={[
          { id: "pending", name: "Pending" },
          { id: "confirmed", name: "Confirmed" },
          { id: "declined", name: "Declined" },
        ]}
      />
      <TextInput source="t_shirt_size" />
      <BooleanInput source="additional_paid_shirt" />
      <BooleanInput source="public_website_consent" />
    </SimpleForm>
  </Create>
);

export default ParticipationCreate;
