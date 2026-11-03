import { ResourceProps } from "react-admin";
import list from "./CodeOfConductList";
import show from "./CodeOfConductShow";
import edit from "./CodeOfConductEdit";
import create from "./CodeOfConductCreate";
import ArticleIcon from "@mui/icons-material/Article";

const codeOfConducts: ResourceProps = {
  name: "code-of-conducts",
  options: { label: "Code of conduct" },
  list,
  show,
  edit,
  create,
  icon: ArticleIcon,
  recordRepresentation: "version",
};

export default codeOfConducts;
