import type { FullConfig } from "@playwright/test";
import { mkdirSync, writeFileSync } from "node:fs";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const here = dirname(fileURLToPath(import.meta.url));
const workDir = join(here, "..", "..", ".e2e-work");
const API = "http://127.0.0.1:8080";

export const ADMIN = { username: "admin", password: "e2e-admin-pass" };
export const ALICE = { username: "alice", password: "alice-pass-123" };
export const BOB = { username: "bob", password: "bob-pass-1234" };

// 1x1 transparent PNG.
const PIXEL_PNG = Buffer.from(
  "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==",
  "base64"
);

async function apiLogin(username: string, password: string): Promise<string> {
  const res = await fetch(`${API}/api/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
  });
  if (!res.ok) throw new Error(`login ${username} failed: ${res.status}`);
  return (await res.text()).trim();
}

async function api(token: string, method: string, path: string, body?: unknown) {
  const res = await fetch(`${API}${path}`, {
    method,
    headers: {
      "Content-Type": "application/json",
      "X-Auth": token,
    },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  if (!res.ok) {
    throw new Error(`${method} ${path} failed: ${res.status} ${await res.text()}`);
  }
  const text = await res.text();
  if (!text) return null;
  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
}

/** Idempotent: reruns reuse the existing DB instead of failing on 409. */
async function apiIdempotent(token: string, method: string, path: string, body?: unknown) {
  const res = await fetch(`${API}${path}`, {
    method,
    headers: {
      "Content-Type": "application/json",
      "X-Auth": token,
    },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  if (res.status === 409) return null;
  if (!res.ok) {
    throw new Error(`${method} ${path} failed: ${res.status} ${await res.text()}`);
  }
  const text = await res.text();
  if (!text) return null;
  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
}

async function globalSetup(_config: FullConfig) {
  // Seed files on disk (server root is .e2e-work/root; user scopes are
  // created by the user-creation calls below, MkdirAll is idempotent).
  const docs = join(workDir, "root", "alice", "docs");
  mkdirSync(docs, { recursive: true });
  writeFileSync(join(docs, "report.txt"), "alice-report\n");
  writeFileSync(join(docs, "notes.md"), "# Notes\n\nshared notes\n");
  writeFileSync(join(docs, "pixel.png"), PIXEL_PNG);

  const adminTok = await apiLogin(ADMIN.username, ADMIN.password);

  const existing: Array<{ username: string }> = await api(
    adminTok,
    "GET",
    "/api/users"
  );
  const haveUser = new Set(existing.map((u) => u.username));

  const mkUser = (username: string, password: string, scope: string, perm: object) => {
    if (haveUser.has(username)) return Promise.resolve(null);
    return api(adminTok, "POST", "/api/users", {
      what: "user",
      which: [],
      current_password: ADMIN.password,
      data: {
        username,
        password,
        scope,
        perm,
        viewMode: "list",
        sorting: { by: "name", asc: true },
      },
    });
  };

  await mkUser(ALICE.username, ALICE.password, "alice", {
    admin: false,
    share: true,
    download: true,
    create: true,
    modify: true,
    rename: true,
    delete: true,
  });
  await mkUser(BOB.username, BOB.password, "bob", {
    admin: false,
    share: false,
    download: true,
    create: true,
    modify: true,
    rename: true,
    delete: false,
  });

  // The shared fixture: alice's /docs visible to bob.
  const current: Array<{ path: string; ownerID: number; granteeID: number }> =
    await api(adminTok, "GET", "/api/grants");
  const usersByName = new Map(existing.map((u: any) => [u.username, u.id]));
  const aliceID = usersByName.get("alice");
  const bobID = usersByName.get("bob");
  const hasGrant = current.some(
    (g) => g.path === "/docs" && g.ownerID === aliceID && g.granteeID === bobID
  );
  if (!hasGrant) {
    await api(adminTok, "POST", "/api/grants", {
      path: "/docs",
      owner: "alice",
      grantee: "bob",
      role: "viewer",
    });
  } else {
    // A previous interrupted run may have left the editor role behind.
    const stale = current.find(
      (g: any) =>
        g.path === "/docs" && g.ownerID === aliceID && g.granteeID === bobID && g.role !== "viewer"
    ) as any;
    if (stale) {
      await api(adminTok, "PATCH", `/api/grants/${stale.id}`, { role: "viewer" });
    }
  }
}

export default globalSetup;
