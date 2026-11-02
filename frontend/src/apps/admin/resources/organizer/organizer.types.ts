import { RaRecord } from "react-admin";
import { ParticipationRecord } from "../participation/participation.types";

export interface OrganizerRecord extends RaRecord {
  id: string;
  oidc_sub?: string;
  handle: string;
  slug: string;
  realname: string;
  email: string;
  phone: string;
  slack_handle?: string;
  slack_profile_url?: string;
  avatar_url?: string;
  demozoo_id?: number;
  pronouns?: string;
  languages?: string[];
  participations?: ParticipationRecord[];
  current_participation_status?: string;
  current_participation_teams?: {
    id: string;
    name: string;
    color?: string;
    active?: boolean;
  }[];
}
