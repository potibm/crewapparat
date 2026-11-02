import { describe, expect, it } from "vitest";
import {
  applyShiftStatus,
  buildSchedule,
  nextShiftStatus,
  parseShiftStart,
  shiftStatus,
} from "./shiftSchedule.logic";

describe("parseShiftStart", () => {
  it("uses the wall clock and ignores the timezone offset", () => {
    const date = parseShiftStart("2026-08-07T16:30:00+02:00");

    expect(date).not.toBeNull();
    expect(date?.getFullYear()).toBe(2026);
    expect(date?.getMonth()).toBe(7);
    expect(date?.getDate()).toBe(7);
    expect(date?.getHours()).toBe(16);
    expect(date?.getMinutes()).toBe(30);
  });

  it("returns null for invalid input", () => {
    expect(parseShiftStart("")).toBeNull();
    expect(parseShiftStart("not-a-date")).toBeNull();
  });
});

describe("shiftStatus", () => {
  it("defaults to ok", () => {
    expect(shiftStatus(1, [], [])).toBe("ok");
  });

  it("prefers the preferred list over unavailable", () => {
    expect(shiftStatus(1, [1], [1])).toBe("preferred");
  });

  it("reports unavailable", () => {
    expect(shiftStatus(1, [], [1])).toBe("unavailable");
  });
});

describe("nextShiftStatus", () => {
  it("cycles ok -> preferred -> unavailable -> ok", () => {
    expect(nextShiftStatus("ok")).toBe("preferred");
    expect(nextShiftStatus("preferred")).toBe("unavailable");
    expect(nextShiftStatus("unavailable")).toBe("ok");
  });
});

describe("applyShiftStatus", () => {
  it("moves a block between lists and keeps them sorted", () => {
    expect(applyShiftStatus(3, "preferred", [1], [5])).toEqual({
      preferred: [1, 3],
      unavailable: [5],
    });
    expect(applyShiftStatus(5, "preferred", [1, 3], [5])).toEqual({
      preferred: [1, 3, 5],
      unavailable: [],
    });
    expect(applyShiftStatus(3, "ok", [1, 3], [5])).toEqual({
      preferred: [1],
      unavailable: [5],
    });
  });
});

describe("buildSchedule", () => {
  it("groups the half-hour blocks into one row per hour", () => {
    const grid = buildSchedule("2026-08-07T16:00:00+02:00", 6);

    expect(grid.days.map((day) => day.label)).toEqual(["Fri"]);
    expect(grid.rows.map((row) => row.label)).toEqual([
      "16:00",
      "17:00",
      "18:00",
    ]);
    expect(grid.rows[0].cells[0].blocks.map((block) => block.index)).toEqual([
      0, 1,
    ]);
    expect(grid.rows[1].cells[0].blocks.map((block) => block.index)).toEqual([
      2, 3,
    ]);
    expect(grid.rows[2].cells[0].blocks.map((block) => block.index)).toEqual([
      4, 5,
    ]);
  });

  it("orders the hours from 00:00 to 23:00", () => {
    const grid = buildSchedule("2026-08-07T23:00:00+02:00", 4);

    expect(grid.days.map((day) => day.label)).toEqual(["Fri", "Sat"]);
    expect(grid.rows.map((row) => row.label)).toEqual(["00:00", "23:00"]);

    // 00:00 belongs to Saturday only, 23:00 to Friday only.
    expect(grid.rows[0].cells[0].blocks).toEqual([]);
    expect(grid.rows[0].cells[1].blocks.map((block) => block.index)).toEqual([
      2, 3,
    ]);
    expect(grid.rows[1].cells[0].blocks.map((block) => block.index)).toEqual([
      0, 1,
    ]);
    expect(grid.rows[1].cells[1].blocks).toEqual([]);
  });

  it("returns an empty grid for missing or invalid input", () => {
    expect(buildSchedule("", 10)).toEqual({ days: [], rows: [] });
    expect(buildSchedule("2026-08-07T16:00:00+02:00", 0)).toEqual({
      days: [],
      rows: [],
    });
  });
});
