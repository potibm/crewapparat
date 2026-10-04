import CalendarTodayIcon from "@mui/icons-material/CalendarToday";
import EventYearList from "./EventYearList";
import EventYearCreate from "./EventYearCreate";
import EventYearEdit from "./EventYearEdit";

export default {
  name: "event-years",
  options: { label: "Event Years", id: "year" },
  list: EventYearList,
  create: EventYearCreate,
  edit: EventYearEdit,
  icon: CalendarTodayIcon,
};
