export interface KeyBindingMap {
  [key: string]: {
    description: string;
    category: string;
  };
}

export const BASE_BINDINGS: KeyBindingMap = {
  f: { description: "Toggle favorite", category: "action" },
  d: { description: "Delete article", category: "action" },
  s: { description: "Send to device", category: "action" },
  n: { description: "Add article", category: "navigation" },
  h: { description: "Go home", category: "navigation" },
  a: { description: "Account page", category: "navigation" },
};

export const LIST_BINDINGS: KeyBindingMap = {
  ...BASE_BINDINGS,
  k: { description: "Previous article", category: "list" },
  j: { description: "Next article", category: "list" },
  Enter: { description: "Open article", category: "list" },
};

export const DETAIL_BINDINGS: KeyBindingMap = {
  ...BASE_BINDINGS,
  Escape: { description: "Back to list", category: "navigation" },
};

export const HELP_KEY = "?";
