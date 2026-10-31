import EventIcon from "@mui/icons-material/Event";
import OrganizerList from "./OrganizerList";
import OrganizerCreate from "./OrganizerCreate";
import OrganizerEdit from "./OrganizerEdit";
import OrganizerShow from "./OrganizerShow";

export default {
  name: "organizers",
  options: { label: "Organizers" },
  list: OrganizerList,
  create: OrganizerCreate,
  edit: OrganizerEdit,
  show: OrganizerShow,
  icon: EventIcon,
};
