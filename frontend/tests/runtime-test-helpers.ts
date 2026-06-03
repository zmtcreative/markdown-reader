import { expect, Page } from '@playwright/test';

export async function emitRuntimeEvent(page: Page, eventName: string, ...eventArgs: any[]): Promise<void> {
  await page.evaluate(({ eventName, eventArgs }) => {
    const runtime = (window as any).runtime;
    if (!runtime?.EventsEmit) {
      throw new Error(`window.runtime.EventsEmit is not available for event: ${eventName}`);
    }

    runtime.EventsEmit(eventName, ...eventArgs);
  }, { eventName, eventArgs });
}

export async function expectTheme(page: Page, theme: 'light' | 'dark'): Promise<void> {
  const expectedTitle = theme === 'dark' ? 'Switch to light theme' : 'Switch to dark theme';

  await expect(page.locator('html')).toHaveClass(theme);
  await expect(page.locator('body')).toHaveClass(new RegExp(`\\b${theme}\\b`));
  await expect(page.locator('.theme-toggle-btn')).toHaveAttribute('title', expectedTitle);
}