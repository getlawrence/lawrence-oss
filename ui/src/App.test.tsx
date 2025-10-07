import { render } from "@testing-library/react";
import { describe, it, expect } from "vitest";

import App from "./App";

describe("App", () => {
  it("renders without crashing", () => {
    render(<App />);
    expect(document.body).toBeInTheDocument();
  });

  it("renders app container", () => {
    const { container } = render(<App />);
    expect(container.firstChild).toBeTruthy();
  });
});
