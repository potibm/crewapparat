import { RaRecord } from "react-admin";

export interface OrganizerRecord extends RaRecord {
  id: string;
  handle: string;
  realname: string;
  email: string;
  phone: string;
  slack_handle?: string;
  pronouns?: string;
}
