// Mock for monaco-editor in tests
export const editor = {
  create: () => ({
    dispose: () => {},
    getValue: () => "",
    setValue: () => {},
    onDidChangeModelContent: () => ({ dispose: () => {} }),
    updateOptions: () => {},
    layout: () => {},
  }),
  createModel: () => ({}),
  setModelLanguage: () => {},
  defineTheme: () => {},
  setTheme: () => {},
};

export const languages = {
  register: () => {},
  setMonarchTokensProvider: () => {},
  registerCompletionItemProvider: () => ({ dispose: () => {} }),
};

export class Range {
  startLineNumber: number;
  startColumn: number;
  endLineNumber: number;
  endColumn: number;

  constructor(
    startLineNumber: number,
    startColumn: number,
    endLineNumber: number,
    endColumn: number,
  ) {
    this.startLineNumber = startLineNumber;
    this.startColumn = startColumn;
    this.endLineNumber = endLineNumber;
    this.endColumn = endColumn;
  }
}

export default {
  editor,
  languages,
  Range,
};
