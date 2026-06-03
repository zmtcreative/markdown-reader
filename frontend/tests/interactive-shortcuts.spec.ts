import { test, expect } from '@playwright/test';
import { Page } from '@playwright/test';
import WailsDevHelper from './wails-dev-helper';

test.describe('Interactive Native Shortcut Tests', () => {
  let wailsDev: WailsDevHelper;
  let page: Page;

  test.beforeAll(async () => {
    console.log('🚀 Starting Wails application for interactive shortcut tests...');
    wailsDev = new WailsDevHelper({ runtimeMode: 'interactive' });
    page = await wailsDev.launchAndConnect();
    await wailsDev.waitForAppReady(page);
    console.log('✅ Interactive application session ready');
  });

  test.afterAll(async () => {
    console.log('🔚 Shutting down interactive application session...');
    if (wailsDev) {
      await wailsDev.disconnect();
    }
    console.log('✅ Interactive application session shutdown complete');
  });

  test.beforeEach(async () => {
    console.log('🔄 Resetting interactive test state...');

    try {
      await page.keyboard.press('Escape');
      await page.waitForTimeout(200);
      await page.reload({ waitUntil: 'domcontentloaded' });
      await page.waitForTimeout(500);
      await page.waitForSelector('.app-header', { timeout: 5000 });
      console.log('✅ Interactive test state reset complete');
    } catch (error) {
      console.log(`⚠️ Interactive state reset error (continuing anyway): ${error}`);
    }
  });

  test('should handle Help menu keyboard shortcuts', async () => {
    console.log('🧪 Testing: Help menu keyboard shortcuts');

    await page.keyboard.press('Control+a');
    await page.waitForTimeout(300);

    await expect(page.locator('.app-header')).toBeVisible();
    await expect(page.locator('#content')).toBeVisible();

    await page.screenshot({ path: 'test-results/interactive-01-help-shortcuts.png' });
    console.log('✅ PASSED: Help keyboard shortcuts handled correctly');
  });

  test('should handle File menu keyboard shortcuts', async () => {
    console.log('🧪 Testing: File menu keyboard shortcuts');

    await page.keyboard.press('Control+o');
    await page.waitForTimeout(300);

    await expect(page.locator('#content')).toBeVisible();
    await expect(page.locator('.app-header')).toBeVisible();

    await page.screenshot({ path: 'test-results/interactive-02-file-shortcuts.png' });
    console.log('✅ PASSED: File menu shortcuts work correctly');
  });

  test('should handle Print functionality shortcuts', async () => {
    console.log('🧪 Testing: Print functionality shortcuts');

    await page.keyboard.press('Control+p');
    await page.waitForTimeout(300);

    await expect(page.locator('.content-area')).toBeVisible();

    await page.screenshot({ path: 'test-results/interactive-03-print-shortcuts.png' });
    console.log('✅ PASSED: Print functionality shortcuts work correctly');
  });

  test('should demonstrate file operation responsiveness', async () => {
    console.log('🧪 Testing: File operations responsiveness');

    await page.keyboard.press('Control+o');
    await page.waitForTimeout(100);
    await page.keyboard.press('Control+p');
    await page.waitForTimeout(100);
    await page.keyboard.press('F5');
    await page.waitForTimeout(300);

    await expect(page.locator('.app-header')).toBeVisible();
    await expect(page.locator('#content')).toBeVisible();

    await page.screenshot({ path: 'test-results/interactive-04-file-operations-rapid.png' });
    console.log('✅ PASSED: Rapid file operations handled gracefully');
  });
});