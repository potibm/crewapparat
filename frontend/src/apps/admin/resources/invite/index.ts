import MailIcon from "@mui/icons-material/Mail";
import InviteList from "./InviteList";
import InviteCreate from "./InviteCreate";
import InviteEdit from "./InviteEdit";

export default {
  name: "invites",
  options: { label: "Invites" },
  list: InviteList,
  create: InviteCreate,
  edit: InviteEdit,
  icon: MailIcon,
};
