import CalendarTodayIcon from "@mui/icons-material/CalendarToday";
import EventYearList from "./EventYearList";

export default {
  name: "event-years",
  options: { label: "Event Years", id: "year" },
  list: EventYearList,
  icon: CalendarTodayIcon,
};
