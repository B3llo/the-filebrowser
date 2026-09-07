// Prepares the Playwright work dir: fresh DB/root and a bcrypt hash for the
// backend quick-setup admin. Run: `pnpm run test:e2e:prepare`
import { execFileSync } from "node:child_process";
import { mkdirSync, rmSync, writeFileSync } from "node:fs";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const frontendDir = join(dirname(fileURLToPath(import.meta.url)), "..");
const repoRoot = join(frontendDir, "..");
// NOTE: outside frontend/ — vite watches its root and would hard-reload
// the page on every backend DB write inside e2e/.work.
const workDir = join(repoRoot, ".e2e-work");

export const ADMIN_PASSWORD = "e2e-admin-pass";

rmSync(workDir, { recursive: true, force: true });
mkdirSync(join(workDir, "root"), { recursive: true });

const hash = execFileSync("go", ["run", ".", "hash", ADMIN_PASSWORD], {
  cwd: repoRoot,
  encoding: "utf8",
  stdio: ["ignore", "pipe", "inherit"],
}).trim();

writeFileSync(join(workDir, "admin-hash"), hash + "\n");

// Prebuild the backend once: `go run` would recompile on every webServer
// start, adding minutes to each e2e run. The binary lives in git-ignored .work.
console.log("building backend test binary (one-time cost)...");
execFileSync(
  "go",
  ["build", "-o", join(workDir, "fb-test"), "."],
  { cwd: repoRoot, stdio: "inherit" }
);
console.log("e2e work dir ready at", workDir);
