import { defineConfig, devices } from "@playwright/test";
import path from "node:path";

const browserDatabase = path.join("/tmp", `finny-browser-${process.pid}.db`);

export default defineConfig({
  testDir: "./browser-tests",
  fullyParallel: false,
  timeout: 30_000,
  use: {
    baseURL: "http://127.0.0.1:4173",
    trace: "on-first-retry",
    ...devices["Desktop Chrome"],
  },
  webServer: [
    {
      command: "go run ./cmd/finny",
      cwd: path.join(process.cwd(), "../server"),
      env: {
        FINNY_PORT: "18080",
        FINNY_DB_PATH: browserDatabase,
      },
      url: "http://127.0.0.1:18080/health",
      reuseExistingServer: false,
    },
    {
      command: "npm run dev -- --host 127.0.0.1 --port 4173",
      cwd: process.cwd(),
      env: { VITE_API_PROXY_TARGET: "http://127.0.0.1:18080" },
      url: "http://127.0.0.1:4173/",
      reuseExistingServer: false,
    },
  ],
});
