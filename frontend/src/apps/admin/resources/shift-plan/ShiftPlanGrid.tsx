import React from "react";
import {
  Box,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Tooltip,
  Typography,
  useTheme,
} from "@mui/material";
import { parseShiftStart } from "../organizer/shiftSchedule.logic";
import type { TimetableOverlayEvent } from "../organizer/ShiftSchedule";

export interface ShiftPlanGridProps {
  start: string;
  blockCount: number;
  slots: { slot: number; main_organizer_id?: string; organizers: string[] }[];
  organizerHandle: (id: string) => string;
  timetableEvents?: TimetableOverlayEvent[];
  currentOrganizerId?: string;
}

const LIGHT_PALETTE = [
  "#f0fdf4",
  "#eff6ff",
  "#fefce8",
  "#fdf2f8",
  "#fff7ed",
  "#faf5ff",
  "#f0fdfa",
  "#fef2f2",
  "#f5f3ff",
  "#ecfdf5",
  "#f0f9ff",
  "#fdf4ff",
  "#fff1f2",
  "#fef8ec",
  "#ecfeff",
  "#f8fafc",
];

// Dark-mode equivalents: same hue family, lowered lightness so they read as
// distinct, muted tints on a dark background instead of bright pastel patches.
const DARK_PALETTE = [
  "#14532d",
  "#1e3a8a",
  "#713f12",
  "#831843",
  "#7c2d12",
  "#581c87",
  "#134e4a",
  "#7f1d1d",
  "#4c1d95",
  "#064e3b",
  "#0c4a6e",
  "#701a75",
  "#881337",
  "#78350f",
  "#164e63",
  "#1e293b",
];

const colorIndexFor = (key: string): number => {
  let sum = 0;
  for (let i = 0; i < key.length; i++) sum += key.charCodeAt(i);
  return sum % 16;
};

const colorFor = (key: string, dark: boolean): string =>
  (dark ? DARK_PALETTE : LIGHT_PALETTE)[colorIndexFor(key)];

const timeFormat = new Intl.DateTimeFormat("en-GB", {
  hour: "2-digit",
  minute: "2-digit",
  hour12: false,
});

interface SlotInfo {
  date: Date;
  dayKey: string;
  clockMinutes: number;
}

const buildSlotInfos = (start: string, blockCount: number): SlotInfo[] => {
  const base = parseShiftStart(start);
  if (!base || blockCount <= 0) return [];
  const out: SlotInfo[] = [];
  for (let i = 0; i < blockCount; i++) {
    const date = new Date(base.getTime() + i * 30 * 60 * 1000);
    out.push({
      date,
      dayKey: `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}`,
      clockMinutes: date.getHours() * 60 + date.getMinutes(),
    });
  }
  return out;
};

const eventTitleForSlot = (
  slotDate: Date,
  events: TimetableOverlayEvent[] | undefined,
): string => {
  if (!events) return "";
  for (const ev of events) {
    const s = parseShiftStart(ev.start_time);
    if (!s) continue;
    const e = ev.end_time ? parseShiftStart(ev.end_time) : null;
    const startMs = s.getTime();
    const endMs = e ? e.getTime() : startMs;
    // Skip point events (start == end): they have no duration.
    if (startMs === endMs) continue;
    if (slotDate.getTime() >= startMs && slotDate.getTime() < endMs) {
      return ev.title;
    }
  }
  return "";
};

const isNight = (clockMinutes: number): boolean =>
  clockMinutes >= 120 && clockMinutes < 480;

const formatClock = (clockMinutes: number): string => {
  const h = Math.floor(clockMinutes / 60);
  const m = clockMinutes % 60;
  return `${String(h).padStart(2, "0")}:${String(m).padStart(2, "0")}`;
};

interface SpanInfo {
  skip: boolean;
  span: number;
}

const computeSpans = (contents: string[]): SpanInfo[] => {
  const n = contents.length;
  const result: SpanInfo[] = Array.from({ length: n }, () => ({
    skip: false,
    span: 1,
  }));
  let i = 0;
  while (i < n) {
    let j = i;
    while (j + 1 < n && contents[j + 1] === contents[i]) j++;
    result[i].span = j - i + 1;
    for (let k = i + 1; k <= j; k++) result[k].skip = true;
    i = j + 1;
  }
  return result;
};

export const ShiftPlanGrid = ({
  start,
  blockCount,
  slots,
  organizerHandle,
  timetableEvents,
  currentOrganizerId,
}: ShiftPlanGridProps) => {
  const theme = useTheme();
  const isDark = theme.palette.mode === "dark";
  const slotInfos = buildSlotInfos(start, blockCount);
  const slotByIndex = new Map(slots.map((s) => [s.slot, s]));

  const days: string[] = [];
  const seenDays = new Set<string>();
  for (const info of slotInfos) {
    if (seenDays.has(info.dayKey)) continue;
    seenDays.add(info.dayKey);
    days.push(info.dayKey);
  }

  const dayLabel = (dayKey: string): string => {
    const info = slotInfos.find((s) => s.dayKey === dayKey);
    if (!info) return dayKey;
    return new Intl.DateTimeFormat("en-GB", { weekday: "short" }).format(
      info.date,
    );
  };

  const rows: number[] = [];
  const seenRows = new Set<number>();
  for (const info of slotInfos) {
    if (seenRows.has(info.clockMinutes)) continue;
    seenRows.add(info.clockMinutes);
    rows.push(info.clockMinutes);
  }
  rows.sort((a, b) => a - b);

  // slot index lookup by (dayKey, clockMinutes)
  const slotIndexAt = (dayKey: string, clockMinutes: number): number | null => {
    const idx = slotInfos.findIndex(
      (s) => s.dayKey === dayKey && s.clockMinutes === clockMinutes,
    );
    return idx >= 0 ? idx : null;
  };

  // content keys per day per column
  const eventContent: Record<string, string[]> = {};
  const mainContent: Record<string, string[]> = {};
  const infodeskContent: Record<string, string[]> = {};

  for (const day of days) {
    eventContent[day] = [];
    mainContent[day] = [];
    infodeskContent[day] = [];
    for (const clock of rows) {
      const idx = slotIndexAt(day, clock);
      if (idx === null) {
        eventContent[day].push("");
        mainContent[day].push("");
        infodeskContent[day].push("");
        continue;
      }
      const slot = slotByIndex.get(idx);
      const slotDate = slotInfos[idx].date;
      const evTitle = eventTitleForSlot(slotDate, timetableEvents);
      const mainHandle = slot?.main_organizer_id
        ? organizerHandle(slot.main_organizer_id)
        : "";
      const infoHandles = (slot?.organizers ?? [])
        .map((id) => organizerHandle(id))
        .sort()
        .join(", ");
      eventContent[day].push(evTitle);
      mainContent[day].push(mainHandle);
      infodeskContent[day].push(infoHandles);
    }
  }

  // spans per day per column
  const eventSpans: Record<string, SpanInfo[]> = {};
  const mainSpans: Record<string, SpanInfo[]> = {};
  const infodeskSpans: Record<string, SpanInfo[]> = {};
  for (const day of days) {
    eventSpans[day] = computeSpans(eventContent[day]);
    mainSpans[day] = computeSpans(mainContent[day]);
    infodeskSpans[day] = computeSpans(infodeskContent[day]);
  }

  if (days.length === 0) {
    return (
      <Typography variant="body2" color="text.secondary">
        Shift planning is not configured.
      </Typography>
    );
  }

  return (
    <TableContainer sx={{ overflowX: "auto" }}>
      <Table
        size="small"
        sx={{
          minWidth: 1100,
          borderCollapse: "collapse",
          tableLayout: "fixed",
          width: "100%",
        }}
      >
        <TableHead>
          <TableRow>
            <TableCell
              rowSpan={2}
              sx={{ fontWeight: 700, width: 56, textAlign: "center" }}
            >
              Time
            </TableCell>
            {days.map((day) => (
              <TableCell
                key={day}
                colSpan={3}
                align="center"
                sx={{ fontWeight: 700, width: 340 }}
              >
                {dayLabel(day)}
              </TableCell>
            ))}
          </TableRow>
          <TableRow>
            {days.map((day) => (
              <React.Fragment key={day}>
                <TableCell
                  align="center"
                  sx={{ fontSize: 10, fontWeight: 700, width: 70 }}
                >
                  Event
                </TableCell>
                <TableCell
                  align="center"
                  sx={{ fontSize: 10, fontWeight: 700, width: 70 }}
                >
                  Main
                </TableCell>
                <TableCell
                  align="center"
                  sx={{ fontSize: 10, fontWeight: 700, width: 200 }}
                >
                  Infodesk
                </TableCell>
              </React.Fragment>
            ))}
          </TableRow>
        </TableHead>
        <TableBody>
          {rows.map((clock, rowIndex) => {
            const night = isNight(clock);
            return (
              <TableRow key={clock}>
                {clock % 60 === 0 && (
                  <TableCell
                    rowSpan={2}
                    sx={{
                      fontFamily: "monospace",
                      fontWeight: 700,
                      textAlign: "center",
                      verticalAlign: "middle",
                      whiteSpace: "nowrap",
                      width: 64,
                      bgcolor: night ? "action.hover" : undefined,
                      borderRight: "1px solid",
                      borderColor: "divider",
                      px: 1,
                    }}
                  >
                    {formatClock(clock)}
                  </TableCell>
                )}
                {days.map((day) => {
                  const idx = slotIndexAt(day, clock);
                  const slotDate = idx !== null ? slotInfos[idx].date : null;

                  // Event cell
                  const eSpan = eventSpans[day][rowIndex];
                  const mSpan = mainSpans[day][rowIndex];
                  const iSpan = infodeskSpans[day][rowIndex];

                  const evTitle =
                    idx !== null ? eventContent[day][rowIndex] : "";
                  const mainHandle =
                    idx !== null ? mainContent[day][rowIndex] : "";
                  const infoHandles =
                    idx !== null ? infodeskContent[day][rowIndex] : "";

                  const cells: React.ReactNode[] = [];

                  if (!eSpan.skip) {
                    cells.push(
                      <TableCell
                        key={`${day}-ev`}
                        rowSpan={eSpan.span > 1 ? eSpan.span : undefined}
                        sx={{
                          verticalAlign: "middle",
                          textAlign: "center",
                          bgcolor: evTitle
                            ? "action.hover"
                            : "background.paper",
                          borderRight: "1px solid",
                          borderColor: "divider",
                          p: 0.5,
                        }}
                      >
                        {evTitle ? (
                          <Typography
                            sx={{
                              fontSize: 10,
                              fontWeight: 600,
                              color: "text.primary",
                              whiteSpace: "normal",
                              wordBreak: "break-word",
                            }}
                          >
                            {evTitle}
                          </Typography>
                        ) : null}
                      </TableCell>,
                    );
                  }

                  if (!mSpan.skip) {
                    const color = mainHandle
                      ? colorFor(mainHandle, isDark)
                      : undefined;
                    cells.push(
                      <TableCell
                        key={`${day}-main`}
                        rowSpan={mSpan.span > 1 ? mSpan.span : undefined}
                        sx={{
                          verticalAlign: "middle",
                          textAlign: "center",
                          bgcolor:
                            color ?? (night ? "action.hover" : undefined),
                          borderRight: "1px solid",
                          borderColor: "divider",
                          p: 0.5,
                        }}
                      >
                        {mainHandle && slotDate ? (
                          <Stack spacing={0.25} sx={{ alignItems: "center" }}>
                            <Typography
                              sx={{
                                fontSize: 9,
                                color: "text.secondary",
                                fontWeight: 700,
                              }}
                            >
                              {timeFormat.format(slotDate)} –{" "}
                              {timeFormat.format(
                                new Date(
                                  slotDate.getTime() +
                                    mSpan.span * 30 * 60 * 1000,
                                ),
                              )}
                            </Typography>
                            <Tooltip title={mainHandle}>
                              <Box
                                sx={{
                                  bgcolor:
                                    idx !== null &&
                                    slotByIndex.get(idx)?.main_organizer_id ===
                                      currentOrganizerId
                                      ? "warning.main"
                                      : "background.paper",
                                  color:
                                    idx !== null &&
                                    slotByIndex.get(idx)?.main_organizer_id ===
                                      currentOrganizerId
                                      ? "warning.contrastText"
                                      : "text.primary",
                                  border: "1px solid",
                                  borderColor: "divider",
                                  borderRadius: 0.5,
                                  px: 0.5,
                                  py: 0.25,
                                  fontSize: 10,
                                  fontWeight: 500,
                                  whiteSpace: "normal",
                                  wordBreak: "break-word",
                                }}
                              >
                                {mainHandle}
                              </Box>
                            </Tooltip>
                          </Stack>
                        ) : null}
                      </TableCell>,
                    );
                  }

                  if (!iSpan.skip) {
                    const color = infoHandles
                      ? colorFor(infoHandles, isDark)
                      : undefined;
                    cells.push(
                      <TableCell
                        key={`${day}-info`}
                        rowSpan={iSpan.span > 1 ? iSpan.span : undefined}
                        sx={{
                          verticalAlign: "middle",
                          textAlign: "center",
                          bgcolor:
                            color ?? (night ? "action.hover" : undefined),
                          p: 0.5,
                        }}
                      >
                        {infoHandles && slotDate ? (
                          <Stack spacing={0.25} sx={{ alignItems: "center" }}>
                            <Typography
                              sx={{
                                fontSize: 9,
                                color: "text.secondary",
                                fontWeight: 700,
                              }}
                            >
                              {timeFormat.format(slotDate)} –{" "}
                              {timeFormat.format(
                                new Date(
                                  slotDate.getTime() +
                                    iSpan.span * 30 * 60 * 1000,
                                ),
                              )}
                            </Typography>
                            <Stack
                              direction="row"
                              spacing={0.5}
                              sx={{
                                flexWrap: "wrap",
                                justifyContent: "center",
                              }}
                            >
                              {(idx !== null
                                ? (slotByIndex.get(idx)?.organizers ?? [])
                                    .slice()
                                    .sort((a, b) =>
                                      organizerHandle(a).localeCompare(
                                        organizerHandle(b),
                                      ),
                                    )
                                : []
                              ).map((id) => {
                                const isMe = id === currentOrganizerId;
                                return (
                                  <Tooltip key={id} title={organizerHandle(id)}>
                                    <Box
                                      sx={{
                                        bgcolor: isMe
                                          ? "warning.main"
                                          : "background.paper",
                                        color: isMe
                                          ? "warning.contrastText"
                                          : "text.primary",
                                        border: "1px solid",
                                        borderColor: "divider",
                                        borderRadius: 0.5,
                                        px: 0.5,
                                        py: 0.25,
                                        fontSize: 10,
                                        fontWeight: 500,
                                        whiteSpace: "normal",
                                        wordBreak: "break-word",
                                      }}
                                    >
                                      {organizerHandle(id)}
                                    </Box>
                                  </Tooltip>
                                );
                              })}
                            </Stack>
                          </Stack>
                        ) : null}
                      </TableCell>,
                    );
                  }

                  return <React.Fragment key={day}>{cells}</React.Fragment>;
                })}
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
    </TableContainer>
  );
};

export default ShiftPlanGrid;
