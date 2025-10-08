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

export const Range = class {
  constructor(
    public startLineNumber: number,
    public startColumn: number,
    public endLineNumber: number,
    public endColumn: number
  ) {}
};

export default {
  editor,
  languages,
  Range,
};

