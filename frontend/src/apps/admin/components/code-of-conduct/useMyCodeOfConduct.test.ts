import { describe, it, expect, vi } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";

const mocks = vi.hoisted(() => ({
  fetchMe: vi.fn(),
  getAccessToken: vi.fn(async () => "test-token"),
  invalidateMe: vi.fn(),
}));

vi.mock("../../providers/authProvider", () => mocks);

import { useMyCodeOfConduct } from "./useMyCodeOfConduct";

const meWith = (acceptedAt: string | null) => ({
  id: "organizer-1",
  handle: "tester",
  realname: "Test Organizer",
  email: "tester@example.com",
  is_admin: false,
  code_of_conduct: {
    id: "coc-1",
    version: "1.1.0",
    status: "published",
    text: "# CoC",
  },
  code_of_conduct_accepted_at: acceptedAt,
});

const mockResponse = (status: number, data: unknown) => ({
  ok: status >= 200 && status < 300,
  status,
  json: async () => data,
});

describe("useMyCodeOfConduct", () => {
  it("requires acceptance when the current version is not accepted", async () => {
    mocks.fetchMe.mockResolvedValue(meWith(null));

    const { result } = renderHook(() => useMyCodeOfConduct());

    await waitFor(() => expect(result.current.state).toBe("required"));
    expect(result.current.codeOfConduct?.id).toBe("coc-1");
  });

  it("passes when the current version was accepted", async () => {
    mocks.fetchMe.mockResolvedValue(meWith("2026-05-01T12:00:00Z"));

    const { result } = renderHook(() => useMyCodeOfConduct());

    await waitFor(() => expect(result.current.state).toBe("accepted"));
  });

  it("passes when nothing is published", async () => {
    mocks.fetchMe.mockResolvedValue({
      ...meWith(null),
      code_of_conduct: null,
    });

    const { result } = renderHook(() => useMyCodeOfConduct());

    await waitFor(() => expect(result.current.state).toBe("accepted"));
  });

  it("passes for unlinked users", async () => {
    mocks.fetchMe.mockResolvedValue(null);

    const { result } = renderHook(() => useMyCodeOfConduct());

    await waitFor(() => expect(result.current.state).toBe("accepted"));
  });

  it("accepts and re-checks the status", async () => {
    mocks.fetchMe
      .mockResolvedValueOnce(meWith(null))
      .mockResolvedValueOnce(meWith("2026-05-01T12:00:00Z"));

    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        mockResponse(200, { accepted_at: "2026-05-01T12:00:00Z" }),
      ),
    );

    const { result } = renderHook(() => useMyCodeOfConduct());

    await waitFor(() => expect(result.current.state).toBe("required"));

    await result.current.accept();

    await waitFor(() => expect(result.current.state).toBe("accepted"));
    expect(mocks.invalidateMe).toHaveBeenCalled();
  });

  it("reloads the prompt when the version changed in the meantime", async () => {
    mocks.fetchMe
      .mockResolvedValueOnce(meWith(null))
      .mockResolvedValueOnce(meWith(null));

    vi.stubGlobal(
      "fetch",
      vi.fn(async () => mockResponse(409, {})),
    );

    const { result } = renderHook(() => useMyCodeOfConduct());

    await waitFor(() => expect(result.current.state).toBe("required"));

    await expect(result.current.accept()).rejects.toThrow(
      "The Code of Conduct changed",
    );

    await waitFor(() => expect(result.current.state).toBe("required"));
  });
});
