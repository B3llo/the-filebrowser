import { test, expect } from "@playwright/test";
import { BACKEND, bootstrapSession, apiLogin } from "./helpers";
import { ADMIN, ALICE, BOB } from "./global-setup";

test("editor grant: upload and rename inside the grant", async ({ page, request }) => {
  const adminTok = await apiLogin(request, ADMIN.username, ADMIN.password);

  // Upgrade the fixture grant to editor.
  const all = await (await request.get(`${BACKEND}/api/grants`, { headers: { "X-Auth": adminTok } })).json();
  const grant = all.find((g: any) => g.path === "/docs");
  expect(grant).toBeDefined();
  await request.patch(`${BACKEND}/api/grants/${grant.id}`, {
    headers: { "X-Auth": adminTok },
    data: { role: "editor" },
  });

  try {
    const bobTok = await apiLogin(request, BOB.username, BOB.password);

    // New upload inside the granted folder lands in alice's scope.
    const post = await request.post(`${BACKEND}/api/resources/docs/fresh.txt`, {
      headers: { "X-Auth": bobTok },
      data: "fresh-by-bob",
    });
    expect(post.ok()).toBeTruthy();

    // Overwrite works too.
    const put = await request.put(`${BACKEND}/api/resources/docs/fresh.txt`, {
      headers: { "X-Auth": bobTok },
      data: "fresh-by-bob-v2",
    });
    expect(put.ok()).toBeTruthy();

    // Rename inside the grant works...
    const rename = await request.patch(
      `${BACKEND}/api/resources/docs/fresh.txt?action=rename&destination=${encodeURIComponent("/docs/renamed.txt")}`,
      { headers: { "X-Auth": bobTok } }
    );
    expect(rename.ok()).toBeTruthy();

    // ...but escaping the grant is forbidden and nothing leaks out.
    const escape = await request.patch(
      `${BACKEND}/api/resources/docs/renamed.txt?action=rename&destination=${encodeURIComponent("/escaped.txt")}`,
      { headers: { "X-Auth": bobTok } }
    );
    expect(escape.status()).toBe(403);

    // The UI shows the new file through the grant.
    await bootstrapSession(page, request, BOB.username, BOB.password);
    await page.goto("/files/0/docs");
    await expect(page.getByText("renamed.txt")).toBeVisible();
  } finally {
    // Restore the viewer fixture for the other specs.
    await request.patch(`${BACKEND}/api/grants/${grant.id}`, {
      headers: { "X-Auth": adminTok },
      data: { role: "viewer" },
    });
    const aliceTok = await apiLogin(request, ALICE.username, ALICE.password);
    await request
      .delete(`${BACKEND}/api/resources/docs/renamed.txt`, {
        headers: { "X-Auth": aliceTok },
      })
      .catch(() => {});
  }
});
