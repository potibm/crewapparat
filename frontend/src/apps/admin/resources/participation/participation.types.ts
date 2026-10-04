import { RaRecord } from "react-admin";

export interface ParticipationRecord extends RaRecord {
  id: string;
  organizer_id: string;
  event_year_id: string;
  status: "pending" | "confirmed" | "declined";
  t_shirt_size?: string;
  additional_paid_shirt: boolean;
  public_website_consent: boolean;
}
