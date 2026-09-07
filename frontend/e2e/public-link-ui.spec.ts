import { test, expect } from "@playwright/test";
import { BACKEND, apiLogin } from "./helpers";
import { ALICE } from "./global-setup";

/**
 * Anonymous share page against the PRODUCTION artifact: the backend serves
 * the embedded dist on :8080, so no vite involved.
 *
 * Prerequisite: fresh `pnpm build` before `test:e2e:prepare` so the embedded
 * dist matches the sources (CI does this; see .github/workflows/ci.yaml).
 */
test.use({ baseURL: "http://127.0.0.1:8080" });

test("anonymous share page renders file info", async ({ page, request }) => {
  const aliceTok = await apiLogin(request, ALICE.username, ALICE.password);
  const link = await (
    await request.post(`${BACKEND}/api/share/docs/report.txt`, {
      headers: { "X-Auth": aliceTok },
      data: {},
    })
  ).json();
  expect(link.hash).toBeTruthy();

  try {
    await page.goto(`/share/${link.hash}`);
    await expect(page.getByText("report.txt").first()).toBeVisible({ timeout: 15000 });
    await expect(page.getByRole("link", { name: /download/i }).first()).toBeVisible();
  } finally {
    await request
      .delete(`${BACKEND}/api/share/${link.hash}`, { headers: { "X-Auth": aliceTok } })
      .catch(() => {});
  }
});
