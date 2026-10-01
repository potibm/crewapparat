import { describe, expect, it } from "vitest";
import { parseQueryFromLocation } from "react-admin";

describe("react-admin list query string parsing", () => {
  it("decodes pagination, sort and filter params from the URL", () => {
    expect(
      parseQueryFromLocation({
        search:
          "?page=2&perPage=25&sort=%5B%22id%22%2C%22ASC%22%5D&order=ASC&filter=%7B%22q%22%3A%22a%20b%22%7D",
      }),
    ).toEqual({
      page: "2",
      perPage: "25",
      sort: '["id","ASC"]',
      order: "ASC",
      filter: { q: "a b" },
    });
  });

  it("decodes displayedFilters as JSON", () => {
    expect(
      parseQueryFromLocation({
        search: "?displayedFilters=%7B%22foo%22%3Atrue%7D",
      }),
    ).toEqual({ displayedFilters: { foo: true } });
  });
});
