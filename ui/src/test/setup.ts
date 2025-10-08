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

// Mock monaco-editor
vi.mock("monaco-editor", () => ({
  editor: {
    create: vi.fn(),
    createModel: vi.fn(),
    setModelLanguage: vi.fn(),
  },
  languages: {
    register: vi.fn(),
    setMonarchTokensProvider: vi.fn(),
  },
}));

// Mock @monaco-editor/react
vi.mock("@monaco-editor/react", () => ({
  default: vi.fn(() => null),
}));
