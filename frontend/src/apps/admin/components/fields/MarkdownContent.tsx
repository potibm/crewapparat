import MarkdownPreview from "@uiw/react-markdown-preview";
import "@uiw/react-markdown-preview/markdown.css";
import { useTheme } from "@mui/material";

export interface MarkdownContentProps {
  value: string;
}

export const MarkdownContent = ({ value }: MarkdownContentProps) => {
  const theme = useTheme();

  return (
    <MarkdownPreview
      source={value}
      wrapperElement={{ "data-color-mode": theme.palette.mode }}
    />
  );
};
