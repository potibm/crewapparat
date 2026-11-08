import ScheduleIcon from "@mui/icons-material/Schedule";
import ShiftPlanList from "./ShiftPlanList";
import ShiftPlanShow from "./ShiftPlanShow";

export default {
  name: "shift-plans",
  options: { label: "Shift plan" },
  list: ShiftPlanList,
  show: ShiftPlanShow,
  icon: ScheduleIcon,
  recordRepresentation: "version",
};