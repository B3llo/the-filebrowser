import { test, expect } from "@playwright/test";
import { BACKEND, bootstrapSession, apiLogin } from "./helpers";
import { ADMIN, ALICE, BOB } from "./global-setup";

/**
 * Uses its own file+grant (privnote, outside /docs) so revoking never
 * disturbs the shared /docs fixture used by the other specs: bob keeps
 * /docs access via the parent grant, but loses privnote exclusively.
 */
test("admin revokes a grant through the access tab", async ({ page, request }) => {
  const adminTok = await apiLogin(request, ADMIN.username, ADMIN.password);
  const aliceTok = await apiLogin(request, ALICE.username, ALICE.password);

  await request.post(`${BACKEND}/api/resources/privnote.txt`, {
    headers: { "X-Auth": aliceTok },
    data: "private note",
  });
  const created = await (
    await request.post(`${BACKEND}/api/grants`, {
      headers: { "X-Auth": adminTok },
      data: { path: "/privnote.txt", owner: "alice", grantee: "bob", role: "viewer" },
    })
  ).json();

  try {
    // Sanity: bob reads it through the grant before revoking.
    const bobTok = await apiLogin(request, BOB.username, BOB.password);
    const before = await request.get(`${BACKEND}/api/resources/privnote.txt`, {
      headers: { "X-Auth": bobTok },
    });
    expect(before.ok()).toBeTruthy();

    await bootstrapSession(page, request, ADMIN.username, ADMIN.password);
    await page.goto("/settings/shares");

    // Switch to the Access tab and isolate our row with the filter.
    await page.locator(".fb-share-tabs button", { hasText: "Access" }).click();
    await page.getByPlaceholder("Search").fill("privnote");
    const row = page.locator("table tbody tr", { hasText: "privnote" });
    await expect(row).toBeVisible();

    // Revoke via the confirm prompt.
    await row.locator("button[title='Revoke access']").click();
    await page.locator("#focus-prompt").click();

    await expect(page.locator("table tbody tr", { hasText: "privnote" })).toHaveCount(0);

    // Bob lost access immediately.
    const res = await request.get(`${BACKEND}/api/resources/privnote.txt`, {
      headers: { "X-Auth": bobTok },
    });
    expect(res.status()).toBe(404);
  } finally {
    await request
      .delete(`${BACKEND}/api/grants/${created.id}`, { headers: { "X-Auth": adminTok } })
      .catch(() => {});
    await request
      .delete(`${BACKEND}/api/resources/privnote.txt`, { headers: { "X-Auth": aliceTok } })
      .catch(() => {});
  }
});
