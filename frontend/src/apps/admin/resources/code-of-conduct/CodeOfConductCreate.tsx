import { Create, SimpleForm, SelectInput, TextInput } from "react-admin";
import { MarkdownInput } from "../../components/inputs/MarkdownInput";
import { useLatestPublishedCodeOfConduct } from "./useLatestPublishedCodeOfConduct";

const CodeOfConductCreateForm = () => {
  const { data, isPending } = useLatestPublishedCodeOfConduct();

  // Render the form only after the lookup resolved: defaultValues are applied
  // on the form's first mount and would otherwise stay empty.
  if (isPending) return null;

  const previous = data?.[0];

  return (
    <SimpleForm
      defaultValues={{ status: "pending", text: previous?.text ?? "" }}
    >
      <SelectInput
        source="status"
        choices={[
          { id: "pending", name: "Pending" },
          { id: "published", name: "Published" },
        ]}
        defaultValue="pending"
      />
      <MarkdownInput
        source="text"
        label="Text (Markdown)"
        helperText={
          previous
            ? "Prefilled from the most recent published version."
            : undefined
        }
      />
      <TextInput
        source="change_note"
        label="Changes"
        multiline
        helperText="What changed compared to the previous version?"
      />
    </SimpleForm>
  );
};

const CodeOfConductCreate = () => (
  <Create>
    <CodeOfConductCreateForm />
  </Create>
);

export default CodeOfConductCreate;
