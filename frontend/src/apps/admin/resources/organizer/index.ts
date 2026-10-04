import EventIcon from "@mui/icons-material/Event";
import OrganizerList from "./OrganizerList";
import OrganizerCreate from "./OrganizerCreate";
import OrganizerEdit from "./OrganizerEdit";

export default {
  name: "organizers",
  options: { label: "Organizers" },
  list: OrganizerList,
  create: OrganizerCreate,
  edit: OrganizerEdit,
  icon: EventIcon,
};
