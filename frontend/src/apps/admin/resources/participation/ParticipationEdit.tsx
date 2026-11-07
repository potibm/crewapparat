import { useState } from "react";
import {
  Edit,
  SimpleForm,
  TextInput,
  SelectInput,
  BooleanInput,
} from "react-admin";
import { useAppConfig } from "@core/config/useConfig";

const ParticipationEdit = () => {
  const config = useAppConfig();
  const deadline = config.tshirt_deadline;
  const [pastDeadline] = useState(
    () =>
      !!deadline &&
      !Number.isNaN(Date.parse(deadline)) &&
      Date.now() > Date.parse(deadline),
  );

  return (
    <Edit>
      <SimpleForm>
        <TextInput source="organizer_id" />
        <TextInput source="event_year_id" />
        <SelectInput
          source="status"
          choices={[
            { id: "confirmed", name: "Confirmed" },
            { id: "declined", name: "Declined" },
          ]}
        />
        <TextInput
          source="t_shirt_size"
          helperText={
            deadline
              ? `Changes allowed until ${new Date(deadline).toLocaleString()}`
              : undefined
          }
          disabled={pastDeadline}
        />
        <BooleanInput
          source="additional_paid_shirt"
          disabled={pastDeadline}
        />
        <BooleanInput source="public_website_consent" />
      </SimpleForm>
    </Edit>
  );
};

export default ParticipationEdit;