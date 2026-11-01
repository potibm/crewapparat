import { RaRecord } from "react-admin";

export interface TeamRecord extends RaRecord {
  id: string;
  name: string;
  color?: string;
  has_shifts: boolean;
}
