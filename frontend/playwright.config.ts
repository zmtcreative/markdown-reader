import { defineConfig, devices } from '@playwright/test';

/**
 * Optimized Playwright configuration for Wails WebView2 testing
 * See https://playwright.dev/docs/test-configuration.
 */
export default defineConfig({
  testDir: './tests',

  /* Optimized for sequential testing with shared app instances */
  fullyParallel: false, // Disabled for WebView2 to avoid port conflicts
  workers: 1, // Single worker essential for Wails dev server stability

  /* Test execution settings */
  forbidOnly: !!process.env.CI, // Fail build if test.only left in code
  retries: process.env.CI ? 2 : 0, // Retry on CI only

  /* Optimized timeouts for Wails application testing */
  timeout: 90000, // Extended for Wails dev startup (only happens once per suite now!)
  globalTimeout: 600000, // 10 minutes total for entire test suite
  expect: {
    timeout: 15000, // Extended for WebView2 element waiting
  },

  /* Reporter configuration */
  reporter: [
    ['html', { open: 'never' }], // HTML report for detailed analysis
    ['line'], // Concise output during test runs
  ],

  /* Optimized test execution settings */
  use: {
    /* Tracing and debugging */
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',

    /* Action timeouts optimized for Wails */
    actionTimeout: 15000, // Extended for WebView2 interactions
    navigationTimeout: 30000, // Extended for Wails dev server connection

    /* Browser settings */
    headless: false, // Keep visible for debugging Wails applications
  },

  /* WebView2 testing project */
  projects: [
    {
      name: 'webview2',
      testMatch: ['main-test-suite.spec.ts', 'fast-sequential-tests.spec.ts'],
      use: {
        channel: 'msedge', // Use Edge WebView2 runtime for compatibility
        viewport: { width: 1280, height: 720 }, // Consistent viewport for screenshots
      },
    },
  ],

  /* Test output configuration */
  outputDir: 'test-results/',

  /* No web server needed - Wails dev handles this internally */
  // webServer: { ... } // Removed - Wails dev server is managed by test helpers
});
