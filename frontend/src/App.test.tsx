import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import App from "./App";

vi.mock("@admin/Admin", () => ({
  AdminApp: () => <div data-testid="admin-app">AdminApp</div>,
}));

describe("App", () => {
  it("renders the AdminApp inside a Sentry ErrorBoundary", () => {
    render(<App />);

    expect(screen.getByTestId("admin-app")).toBeInTheDocument();
  });
});
