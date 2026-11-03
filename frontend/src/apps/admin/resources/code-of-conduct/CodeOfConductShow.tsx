import { Show, SimpleShowLayout, TextField } from "react-admin";
import { MarkdownField } from "../../components/fields/MarkdownField";
import { CurrentVersionBadge } from "./CurrentVersionBadge";
import { CodeOfConductAcceptances } from "./CodeOfConductAcceptances";

const CodeOfConductShow = () => (
  <Show>
    <>
      <SimpleShowLayout>
        <TextField source="version" />
        <TextField source="status" />
        <CurrentVersionBadge label="Current" />
        <TextField source="published_at" />
        <TextField source="accepted_count" label="Accepted" />
        <MarkdownField source="text" />
        <TextField source="change_note" label="Changes" />
      </SimpleShowLayout>
      <CodeOfConductAcceptances />
    </>
  </Show>
);

export default CodeOfConductShow;
