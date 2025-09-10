const { defineConfig, devices } = require('@playwright/test');

module.exports = defineConfig({
  testDir: './',
  fullyParallel: false, // Run tests sequentially to avoid conflicts
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1, // Use single worker to avoid login conflicts
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:22222',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    headless: process.env.CI ? true : false, // Show browser in development
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],

  webServer: {
    command: 'echo "Blog server should already be running on port 22222"',
    port: 22222,
    reuseExistingServer: true,
    timeout: 5000,
  },
});