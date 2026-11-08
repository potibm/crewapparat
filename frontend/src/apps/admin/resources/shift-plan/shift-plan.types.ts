import { RaRecord } from "react-admin";

export interface ShiftPlanSlotRecord {
  slot: number;
  main_organizer_id?: string;
  organizers: string[];
}

export interface ShiftPlanRecord extends RaRecord {
  id: string;
  event_year_id: string;
  version: string;
  status: "pending" | "published";
  generated_at?: string;
  slots: ShiftPlanSlotRecord[];
}