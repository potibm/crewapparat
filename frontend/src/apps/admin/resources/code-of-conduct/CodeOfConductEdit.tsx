import {
  Edit,
  SimpleForm,
  SimpleShowLayout,
  SelectInput,
  TextField,
  TextInput,
  useRecordContext,
} from "react-admin";
import { Alert, Box } from "@mui/material";
import { MarkdownInput } from "../../components/inputs/MarkdownInput";
import { MarkdownField } from "../../components/fields/MarkdownField";
import type { CodeOfConductRecord } from "./code-of-conduct.types";
import { validateVersion } from "./code-of-conduct.version";

const CodeOfConductEditContent = () => {
  const record = useRecordContext<CodeOfConductRecord>();

  if (record?.status === "published") {
    return (
      <>
        <Box sx={{ p: 2 }}>
          <Alert severity="info">
            This code of conduct is published and can no longer be changed.
            Create a new version instead.
          </Alert>
        </Box>
        <SimpleShowLayout>
          <TextField source="version" />
          <TextField source="status" />
          <TextField source="published_at" />
          <MarkdownField source="text" />
          <TextField source="change_note" label="Changes" />
        </SimpleShowLayout>
      </>
    );
  }

  return (
    <SimpleForm>
      <TextInput
        source="version"
        label="Version"
        validate={validateVersion}
        helperText="Semantic version, e.g. 1.0.0"
      />
      <SelectInput
        source="status"
        choices={[
          { id: "pending", name: "Pending" },
          { id: "published", name: "Published" },
        ]}
      />
      <MarkdownInput source="text" label="Text (Markdown)" />
      <TextInput
        source="change_note"
        label="Changes"
        multiline
        helperText="What changed compared to the previous version?"
      />
    </SimpleForm>
  );
};

const CodeOfConductEdit = () => (
  <Edit>
    <CodeOfConductEditContent />
  </Edit>
);

export default CodeOfConductEdit;
