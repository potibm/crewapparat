import { RaRecord } from "react-admin";

export interface InviteRecord extends RaRecord {
  id: string;
  email: string;
  status: "sent" | "pending_review" | "approved" | "rejected";
  expires_at: string;
  provided_name?: string;
  provided_phone?: string;
  notes?: string;
  coc_accepted: boolean;
}
