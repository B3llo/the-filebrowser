import { test, expect, type Page } from "@playwright/test";
import { BACKEND, apiLogin, bootstrapSession } from "./helpers";
import { ADMIN } from "./global-setup";

// A lone arrow used to match both :first-of-type and :last-of-type,
// receiving both left and right offsets and stretching across the preview.
const FOLDER = "e2e-preview-nav";
const EXTENSIONS = ["txt", "png"];
const PNG = Buffer.from(
  "iVBORw0KGgoAAAANSUhEUgAAAAgAAAAICAIAAABLbSncAAAAEUlEQVR4nGPQaDiAFTEMLQkABNVaAT2TTw0AAAAASUVORK5CYII=",
  "base64"
);
type Label = "Previous" | "Next";
type Size = { width: number; height: number };

test.beforeAll(async ({ request }) => {
  // Updating a user revokes earlier tokens: configure sorting before login.
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
  for (const extension of EXTENSIONS) {
    const folder = `${FOLDER}-${extension}`;
    await request.delete(`${BACKEND}/api/resources/${folder}/`, {
      headers,
      failOnStatusCode: false,
    });
    expect(
      (
        await request.post(`${BACKEND}/api/resources/${folder}/`, { headers })
      ).ok()
    ).toBeTruthy();
    for (const letter of ["a", "b", "c"]) {
      const filename = `${letter}.${extension}`;
      expect(
        (
          await request.post(`${BACKEND}/api/resources/${folder}/${filename}`, {
            headers,
            data: extension === "png" ? PNG : `${filename}\n`,
          })
        ).ok()
      ).toBeTruthy();
    }
  }
});

async function wakeArrows(page: Page) {
  const viewport = page.viewportSize()!;
  await page.mouse.move(viewport.width / 2, viewport.height / 2);
}

async function arrowSize(page: Page, label: Label): Promise<Size> {
  await wakeArrows(page);
  const arrow = page.getByRole("button", { name: label, exact: true });
  await expect(arrow).toBeVisible();
  const box = await arrow.boundingBox();
  expect(box).not.toBeNull();
  expect(box!.width).toBeLessThanOrEqual(48);
  expect(box!.height).toBeLessThanOrEqual(48);
  return { width: box!.width, height: box!.height };
}

async function clickArrow(page: Page, label: Label) {
  await wakeArrows(page);
  await page.getByRole("button", { name: label, exact: true }).click();
}

function expectSameSize(actual: Size, expected: Size) {
  expect(Math.abs(actual.width - expected.width)).toBeLessThanOrEqual(1);
  expect(Math.abs(actual.height - expected.height)).toBeLessThanOrEqual(1);
}

for (const extension of EXTENSIONS) {
  for (const viewport of [
    { width: 1280, height: 720 },
    { width: 390, height: 844 },
  ]) {
    test(`the lone ${extension} arrow keeps its size and side at width ${viewport.width}`, async ({
      page,
      request,
    }) => {
      await page.setViewportSize(viewport);
      await bootstrapSession(page, request, ADMIN.username, ADMIN.password);
      await page.goto(`/files/${FOLDER}-${extension}/b.${extension}`);
      await expect(page.locator(".fb-preview-name")).toHaveText(
        `b.${extension}`
      );
      if (extension === "png") {
        const image = page.locator("#previewer img").first();
        await expect(image).toBeVisible();
        await expect
          .poll(() =>
            image.evaluate((el) => (el as HTMLImageElement).naturalWidth)
          )
          .toBeGreaterThan(0);
      }
      const reference = {
        Previous: await arrowSize(page, "Previous"),
        Next: await arrowSize(page, "Next"),
      };

      await clickArrow(page, "Previous");
      await expect(page.locator(".fb-preview-name")).toHaveText(
        `a.${extension}`
      );
      await expect(
        page.getByRole("button", { name: "Previous", exact: true })
      ).toHaveCount(0);
      expectSameSize(await arrowSize(page, "Next"), reference.Next);
      const nextBox = await page
        .getByRole("button", { name: "Next", exact: true })
        .boundingBox();
      expect(nextBox!.x).toBeGreaterThan(viewport.width - 64);

      await clickArrow(page, "Next");
      await expect(page.locator(".fb-preview-name")).toHaveText(
        `b.${extension}`
      );
      await clickArrow(page, "Next");
      await expect(page.locator(".fb-preview-name")).toHaveText(
        `c.${extension}`
      );
      await expect(
        page.getByRole("button", { name: "Next", exact: true })
      ).toHaveCount(0);
      expectSameSize(await arrowSize(page, "Previous"), reference.Previous);
      const prevBox = await page
        .getByRole("button", { name: "Previous", exact: true })
        .boundingBox();
      expect(prevBox!.x).toBeLessThan(24);
    });
  }
}
