import { FieldProps, useRecordContext } from "react-admin";
import { MarkdownContent } from "./MarkdownContent";

export interface MarkdownFieldProps extends FieldProps {
  source: string;
}

export const MarkdownField = ({ source }: MarkdownFieldProps) => {
  const record = useRecordContext<Record<string, unknown>>();

  const value = record?.[source];

  if (typeof value !== "string" || value === "") return null;

  return <MarkdownContent value={value} />;
};
