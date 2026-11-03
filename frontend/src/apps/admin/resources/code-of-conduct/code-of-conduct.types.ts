import { RaRecord } from "react-admin";

export interface CodeOfConductRecord extends RaRecord {
  id: string;
  status: "pending" | "published";
  change_note?: string;
  text: string;
  published_at?: string;
}
