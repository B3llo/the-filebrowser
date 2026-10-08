import { test, expect, type Page } from "@playwright/test";
import { BACKEND, apiLogin, bootstrapSession } from "./helpers";
import { ADMIN } from "./global-setup";

// Regression: on the first and last file only one arrow is rendered. The
// remaining arrow matched both `:first-of-type` and `:last-of-type`, so it got
// `left` and `right` at once and stretched across the viewport. Every arrow
// must keep the size it has on a middle file.
const FOLDER = "e2e-preview-nav";
const NAMES = ["a.txt", "b.txt", "c.txt"];
const MAX_ARROW_PX = 160;

type Label = "Previous" | "Next";
type Size = { width: number; height: number };

test.beforeAll(async ({ request }) => {
  // A user update revokes tokens issued before it, so sort first and log in after.
  const first = await apiLogin(request, ADMIN.username, ADMIN.password);
  const users: Array<{ id: number; username: string }> = await (
    await request.get(`${BACKEND}/api/users`, { headers: { "X-Auth": first } })
  ).json();
  const admin = users.find((u) => u.username === ADMIN.username)!;
  expect(
    (
      await request.put(`${BACKEND}/api/users/${admin.id}`, {
        headers: { "X-Auth": first },
        data: {
          what: "user",
          which: ["sorting"],
          data: { id: admin.id, sorting: { by: "name", asc: true } },
        },
      })
    ).ok()
  ).toBeTruthy();

  const headers = {
    "X-Auth": await apiLogin(request, ADMIN.username, ADMIN.password),
  };
  await request.delete(`${BACKEND}/api/resources/${FOLDER}/`, {
    headers,
    failOnStatusCode: false,
  });
  expect(
    (
      await request.post(`${BACKEND}/api/resources/${FOLDER}/`, { headers })
    ).ok()
  ).toBeTruthy();
  for (const name of NAMES) {
    expect(
      (
        await request.post(`${BACKEND}/api/resources/${FOLDER}/${name}`, {
          headers,
          data: `${name}\n`,
        })
      ).ok()
    ).toBeTruthy();
  }
});

async function arrowSize(page: Page, label: Label): Promise<Size> {
  // Arrows fade out after 1.5s without pointer activity, so wake them up first.
  await page.mouse.move(640, 360);
  const arrow = page.getByRole("button", { name: label, exact: true });
  await expect(arrow).toBeVisible();
  const box = await arrow.boundingBox();
  expect(box, `${label} arrow has no layout box`).not.toBeNull();
  expect(box!.width, `${label} arrow width`).toBeLessThan(MAX_ARROW_PX);
  expect(box!.height, `${label} arrow height`).toBeLessThan(MAX_ARROW_PX);
  return { width: box!.width, height: box!.height };
}

async function clickArrow(page: Page, label: Label) {
  await page.mouse.move(640, 360);
  await page.getByRole("button", { name: label, exact: true }).click();
}

function expectSameSize(actual: Size, expected: Size) {
  expect(Math.abs(actual.width - expected.width)).toBeLessThanOrEqual(1);
  expect(Math.abs(actual.height - expected.height)).toBeLessThanOrEqual(1);
}

test("the lone arrow on the first and last file keeps the size of the other arrows", async ({
  page,
  request,
}) => {
  await bootstrapSession(page, request, ADMIN.username, ADMIN.password);

  // Middle file: both arrows exist, so they give the reference sizes.
  await page.goto(`/files/${FOLDER}/b.txt`);
  await expect(page.locator(".fb-preview-name")).toHaveText("b.txt");
  const reference = {
    Previous: await arrowSize(page, "Previous"),
    Next: await arrowSize(page, "Next"),
  };

  // First file: only "Next" is rendered.
  await clickArrow(page, "Previous");
  await expect(page.locator(".fb-preview-name")).toHaveText("a.txt");
  await expect(
    page.getByRole("button", { name: "Previous", exact: true })
  ).toHaveCount(0);
  expectSameSize(await arrowSize(page, "Next"), reference.Next);

  // Last file: only "Previous" is rendered.
  await clickArrow(page, "Next");
  await expect(page.locator(".fb-preview-name")).toHaveText("b.txt");
  await clickArrow(page, "Next");
  await expect(page.locator(".fb-preview-name")).toHaveText("c.txt");
  await expect(
    page.getByRole("button", { name: "Next", exact: true })
  ).toHaveCount(0);
  expectSameSize(await arrowSize(page, "Previous"), reference.Previous);
});
