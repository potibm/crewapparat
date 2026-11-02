import { RaRecord } from "react-admin";

export interface EventYearRecord extends RaRecord {
  id: string;
  year: string;
  is_active: boolean;
}
