import { FieldProps, useRecordContext } from "react-admin";
import MarkdownPreview from "@uiw/react-markdown-preview";
import "@uiw/react-markdown-preview/markdown.css";
import { useTheme } from "@mui/material";

export interface MarkdownFieldProps extends FieldProps {
  source: string;
}

export const MarkdownField = ({ source }: MarkdownFieldProps) => {
  const record = useRecordContext<Record<string, unknown>>();
  const theme = useTheme();

  const value = record?.[source];

  if (typeof value !== "string" || value === "") return null;

  return (
    <MarkdownPreview
      source={value}
      wrapperElement={{ "data-color-mode": theme.palette.mode }}
    />
  );
};
