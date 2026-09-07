import { test, expect } from "@playwright/test";
import { BACKEND, apiLogin } from "./helpers";
import { ADMIN, ALICE } from "./global-setup";

/** Public link sharing keeps working alongside grants. */
test("public link flow is intact", async ({ page, request, context }) => {
  const aliceTok = await apiLogin(request, ALICE.username, ALICE.password);

  const link = await (
    await request.post(`${BACKEND}/api/share/docs/report.txt`, {
      headers: { "X-Auth": aliceTok },
      data: { password: "", expires: "", unit: "hours" },
    })
  ).json();
  expect(link.hash).toBeTruthy();

  try {
    // Anonymous metadata + download without any credentials.
    const meta = await context.request.get(`${BACKEND}/api/public/share/${link.hash}`);
    expect(meta.ok()).toBeTruthy();
    expect((await meta.json()).name).toBe("report.txt");

    const dl = await context.request.get(
      `${BACKEND}/api/public/dl/${link.hash}/report.txt`
    );
    expect(dl.ok()).toBeTruthy();
    expect(await dl.text()).toContain("alice-report");

    // The share page itself renders for anonymous visitors (no forced
    // redirect to /login: anonymous 401s must not trigger logout).
    await page.goto(`/share/${link.hash}`);
    await expect(page).toHaveURL(new RegExp(`/share/${link.hash}`));
    await expect(page.getByText("report.txt").first()).toBeVisible({ timeout: 15000 });
  } finally {
    const adminTok = await apiLogin(request, ADMIN.username, ADMIN.password);
    await request
      .delete(`${BACKEND}/api/share/${link.hash}`, { headers: { "X-Auth": adminTok } })
      .catch(() => {});
  }
});
