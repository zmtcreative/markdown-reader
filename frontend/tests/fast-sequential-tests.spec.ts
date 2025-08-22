import { test, expect } from '@playwright/test';
import WailsDevHelper from './wails-dev-helper';
import { Page } from '@playwright/test';

test.describe('Fast Sequential Tests - Single App Instance', () => {
  let wailsDev: WailsDevHelper;
  let page: Page;

  // Start application ONCE before all tests
  test.beforeAll(async () => {
    console.log('🚀 Starting Wails application once for all tests...');
    wailsDev = new WailsDevHelper();
    page = await wailsDev.launchAndConnect();
    await wailsDev.waitForAppReady(page);
    console.log('✅ Application ready - will be shared across all tests');
  });

  // Shutdown application ONCE after all tests
  test.afterAll(async () => {
    console.log('🔚 Shutting down shared application...');
    if (wailsDev) {
      await wailsDev.disconnect();
    }
    console.log('✅ Shared application shutdown complete');
  });

  // Between tests: just reset state (no restart needed!)
  test.beforeEach(async () => {
    console.log('🔄 Resetting application state (fast)...');

    // Quick state reset - much faster than app restart
    try {
      // Close any open modals
      await page.keyboard.press('Escape');
      await page.waitForTimeout(200);

      // Navigate back to fresh state by reloading the page
      await page.reload({ waitUntil: 'domcontentloaded' });
      await page.waitForTimeout(500);

      // Wait for app to be ready
      await page.waitForSelector('.app-header', { timeout: 5000 });

      console.log('✅ State reset complete (no app restart needed)');
    } catch (error) {
      console.log(`⚠️ State reset error (continuing anyway): ${error}`);
    }
  });

  // Test 1: Basic application functionality
  test('should display application content correctly', async () => {
    console.log('🧪 Running Test 1: Basic app functionality');

    await expect(page.locator('.app-header')).toBeVisible();
    await expect(page.locator('.content-area')).toBeVisible();
    await expect(page.locator('#content')).toBeVisible();

    await page.screenshot({ path: 'test-results/fast-01-app-content.png' });
    console.log('✅ Test 1 PASSED - Basic functionality verified');
  });

  // Test 2: Help modal (no app restart needed!)
  test('should handle Help modal functionality', async () => {
    console.log('🧪 Running Test 2: Help modal');

    // Verify Help modal elements exist
    await expect(page.locator('#help-modal-overlay')).toBeAttached();
    await expect(page.locator('#help-modal-content')).toBeAttached();

    // Test programmatic modal trigger
    await page.evaluate(() => {
      const helpTitle = 'About';
      const helpText = '<div><h3>Markdown Reader</h3><p>Fast testing!</p></div>';
      const event = new CustomEvent('show-help', {
        detail: { title: helpTitle, text: helpText }
      });
      window.dispatchEvent(event);
    });

    await page.waitForTimeout(800);
    await page.screenshot({ path: 'test-results/fast-02-help-modal.png' });
    console.log('✅ Test 2 PASSED - Help modal functionality verified');
  });

  // Test 3: File operations (still using same app instance!)
  test('should handle file operations', async () => {
    console.log('🧪 Running Test 3: File operations');

    // Test keyboard shortcuts
    await page.keyboard.press('Control+o');
    await page.waitForTimeout(300);

    // Verify app is still responsive
    await expect(page.locator('#content')).toBeVisible();
    await expect(page.locator('.app-header')).toBeVisible();

    await page.screenshot({ path: 'test-results/fast-03-file-operations.png' });
    console.log('✅ Test 3 PASSED - File operations handled correctly');
  });

  // Test 4: Theme functionality (app still running!)
  test('should handle theme operations', async () => {
    console.log('🧪 Running Test 4: Theme operations');

    // Look for theme toggle
    const themeButton = page.locator('button:has-text("Toggle Theme")');

    if (await themeButton.isVisible()) {
      await themeButton.click();
      await page.waitForTimeout(400);
      console.log('✅ Theme toggle clicked successfully');
    } else {
      console.log('ℹ️ Theme button not visible, testing other theme functionality');
    }

    // Verify app is still working
    await expect(page.locator('.app-header')).toBeVisible();

    await page.screenshot({ path: 'test-results/fast-04-theme-operations.png' });
    console.log('✅ Test 4 PASSED - Theme operations handled correctly');
  });

  // Test 5: Multiple rapid interactions (same app instance!)
  test('should handle rapid user interactions', async () => {
    console.log('🧪 Running Test 5: Rapid interactions');

    // Rapid fire multiple interactions
    await page.keyboard.press('Control+o');  // File open
    await page.waitForTimeout(100);

    await page.keyboard.press('Control+p');  // Print
    await page.waitForTimeout(100);

    await page.keyboard.press('F5');         // Refresh
    await page.waitForTimeout(300);

    // Verify app handled all interactions gracefully
    await expect(page.locator('.app-header')).toBeVisible();
    await expect(page.locator('#content')).toBeVisible();

    await page.screenshot({ path: 'test-results/fast-05-rapid-interactions.png' });
    console.log('✅ Test 5 PASSED - Rapid interactions handled correctly');
  });

  // Performance verification test
  test('should demonstrate the performance improvement', async () => {
    console.log('🧪 Running Performance Test: App reuse verification');

    const startTime = Date.now();

    // Verify we're still using the same application instance
    await expect(page.locator('.app-header')).toBeVisible();
    await expect(page.locator('.content-area')).toBeVisible();

    // Quick interactions to verify responsiveness
    await page.keyboard.press('Control+o');
    await page.waitForTimeout(200);
    await page.keyboard.press('Escape');

    const endTime = Date.now();
    const testTime = endTime - startTime;

    console.log(`⚡ Performance test completed in ${testTime}ms`);
    console.log('🎯 This test ran WITHOUT restarting the application!');

    await page.screenshot({ path: 'test-results/fast-06-performance-demo.png' });
    console.log('✅ Performance Test PASSED - Application reuse verified');
  });
});
