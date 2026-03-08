export interface KeyBindingMap {
  [key: string]: {
    description: string;
    category: string;
  };
}

export const BASE_BINDINGS: KeyBindingMap = {
  f: { description: 'Toggle favorite', category: 'action' },
  d: { description: 'Delete article', category: 'action' },
  s: { description: 'Send to device', category: 'action' },
  n: { description: 'New article', category: 'navigation' },
  h: { description: 'Go home', category: 'navigation' },
  a: { description: 'Account page', category: 'navigation' },
};

export const LIST_BINDINGS: KeyBindingMap = {
  ...BASE_BINDINGS,
  ArrowUp: { description: 'Previous article', category: 'list' },
  k: { description: 'Previous article', category: 'list' },
  ArrowDown: { description: 'Next article', category: 'list' },
  j: { description: 'Next article', category: 'list' },
  ArrowRight: { description: 'Open article', category: 'list' },
  Enter: { description: 'Open article', category: 'list' },
};

export const DETAIL_BINDINGS: KeyBindingMap = {
  ...BASE_BINDINGS,
  ArrowLeft: { description: 'Back to list', category: 'navigation' },
  Escape: { description: 'Back to list', category: 'navigation' },
};

export const HELP_KEY = '?';
