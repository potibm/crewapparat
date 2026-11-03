import { RaRecord } from "react-admin";

export interface CodeOfConductRecord extends RaRecord {
  id: string;
  version: string;
  status: "pending" | "published";
  change_note?: string;
  text: string;
  published_at?: string;
  accepted_count?: number;
}

export interface CodeOfConductAcceptanceRecord {
  id: string;
  organizer_id: string;
  code_of_conduct_id: string;
  accepted_at: string;
  handle: string;
  realname: string;
  email: string;
}
