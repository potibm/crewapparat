import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { ConfigContext } from "@core/config/ConfigContext";
import { AppBootstrapper } from "./Admin";
import type { AppConfig } from "@core/config/config.schemas";

vi.mock("./providers/dataProvider", () => ({
  dataProvider: {},
}));

vi.mock("./theme/MyTheme", () => ({
  MyTheme: {},
  MyDarkTheme: {},
}));

vi.mock("./theme/MyLayout", () => ({
  MyLayout: () => <div>Layout</div>,
}));

vi.mock("./components/auth/OidcLogin", () => ({
  OidcLogin: () => <div>OidcLogin</div>,
}));

vi.mock("react-admin", async () => {
  const actual =
    await vi.importActual<typeof import("react-admin")>("react-admin");
  return {
    ...actual,
    Admin: vi.fn(({ children }) => (
      <div data-testid="admin-mock">{children}</div>
    )),
    Resource: vi.fn(() => <div data-testid="resource-mock" />),
  };
});

import { Admin } from "react-admin";

const baseConfig: AppConfig = {
  version: "1.0.0",
  environment: "test",
  sentry: {
    dsn: "https://abc@example.com/1",
    environment: "test",
    version: "1.0.0",
    replay_session_sample_rate: 0,
    replay_error_sample_rate: 1,
  },
  date_locale: "en-US",
  date_options: {
    year: "numeric",
    month: "long",
    day: "numeric",
  },
};

function renderWithConfig(config: AppConfig) {
  return render(
    <ConfigContext value={config}>
      <AppBootstrapper />
    </ConfigContext>,
  );
}

describe("AppBootstrapper", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders the admin app", () => {
    renderWithConfig(baseConfig);

    expect(screen.getByTestId("admin-mock")).toBeInTheDocument();
  });

  it("passes loginPage=false and no authProvider when OIDC is disabled", () => {
    renderWithConfig(baseConfig);

    const calls = vi.mocked(Admin).mock.calls;
    expect(calls.length).toBeGreaterThan(0);
    expect(calls[0][0]).toMatchObject({
      loginPage: false,
      authProvider: undefined,
    });
  });

  it("passes OidcLogin and authProvider when OIDC is enabled", () => {
    renderWithConfig({
      ...baseConfig,
      auth: {
        type: "oidc",
        name: "TestIdP",
        authority: "https://idp.example.com",
        client_id: "my-client",
      },
    });

    const calls = vi.mocked(Admin).mock.calls;
    expect(calls.length).toBeGreaterThan(0);
    expect(calls[0][0]).toMatchObject({
      loginPage: expect.any(Function),
      authProvider: expect.any(Object),
    });
  });
});
