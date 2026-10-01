import { describe, it, expect } from "vitest";
import { renderHook } from "@testing-library/react";
import { useAppConfig } from "./useConfig";
import { ConfigContext } from "./ConfigContext";
import type { AppConfig } from "./config.schemas";

const dummyConfig: AppConfig = {
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

describe("useAppConfig", () => {
  it("returns config when inside a ConfigContext.Provider", () => {
    const { result } = renderHook(() => useAppConfig(), {
      wrapper: ({ children }) => (
        <ConfigContext value={dummyConfig}>{children}</ConfigContext>
      ),
    });

    expect(result.current).toEqual(dummyConfig);
  });

  it("throws with updated error message when used outside ConfigContext.Provider", () => {
    expect(() => renderHook(() => useAppConfig())).toThrow(
      "useAppConfig must be used within a ConfigContext.Provider",
    );
  });
});
