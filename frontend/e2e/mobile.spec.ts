import { test, expect } from "@playwright/test";
import { bootstrapSession } from "./helpers";
import { BOB } from "./global-setup";

/** Mobile viewport: shared list stays usable, tap navigates into Files. */
test("shared list works on mobile", async ({ page, request }) => {
  await bootstrapSession(page, request, BOB.username, BOB.password);

  await page.goto("/shared");
  const row = page.locator(".fb-grant-row", { hasText: "docs" });
  await expect(row).toBeVisible();
  await row.tap();

  await expect(page).toHaveURL(/\/files\/.*\/docs/);
  await expect(page.getByText("report.txt")).toBeVisible();
});
