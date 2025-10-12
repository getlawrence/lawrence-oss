import "@testing-library/jest-dom";
import { vi } from "vitest";

// Mock window.matchMedia
Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: (query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: () => {}, // deprecated
    removeListener: () => {}, // deprecated
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => true,
  }),
});

// Mock monaco-editor modules
vi.mock("monaco-editor", () => ({
  default: {},
  editor: {},
}));

vi.mock("@monaco-editor/loader", () => ({
  default: {
    init: vi.fn(() => Promise.resolve()),
    config: vi.fn(),
  },
}));

vi.mock("@monaco-editor/react", () => ({
  default: vi.fn(() => null),
  Editor: vi.fn(() => null),
}));
