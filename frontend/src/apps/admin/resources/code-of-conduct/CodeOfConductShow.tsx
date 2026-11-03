import { Show, SimpleShowLayout, TextField } from "react-admin";
import { MarkdownField } from "../../components/fields/MarkdownField";
import { CurrentVersionBadge } from "./CurrentVersionBadge";

const CodeOfConductShow = () => (
  <Show>
    <SimpleShowLayout>
      <TextField source="id" />
      <TextField source="status" />
      <CurrentVersionBadge label="Version" />
      <TextField source="published_at" />
      <MarkdownField source="text" />
      <TextField source="change_note" label="Changes" />
    </SimpleShowLayout>
  </Show>
);

export default CodeOfConductShow;
