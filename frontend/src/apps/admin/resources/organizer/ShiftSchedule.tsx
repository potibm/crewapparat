import {
  Box,
  IconButton,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Tooltip,
  Typography,
} from "@mui/material";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";
import FavoriteIcon from "@mui/icons-material/Favorite";
import CancelIcon from "@mui/icons-material/Cancel";
import EventIcon from "@mui/icons-material/Event";
import {
  applyShiftStatus,
  buildSchedule,
  nextShiftStatus,
  shiftStatus,
  type ShiftBlock,
  type ShiftStatus,
} from "./shiftSchedule.logic";

export interface TimetableOverlayEvent {
  id: number;
  title: string;
  start_time: string;
  end_time: string;
}

export interface ShiftScheduleProps {
  /** ISO 8601 start of block 0. The wall clock is used, the offset is ignored. */
  start: string;
  blockCount: number;
  preferred: number[];
  unavailable: number[];
  onChange: (preferred: number[], unavailable: number[]) => void;
  timetableEvents?: TimetableOverlayEvent[];
}

const PREFERRED_COLOR = "#e91e63";

const timeFormat = new Intl.DateTimeFormat("en-GB", {
  hour: "2-digit",
  minute: "2-digit",
  hour12: false,
});

const statusConfig: Record<
  ShiftStatus,
  { icon: typeof CheckCircleIcon; color: string; label: string; next: string }
> = {
  ok: {
    icon: CheckCircleIcon,
    color: "success.main",
    label: "Available",
    next: "mark as preferred",
  },
  preferred: {
    icon: FavoriteIcon,
    color: PREFERRED_COLOR,
    label: "Preferred",
    next: "mark as unavailable",
  },
  unavailable: {
    icon: CancelIcon,
    color: "error.main",
    label: "Unavailable",
    next: "reset to available",
  },
};

const ShiftCell = ({
  block,
  status,
  onToggle,
  timetableEvents,
}: {
  block: ShiftBlock;
  status: ShiftStatus;
  onToggle: () => void;
  timetableEvents?: TimetableOverlayEvent[];
}) => {
  const config = statusConfig[status];
  const Icon = config.icon;
  const time = timeFormat.format(block.date);

  const blockStart = block.date.getTime();
  const blockEnd = blockStart + 30 * 60 * 1000;
  const overlapping = (timetableEvents ?? []).filter((event) => {
    const start = new Date(event.start_time).getTime();
    const end = event.end_time
      ? new Date(event.end_time).getTime()
      : start + 30 * 60 * 1000;
    return start < blockEnd && end > blockStart;
  });

  const timetableTooltip =
    overlapping.length > 0
      ? overlapping
          .map((event) => {
            const start = timeFormat.format(new Date(event.start_time));
            const end = event.end_time
              ? timeFormat.format(new Date(event.end_time))
              : "";
            return end ? `${event.title} (${start}–${end})` : `${event.title} (${start})`;
          })
          .join("\n")
      : "";

  return (
    <Box sx={{ display: "inline-flex", alignItems: "center", gap: 0.25 }}>
      <Tooltip title={`${time} · ${config.label} — click to ${config.next}`}>
        <IconButton
          size="small"
          onClick={onToggle}
          aria-label={`${time} ${config.label}`}
          sx={{ p: 0.25 }}
        >
          <Icon fontSize="small" sx={{ color: config.color }} />
        </IconButton>
      </Tooltip>
      {overlapping.length > 0 && (
        <Tooltip title={timetableTooltip}>
          <EventIcon fontSize="small" sx={{ color: "info.main" }} />
        </Tooltip>
      )}
    </Box>
  );
};

const Legend = () => (
  <Stack direction="row" spacing={2} sx={{ flexWrap: "wrap", gap: 1 }}>
    {(Object.keys(statusConfig) as ShiftStatus[]).map((status) => {
      const config = statusConfig[status];
      const Icon = config.icon;

      return (
        <Stack
          key={status}
          direction="row"
          spacing={0.5}
          sx={{ alignItems: "center" }}
        >
          <Icon fontSize="small" sx={{ color: config.color }} />
          <Typography variant="caption" color="text.secondary">
            {config.label}
          </Typography>
        </Stack>
      );
    })}
    <Typography variant="caption" color="text.secondary">
      Click a cell to cycle through the states.
    </Typography>
  </Stack>
);

export const ShiftSchedule = ({
  start,
  blockCount,
  preferred,
  unavailable,
  onChange,
  timetableEvents,
}: ShiftScheduleProps) => {
  const { days, rows } = buildSchedule(start, blockCount);

  if (days.length === 0) {
    return (
      <Typography variant="body2" color="text.secondary">
        Shift planning is not configured.
      </Typography>
    );
  }

  const handleToggle = (block: number) => {
    const current = shiftStatus(block, preferred, unavailable);
    const next = nextShiftStatus(current);
    const result = applyShiftStatus(block, next, preferred, unavailable);

    onChange(result.preferred, result.unavailable);
  };

  return (
    <Box>
      <Legend />
      <TableContainer sx={{ mt: 1, overflowX: "auto" }}>
        <Table size="small" sx={{ minWidth: 480 }}>
          <TableHead>
            <TableRow>
              <TableCell>Time</TableCell>
              {days.map((day) => (
                <TableCell key={day.key} align="center">
                  {day.label}
                </TableCell>
              ))}
            </TableRow>
          </TableHead>
          <TableBody>
            {rows.map((row) => (
              <TableRow key={row.key}>
                <TableCell
                  component="th"
                  scope="row"
                  sx={{ whiteSpace: "nowrap" }}
                >
                  {row.label}
                </TableCell>
                {row.cells.map((cell) => (
                  <TableCell key={cell.dayKey} align="center" padding="none">
                    <Box
                      sx={{
                        display: "flex",
                        justifyContent: "center",
                        gap: 0.25,
                      }}
                    >
                      {cell.blocks.map((block) => (
                        <ShiftCell
                          key={block.index}
                          block={block}
                          status={shiftStatus(
                            block.index,
                            preferred,
                            unavailable,
                          )}
                          onToggle={() => handleToggle(block.index)}
                          timetableEvents={timetableEvents}
                        />
                      ))}
                    </Box>
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
};

export default ShiftSchedule;
