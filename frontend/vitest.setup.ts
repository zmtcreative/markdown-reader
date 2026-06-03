import { afterEach, beforeEach, vi } from 'vitest';

beforeEach(() => {
  document.body.className = '';
  document.documentElement.className = '';
  document.documentElement.removeAttribute('style');
  document.body.removeAttribute('style');
  document.title = 'Vitest';

  vi.stubGlobal('alert', vi.fn());
  vi.stubGlobal('print', vi.fn());
  vi.stubGlobal('matchMedia', vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })));
  vi.stubGlobal('ResizeObserver', class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  });
});

afterEach(() => {
  vi.unstubAllGlobals();
});