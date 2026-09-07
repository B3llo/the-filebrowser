import { defineConfig, devices } from "@playwright/test";
import { readFileSync, existsSync } from "node:fs";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const here = dirname(fileURLToPath(import.meta.url));
const workDir = join(here, "e2e", ".work");

function adminHash(): string {
  const f = join(workDir, "admin-hash");
  if (!existsSync(f)) {
    throw new Error(
      "e2e admin hash missing: run `pnpm run test:e2e:prepare` first"
    );
  }
  return readFileSync(f, "utf8").trim();
}

export default defineConfig({
  testDir: "./e2e",
  globalSetup: "./e2e/global-setup.ts",
  timeout: 60_000,
  expect: { timeout: 10_000 },
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  use: {
    baseURL: "http://127.0.0.1:5173",
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
  },
  projects: [
    { name: "chromium", testIgnore: /mobile\.spec\.ts/, use: { ...devices["Desktop Chrome"] } },
    { name: "firefox", testIgnore: /mobile\.spec\.ts/, use: { ...devices["Desktop Firefox"] } },
    { name: "webkit", testIgnore: /mobile\.spec\.ts/, use: { ...devices["Desktop Safari"] } },
    { name: "mobile", testMatch: /mobile\.spec\.ts/, use: { ...devices["Pixel 7"] } },
  ],
  webServer: [
    {
      // Prebuilt binary from `test:e2e:prepare` (instant start, no recompile).
      // NOTE: the bcrypt hash contains `$` chars, so it goes through FB_PASSWORD
      // env (no shell expansion) instead of the --password flag.
      command: `e2e/.work/fb-test --database e2e/.work/filebrowser.db --root e2e/.work/root --address 127.0.0.1 --port 8080 --username admin`,
      cwd: here,
      env: {
        ...process.env,
        FB_PASSWORD: adminHash(),
        FB_DISABLE_THUMBNAILS: "true",
      } as Record<string, string>,
      url: "http://127.0.0.1:8080/health",
      timeout: 240_000,
      reuseExistingServer: !process.env.CI,
      stdout: "pipe",
    },
    {
      // --host 127.0.0.1 is required: bare `localhost` may bind ::1 only,
      // which makes the 127.0.0.1 readiness probe hang until timeout.
      command: "pnpm dev --port 5173 --strictPort --host 127.0.0.1",
      cwd: here,
      url: "http://127.0.0.1:5173/login",
      timeout: 180_000,
      reuseExistingServer: !process.env.CI,
    },
  ],
});
