import { test, expect } from "@playwright/test";
import { BACKEND, apiLogin, bootstrapSession } from "./helpers";
import { ADMIN } from "./global-setup";

const FOLDER = "e2e-upload-refresh";
const EXISTING = Array.from(
  { length: 24 },
  (_, i) => `a${String(i).padStart(2, "0")}.txt`
);
// Sorted before every existing file, so the reload inserts them above the
// item the viewport was anchored on.
const UPLOADED = ["000-upload.txt", "001-upload.txt", "002-upload.txt"];

test.beforeEach(async ({ page, request }) => {
  const token = await apiLogin(request, ADMIN.username, ADMIN.password);
  const headers = { "X-Auth": token };

  const users = await (
    await request.get(`${BACKEND}/api/users`, { headers })
  ).json();
  const admin = users.find(
    (u: { username: string }) => u.username === ADMIN.username
  );
  expect(
    (
      await request.put(`${BACKEND}/api/users/${admin.id}`, {
        headers,
        data: {
          what: "user",
          which: ["singleClick", "sorting", "viewMode"],
          data: {
            id: admin.id,
            singleClick: false,
            viewMode: "list",
            sorting: { by: "name", asc: true },
          },
        },
      })
    ).ok()
  ).toBeTruthy();

  const fresh = {
    "X-Auth": await apiLogin(request, ADMIN.username, ADMIN.password),
  };
  await request.delete(`${BACKEND}/api/resources/${FOLDER}/`, {
    headers: fresh,
  });
  expect(
    (
      await request.post(`${BACKEND}/api/resources/${FOLDER}/`, {
        headers: fresh,
      })
    ).ok()
  ).toBeTruthy();
  await Promise.all(
    EXISTING.map(async (filename) => {
      expect(
        (
          await request.post(`${BACKEND}/api/resources/${FOLDER}/${filename}`, {
            headers: fresh,
            data: filename,
          })
        ).ok()
      ).toBeTruthy();
    })
  );

  await bootstrapSession(page, request, ADMIN.username, ADMIN.password);
  await page.goto(`/files/0/${FOLDER}/`);
  await expect(page.locator("#listing")).toBeVisible();
  await expect(page.locator("#listing .item")).toHaveCount(EXISTING.length);
});

test("uploading keeps completed items in the toast and reveals the new files", async ({
  page,
}) => {
  const scroller = page.locator(".fb-content-main");
  await expect.poll(() => scroller.evaluate((el) => el.scrollTop)).toBe(0);

  await page.locator("#upload-input").setInputFiles(
    UPLOADED.map((name) => ({
      name,
      mimeType: "text/plain",
      buffer: Buffer.from(`contents of ${name}`),
    }))
  );

  // Completed files must remain listed (with their checkmarks) instead of
  // vanishing as the next upload starts.
  await expect(page.locator(".fb-upload-toast-title")).toHaveText(
    "Upload complete",
    { timeout: 15_000 }
  );
  await expect(page.locator(".fb-upload-toast-file-name")).toHaveText(UPLOADED);
  await expect(page.locator(".fb-upload-toast-file-check")).toHaveCount(
    UPLOADED.length
  );

  // The listing reloads with every file, and the viewport stays at the top so
  // the uploads inserted above the old first item are actually visible.
  await expect
    .poll(() => page.locator("#listing .item").count(), { timeout: 15_000 })
    .toBe(EXISTING.length + UPLOADED.length);
  await expect.poll(() => scroller.evaluate((el) => el.scrollTop)).toBe(0);

  const viewport = await scroller.boundingBox();
  const uploadedItem = await page
    .locator(`#listing .item[data-path$="/${UPLOADED[0]}"]`)
    .boundingBox();
  expect(viewport).not.toBeNull();
  expect(uploadedItem).not.toBeNull();
  expect(uploadedItem!.y).toBeGreaterThanOrEqual(viewport!.y);
  expect(uploadedItem!.y).toBeLessThan(viewport!.y + viewport!.height);

  for (const name of UPLOADED) {
    await expect(
      page.locator(`#listing .item[data-path$="/${name}"]`)
    ).toHaveCount(1);
  }
});
