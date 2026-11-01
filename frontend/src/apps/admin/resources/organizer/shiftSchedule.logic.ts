export type ShiftStatus = "ok" | "preferred" | "unavailable";

export interface ShiftBlock {
  index: number;
  date: Date;
  dayKey: string;
  hour: number;
}

export interface ShiftRow {
  key: string;
  label: string;
  cells: { dayKey: string; blocks: ShiftBlock[] }[];
}

export interface ShiftDay {
  key: string;
  label: string;
}

export interface ShiftScheduleGrid {
  days: ShiftDay[];
  rows: ShiftRow[];
}

const shiftStartPattern =
  /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})(?::(\d{2}))?/;

const BLOCK_MINUTES = 30;
const MILLIS_PER_MINUTE = 60_000;
const HOURS_PER_DAY = 24;

const dayFormat = new Intl.DateTimeFormat("en-GB", { weekday: "short" });

/**
 * Parses the wall-clock part of the configured start time, deliberately
 * ignoring the timezone offset so the grid always shows the event's local time
 * regardless of the viewer's timezone.
 */
export function parseShiftStart(start: string): Date | null {
  const match = shiftStartPattern.exec(start ?? "");
  if (!match) return null;

  const [, year, month, day, hour, minute, second] = match;
  const date = new Date(
    Number(year),
    Number(month) - 1,
    Number(day),
    Number(hour),
    Number(minute),
    Number(second ?? "0"),
  );

  return Number.isNaN(date.getTime()) ? null : date;
}

export function shiftStatus(
  block: number,
  preferred: number[],
  unavailable: number[],
): ShiftStatus {
  if (preferred.includes(block)) return "preferred";
  if (unavailable.includes(block)) return "unavailable";

  return "ok";
}

/** Cycle: available -> preferred -> unavailable -> available. */
export function nextShiftStatus(status: ShiftStatus): ShiftStatus {
  switch (status) {
    case "ok":
      return "preferred";
    case "preferred":
      return "unavailable";
    case "unavailable":
      return "ok";
  }
}

export function applyShiftStatus(
  block: number,
  next: ShiftStatus,
  preferred: number[],
  unavailable: number[],
): { preferred: number[]; unavailable: number[] } {
  const nextPreferred = new Set(preferred);
  const nextUnavailable = new Set(unavailable);

  nextPreferred.delete(block);
  nextUnavailable.delete(block);

  if (next === "preferred") nextPreferred.add(block);
  if (next === "unavailable") nextUnavailable.add(block);

  return {
    preferred: [...nextPreferred].sort((a, b) => a - b),
    unavailable: [...nextUnavailable].sort((a, b) => a - b),
  };
}

export function formatHour(hour: number): string {
  return `${String(hour).padStart(2, "0")}:00`;
}

function dayKeyOf(date: Date): string {
  return `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}`;
}

/**
 * Builds the shift grid: one row per clock hour from 00:00 to 23:00, one
 * column per event day. Each hour cell holds the half-hour blocks (:00 and :30)
 * that belong to that day, so a cell can be empty when the event does not run
 * then.
 */
export function buildSchedule(
  start: string,
  blockCount: number,
): ShiftScheduleGrid {
  const base = parseShiftStart(start);
  if (!base || blockCount <= 0) return { days: [], rows: [] };

  const blocks: ShiftBlock[] = [];

  for (let index = 0; index < blockCount; index++) {
    const date = new Date(
      base.getTime() + index * BLOCK_MINUTES * MILLIS_PER_MINUTE,
    );
    blocks.push({
      index,
      date,
      dayKey: dayKeyOf(date),
      hour: date.getHours(),
    });
  }

  const days: ShiftDay[] = [];
  const seenDays = new Set<string>();

  for (const block of blocks) {
    if (seenDays.has(block.dayKey)) continue;

    seenDays.add(block.dayKey);
    days.push({ key: block.dayKey, label: dayFormat.format(block.date) });
  }

  const hoursPresent = new Set(blocks.map((block) => block.hour));
  const rows: ShiftRow[] = [];

  for (let hour = 0; hour < HOURS_PER_DAY; hour++) {
    if (!hoursPresent.has(hour)) continue;

    rows.push({
      key: String(hour),
      label: formatHour(hour),
      cells: days.map((day) => ({
        dayKey: day.key,
        blocks: blocks.filter(
          (block) => block.hour === hour && block.dayKey === day.key,
        ),
      })),
    });
  }

  return { days, rows };
}
