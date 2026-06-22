import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: 'tests/vrt',
  snapshotDir: 'tests/vrt/__snapshots__',
  use: {
    baseURL: 'http://localhost:6006',
    // Consistent viewport for snapshot reproducibility.
    viewport: { width: 1280, height: 720 },
  },
  // Pixel tolerance for minor anti-aliasing differences.
  expect: {
    toHaveScreenshot: { maxDiffPixelRatio: 0.01 },
  },
  webServer: {
    command: 'npm run storybook',
    url: 'http://localhost:6006',
    reuseExistingServer: true,
    timeout: 60_000,
  },
})
