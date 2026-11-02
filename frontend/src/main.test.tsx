import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { bootstrapApp } from "./main";
import * as Sentry from "@sentry/react";
import { configureOidc } from "@admin/providers/authProvider";

vi.mock("react-dom/client", async () => {
  const actual =
    await vi.importActual<typeof import("react-dom/client")>(
      "react-dom/client",
    );
  return {
    ...actual,
    createRoot: vi.fn(),
  };
});

vi.mock("@sentry/react", async () => {
  const actual =
    await vi.importActual<typeof import("@sentry/react")>("@sentry/react");
  return {
    ...actual,
    isInitialized: vi.fn(),
    init: vi.fn(),
  };
});

vi.mock("@admin/providers/authProvider", async () => {
  return {
    configureOidc: vi.fn(),
  };
});

vi.mock("./App.tsx", () => {
  return {
    default: () => null,
  };
});

import { createRoot } from "react-dom/client";

const mockRender = vi.fn();

function setupDom() {
  const rootEl = document.createElement("div");
  rootEl.id = "root";
  document.body.appendChild(rootEl);
  vi.mocked(createRoot).mockReturnValue({
    render: mockRender,
    unmount: vi.fn(),
  } as unknown as ReturnType<typeof createRoot>);
}

function cleanupDom() {
  const rootEl = document.getElementById("root");
  if (rootEl) {
    rootEl.remove();
  }
}

const validConfig = {
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
};

describe("bootstrapApp", () => {
  beforeEach(() => {
    setupDom();
    vi.stubGlobal(
      "fetch",
      vi.fn(() =>
        Promise.resolve({
          ok: true,
          json: () => Promise.resolve(validConfig),
        } as Response),
      ),
    );
    vi.mocked(Sentry.isInitialized).mockReturnValue(false);
  });

  afterEach(() => {
    cleanupDom();
    vi.unstubAllGlobals();
  });

  it("renders the application when config loads successfully", async () => {
    await bootstrapApp();

    expect(fetch).toHaveBeenCalledWith(
      expect.stringContaining("/api/config"),
      expect.objectContaining({ signal: expect.any(AbortSignal) }),
    );
    expect(mockRender).toHaveBeenCalled();
  });

  it("initializes Sentry when config contains a DSN and Sentry is not already initialized", async () => {
    await bootstrapApp();

    expect(Sentry.init).toHaveBeenCalledWith(
      expect.objectContaining({
        dsn: validConfig.sentry.dsn,
        environment: validConfig.sentry.environment,
        release: validConfig.sentry.version,
      }),
    );
  });

  it("skips Sentry initialization when already initialized", async () => {
    vi.mocked(Sentry.isInitialized).mockReturnValue(true);

    await bootstrapApp();

    expect(Sentry.init).not.toHaveBeenCalled();
  });

  it("skips Sentry initialization when DSN is missing", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() =>
        Promise.resolve({
          ok: true,
          json: () =>
            Promise.resolve({
              ...validConfig,
              sentry: {
                ...validConfig.sentry,
                dsn: undefined,
              },
            }),
        } as Response),
      ),
    );

    await bootstrapApp();

    expect(Sentry.init).not.toHaveBeenCalled();
  });

  it("configures OIDC when auth type is oidc with authority and client_id", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() =>
        Promise.resolve({
          ok: true,
          json: () =>
            Promise.resolve({
              ...validConfig,
              auth: {
                type: "oidc",
                name: "TestIdP",
                authority: "https://idp.example.com",
                client_id: "my-client",
              },
            }),
        } as Response),
      ),
    );

    await bootstrapApp();

    expect(configureOidc).toHaveBeenCalledWith(
      "https://idp.example.com",
      "my-client",
    );
  });

  it("does not configure OIDC when auth type is not oidc", async () => {
    await bootstrapApp();

    expect(configureOidc).not.toHaveBeenCalled();
  });

  it("does not configure OIDC when authority or client_id are missing", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() =>
        Promise.resolve({
          ok: true,
          json: () =>
            Promise.resolve({
              ...validConfig,
              auth: {
                type: "oidc",
                name: "TestIdP",
                authority: "",
                client_id: "",
              },
            }),
        } as Response),
      ),
    );

    await bootstrapApp();

    expect(configureOidc).not.toHaveBeenCalled();
  });

  it("renders error UI when config fetch fails", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() =>
        Promise.resolve({
          ok: false,
          statusText: "Internal Server Error",
        } as Response),
      ),
    );

    await bootstrapApp();

    expect(mockRender).toHaveBeenCalledTimes(1);
    const rendered = mockRender.mock.calls[0][0];
    expect(rendered.type).toBe("div");
    expect(rendered.props.children).toEqual(
      expect.arrayContaining([expect.objectContaining({ type: "h2" })]),
    );
  });

  it("renders error UI when config validation fails", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() =>
        Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ invalid: true }),
        } as Response),
      ),
    );

    await bootstrapApp();

    expect(mockRender).toHaveBeenCalledTimes(1);
    const rendered = mockRender.mock.calls[0][0];
    expect(rendered.type).toBe("div");
    expect(rendered.props.children).toEqual(
      expect.arrayContaining([expect.objectContaining({ type: "h2" })]),
    );
  });

  it("renders error UI when root element is missing", async () => {
    cleanupDom();

    await bootstrapApp();

    expect(document.body.innerHTML).toContain(
      "Failed to find the root element",
    );
    expect(mockRender).not.toHaveBeenCalled();
  });
});
