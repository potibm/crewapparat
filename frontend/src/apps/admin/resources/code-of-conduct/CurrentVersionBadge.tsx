import { useRecordContext } from "react-admin";
import { Chip, Typography } from "@mui/material";
import type { CodeOfConductRecord } from "./code-of-conduct.types";
import { useLatestPublishedCodeOfConduct } from "./useLatestPublishedCodeOfConduct";

export interface CurrentVersionBadgeProps {
  // Consumed by React Admin's <Labeled> wrapper, not by this component.
  label?: string;
  emptyText?: string;
}

export const CurrentVersionBadge = ({
  emptyText = "—",
}: CurrentVersionBadgeProps) => {
  const record = useRecordContext<CodeOfConductRecord>();
  const { data } = useLatestPublishedCodeOfConduct();

  if (!record) return null;

  if (data?.[0]?.id !== record.id) {
    return (
      <Typography variant="body2" color="text.secondary">
        {emptyText}
      </Typography>
    );
  }

  return <Chip label="Current version" color="success" size="small" />;
};
