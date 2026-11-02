import AssignmentIcon from "@mui/icons-material/Assignment";
import ParticipationList from "./ParticipationList";
import ParticipationCreate from "./ParticipationCreate";
import ParticipationEdit from "./ParticipationEdit";

export default {
  name: "participations",
  options: { label: "Participations" },
  list: ParticipationList,
  create: ParticipationCreate,
  edit: ParticipationEdit,
  icon: AssignmentIcon,
};
