import { useState } from "react";
import {
  List,
  DataTable,
  TextInput,
  EditButton,
  DeleteButton,
} from "react-admin";
import { AvatarLightbox } from "./AvatarLightbox";
import { OrganizerAvatar } from "./OrganizerAvatar";

const organizerFilters = [
  <TextInput key="q" source="q" label="Search" alwaysOn />,
];

const OrganizerListContent = () => {
  const [lightboxSrc, setLightboxSrc] = useState<string | null>(null);

  return (
    <>
      <List filters={organizerFilters}>
        <DataTable bulkActionButtons={false} rowClick="show">
          <DataTable.Col
            label="Avatar"
            render={(record: {
              handle?: string;
              realname?: string;
              avatar_url?: string;
            }) => {
              return (
                <OrganizerAvatar
                  src={record.avatar_url}
                  handle={record.handle}
                  realname={record.realname}
                  size={40}
                  onClick={() => {
                    if (record.avatar_url) setLightboxSrc(record.avatar_url);
                  }}
                />
              );
            }}
          />
          <DataTable.Col source="handle" />
          <DataTable.Col source="realname" />
          <DataTable.Col source="email" />
          <DataTable.Col source="phone" />
          <DataTable.Col>
            <EditButton />
            <DeleteButton
              confirmTitle="Delete organizer"
              confirmContent="Are you sure you want to delete this organizer?"
            />
          </DataTable.Col>
        </DataTable>
      </List>
      <AvatarLightbox
        src={lightboxSrc}
        open={lightboxSrc !== null}
        onClose={() => setLightboxSrc(null)}
      />
    </>
  );
};

const OrganizerList = () => <OrganizerListContent />;

export default OrganizerList;
