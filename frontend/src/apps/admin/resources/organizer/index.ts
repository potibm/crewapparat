import PersonIcon from "@mui/icons-material/Person";
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
  icon: PersonIcon,
  recordRepresentation: "handle",
};
