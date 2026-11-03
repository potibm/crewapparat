import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { RecordContextProvider } from "react-admin";
import { MarkdownField } from "./MarkdownField";

describe("MarkdownField", () => {
  it("renders the markdown source as HTML", () => {
    render(
      <RecordContextProvider
        value={{ id: "1", text: "# Code of Conduct\n\nBe nice to each other." }}
      >
        <MarkdownField source="text" />
      </RecordContextProvider>,
    );

    expect(
      screen.getByRole("heading", { level: 1, name: "Code of Conduct" }),
    ).toBeInTheDocument();
    expect(screen.getByText("Be nice to each other.")).toBeInTheDocument();
  });

  it("renders nothing when the source is empty", () => {
    const { container } = render(
      <RecordContextProvider value={{ id: "1", text: "" }}>
        <MarkdownField source="text" />
      </RecordContextProvider>,
    );

    expect(container).toBeEmptyDOMElement();
  });
});
