import {
  List,
  DataTable,
  TextInput,
  SelectInput,
  EditButton,
  DeleteButton,
  useRecordContext,
} from "react-admin";
import { Chip } from "@mui/material";
import type { CodeOfConductRecord } from "./code-of-conduct.types";
import { useLatestPublishedCodeOfConduct } from "./useLatestPublishedCodeOfConduct";

const filters = [
  <TextInput key="q" source="q" label="Search" alwaysOn />,
  <SelectInput
    key="status"
    source="status"
    choices={[
      { id: "pending", name: "Pending" },
      { id: "published", name: "Published" },
    ]}
  />,
];

const CodeOfConductRowActions = () => {
  const record = useRecordContext<CodeOfConductRecord>();
  const isPublished = record?.status === "published";

  return (
    <>
      <EditButton
        disabled={isPublished}
        title={isPublished ? "Published versions cannot be edited" : undefined}
      />
      <DeleteButton
        disabled={isPublished}
        title={isPublished ? "Published versions cannot be deleted" : undefined}
        confirmTitle="Delete code of conduct"
        confirmContent="Are you sure you want to delete this code of conduct?"
      />
    </>
  );
};

const CodeOfConductList = () => {
  const { data } = useLatestPublishedCodeOfConduct();
  const currentId = data?.[0]?.id;

  return (
    <List filters={filters}>
      <DataTable bulkActionButtons={false} rowClick="show">
        <DataTable.Col source="id" />
        <DataTable.Col source="status" />
        <DataTable.Col
          label="Version"
          disableSort
          render={(record: CodeOfConductRecord) =>
            record.id === currentId ? (
              <Chip label="Current" color="success" size="small" />
            ) : null
          }
        />
        <DataTable.Col source="published_at" />
        <DataTable.Col>
          <CodeOfConductRowActions />
        </DataTable.Col>
      </DataTable>
    </List>
  );
};

export default CodeOfConductList;
