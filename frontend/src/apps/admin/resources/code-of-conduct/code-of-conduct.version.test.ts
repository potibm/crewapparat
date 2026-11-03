import { describe, it, expect } from "vitest";
import { nextPatchVersion, validateVersion } from "./code-of-conduct.version";

describe("nextPatchVersion", () => {
  it("bumps the patch component", () => {
    expect(nextPatchVersion("1.2.3")).toBe("1.2.4");
  });

  it("falls back to 1.0.0 without a valid base version", () => {
    expect(nextPatchVersion(undefined)).toBe("1.0.0");
    expect(nextPatchVersion("")).toBe("1.0.0");
    expect(nextPatchVersion("draft")).toBe("1.0.0");
  });
});

describe("validateVersion", () => {
  it("accepts semantic versions", () => {
    expect(validateVersion("1.0.0")).toBeUndefined();
    expect(validateVersion("10.20.30")).toBeUndefined();
  });

  it("rejects missing or malformed versions", () => {
    expect(validateVersion("")).toBeTruthy();
    expect(validateVersion("1.0")).toBeTruthy();
    expect(validateVersion("v1.0.0")).toBeTruthy();
    expect(validateVersion("1.0.0-beta")).toBeTruthy();
  });
});
