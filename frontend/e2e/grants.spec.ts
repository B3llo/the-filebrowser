import { test, expect } from "@playwright/test";
import { BACKEND, bootstrapSession, uiLogin, apiLogin } from "./helpers";
import { ALICE, BOB } from "./global-setup";

test("bob logs in through the form", async ({ page }) => {
  await uiLogin(page, BOB.username, BOB.password);
  await expect(page.locator("button:has-text('Shared with me')").first()).toBeVisible();
});

test("viewer grant: shared list, open folder, read file", async ({ page, request }) => {
  await bootstrapSession(page, request, BOB.username, BOB.password);

  await page.goto("/shared");
  const row = page.locator(".fb-grant-row", { hasText: "docs" });
  await expect(row).toBeVisible();
  await expect(row).toContainText("alice");
  await row.click();

  // Landed in the Files view on the owner's path (backend resolves the grant).
  await expect(page).toHaveURL(/\/files\/.*\/docs/);
  await expect(page.getByText("report.txt")).toBeVisible();
  await expect(page.getByText("notes.md")).toBeVisible();

  // Content actually comes from alice's scope.
  const token = await apiLogin(request, BOB.username, BOB.password);
  const raw = await request.get(`${BACKEND}/api/raw/docs/report.txt`, {
    headers: { "X-Auth": token },
  });
  expect(raw.ok()).toBeTruthy();
  expect(await raw.text()).toContain("alice-report");
});

test("viewer cannot write (role gate)", async ({ request }) => {
  const token = await apiLogin(request, BOB.username, BOB.password);
  const put = await request.put(`${BACKEND}/api/resources/docs/report.txt`, {
    headers: { "X-Auth": token },
    data: "hacked",
  });
  expect(put.status()).toBe(403);

  const del = await request.delete(`${BACKEND}/api/resources/docs/report.txt`, {
    headers: { "X-Auth": token },
  });
  expect(del.status()).toBe(403);
});

test("stranger paths stay hidden", async ({ request }) => {
  const token = await apiLogin(request, BOB.username, BOB.password);
  // Outside the grant subtree: 404, not 403 (no existence leak).
  const res = await request.get(`${BACKEND}/api/resources/nope.txt`, {
    headers: { "X-Auth": token },
  });
  expect(res.status()).toBe(404);
  void ALICE;
});
