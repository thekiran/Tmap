import { defineConfig, devices } from '@playwright/test';
import { existsSync } from 'node:fs';

const chromeCandidates = [
  process.env.PLAYWRIGHT_CHROME_EXECUTABLE,
  'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe',
  'C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe',
].filter(Boolean) as string[];
const chromeExecutablePath = chromeCandidates.find((candidate) => existsSync(candidate));
const launchOptions = chromeExecutablePath ? { executablePath: chromeExecutablePath } : undefined;
const baseURL = 'http://127.0.0.1:5188';

export default defineConfig({
  testDir: './e2e',
  timeout: 60_000,
  expect: { timeout: 10_000 },
  fullyParallel: false,
  reporter: [['list']],
  use: {
    baseURL,
    viewport: { width: 1440, height: 960 },
    trace: 'retain-on-failure',
    launchOptions,
  },
  webServer: {
    command: 'npm run dev -- --host 127.0.0.1 --port 5188 --strictPort',
    url: baseURL,
    reuseExistingServer: false,
    timeout: 120_000,
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'], launchOptions },
    },
  ],
});
