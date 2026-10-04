import GroupIcon from "@mui/icons-material/Group";
import TeamList from "./TeamList";
import TeamCreate from "./TeamCreate";
import TeamEdit from "./TeamEdit";

export default {
  name: "teams",
  options: { label: "Teams" },
  list: TeamList,
  create: TeamCreate,
  edit: TeamEdit,
  icon: GroupIcon,
};
