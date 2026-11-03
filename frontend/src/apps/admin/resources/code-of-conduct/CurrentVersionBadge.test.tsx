import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { RecordContextProvider } from "react-admin";
import { CurrentVersionBadge } from "./CurrentVersionBadge";

vi.mock("./useLatestPublishedCodeOfConduct", () => ({
  useLatestPublishedCodeOfConduct: () => ({ data: [{ id: "current" }] }),
}));

describe("CurrentVersionBadge", () => {
  it("marks the latest published version", () => {
    render(
      <RecordContextProvider value={{ id: "current", status: "published" }}>
        <CurrentVersionBadge />
      </RecordContextProvider>,
    );

    expect(screen.getByText("Current version")).toBeInTheDocument();
  });

  it("renders a placeholder for other versions", () => {
    render(
      <RecordContextProvider value={{ id: "older", status: "published" }}>
        <CurrentVersionBadge />
      </RecordContextProvider>,
    );

    expect(screen.queryByText("Current version")).not.toBeInTheDocument();
    expect(screen.getByText("—")).toBeInTheDocument();
  });
});
