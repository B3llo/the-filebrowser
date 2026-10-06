import { test, expect, type Page } from "@playwright/test";
import { readFileSync } from "node:fs";
import { BACKEND, apiLogin, bootstrapSession, uiLogin } from "./helpers";
import { ALICE } from "./global-setup";

const FOLDER = "session-preview";
const VIDEO = readFileSync(
  new URL("./fixtures/session-preview.mp4", import.meta.url)
);

async function waitForVideo(page: Page, filename: string) {
  await expect(page.locator(".fb-preview-name")).toHaveText(filename);
  await expect
    .poll(() =>
      page.locator("video").evaluate((video: HTMLVideoElement) => ({
        ready: video.readyState >= 2,
        error: video.error?.code ?? null,
        src: new URL(video.currentSrc || "http://invalid").pathname,
      }))
    )
    .toEqual({
      ready: true,
      error: null,
      src: `/api/raw/${FOLDER}/${filename}`,
    });
}

test.beforeAll(async ({ request }) => {
  const token = await apiLogin(request, ALICE.username, ALICE.password);
  const headers = { "X-Auth": token };
  const profile = JSON.parse(
    Buffer.from(token.split(".")[1], "base64url").toString()
  ).user;
  expect(
    (
      await request.put(`${BACKEND}/api/users/${profile.id}`, {
        headers,
        data: {
          what: "user",
          which: ["sorting"],
          data: { id: profile.id, sorting: { by: "name", asc: true } },
        },
      })
    ).ok()
  ).toBeTruthy();
  await request.delete(`${BACKEND}/api/resources/${FOLDER}/`, { headers });
  expect(
    (
      await request.post(`${BACKEND}/api/resources/${FOLDER}/`, { headers })
    ).ok()
  ).toBeTruthy();
  for (const filename of ["a.mp4", "b.mp4"]) {
    expect(
      (
        await request.post(`${BACKEND}/api/resources/${FOLDER}/${filename}`, {
          headers,
          data: VIDEO,
        })
      ).ok()
    ).toBeTruthy();
  }
  for (const [filename, content] of [
    ["c.md", "# Third preview"],
    ["d.txt", "Fourth preview"],
  ]) {
    expect(
      (
        await request.post(`${BACKEND}/api/resources/${FOLDER}/${filename}`, {
          headers,
          data: content,
        })
      ).ok()
    ).toBeTruthy();
  }
});

test("video playback, next/previous arrows, text previews and closing preserve the session", async ({
  page,
  request,
}) => {
  const errors: string[] = [];
  page.on("pageerror", (error) => errors.push(error.message));
  await bootstrapSession(page, request, ALICE.username, ALICE.password);
  await page.goto(`/files/0/${FOLDER}/a.mp4`);
  await waitForVideo(page, "a.mp4");
  await page
    .locator("video")
    .evaluate((video: HTMLVideoElement) => video.play());
  await expect
    .poll(() =>
      page
        .locator("video")
        .evaluate((video: HTMLVideoElement) => video.currentTime)
    )
    .toBeGreaterThan(0);
  await page.getByRole("button", { name: "Next", exact: true }).click();
  await waitForVideo(page, "b.mp4");
  await page.getByRole("button", { name: "Previous", exact: true }).click();
  await waitForVideo(page, "a.mp4");
  await page.getByRole("button", { name: "Next", exact: true }).click();
  await waitForVideo(page, "b.mp4");
  await page.getByRole("button", { name: "Next", exact: true }).click();
  await expect(page.locator(".md_preview")).toContainText("Third preview");
  await page.getByRole("button", { name: "Next", exact: true }).click();
  await expect(page.locator("#previewer")).toContainText("Fourth preview");
  await page.getByRole("button", { name: "Previous", exact: true }).click();
  await expect(page.locator(".md_preview")).toContainText("Third preview");
  await page.getByRole("button", { name: "Close", exact: true }).click();
  await expect(page.locator("#listing")).toBeVisible();
  await expect(page).not.toHaveURL(/\/login/);
  expect(errors).toEqual([]);
});

test("a failed refresh in one tab does not revoke another tab's session", async ({
  page,
  context,
  request,
}) => {
  await bootstrapSession(page, request, ALICE.username, ALICE.password);
  await page.goto(`/files/0/${FOLDER}/c.md`);
  await expect(page.locator(".md_preview")).toContainText("Third preview");
  const other = await context.newPage();
  await other.goto(`/files/0/${FOLDER}/a.mp4`);
  await waitForVideo(other, "a.mp4");
  const logoutRequests: string[] = [];
  page.on("request", (req) => {
    if (req.url().endsWith("/api/logout")) logoutRequests.push(req.url());
  });
  await page.route("**/api/renew", (route) =>
    route.fulfill({ status: 401, body: "Unauthorized" })
  );
  await page.route(`**/api/resources/${FOLDER}/d.txt`, (route) =>
    route.fulfill({ status: 401, body: "Unauthorized" })
  );
  await page.getByRole("button", { name: "Next", exact: true }).click();
  await expect(page).toHaveURL(/\/login/);
  expect(logoutRequests).toHaveLength(0);
  await other.reload();
  await waitForVideo(other, "a.mp4");
  const download = await other.request.get(`/api/raw/${FOLDER}/a.mp4`);
  expect(download.status()).toBe(200);
  expect(await download.body()).toEqual(VIDEO);
});

test("real cookie restores video after closing and reopening a tab", async ({
  page,
  context,
}) => {
  await uiLogin(page, ALICE.username, ALICE.password);
  const authCookie = (await context.cookies()).find(
    (cookie) => cookie.name === "auth"
  );
  expect(authCookie?.httpOnly).toBe(true);
  expect(authCookie!.expires - Date.now() / 1000).toBeGreaterThan(
    6 * 24 * 60 * 60
  );
  await page.close();
  const reopened = await context.newPage();
  await reopened.goto(`/files/0/${FOLDER}/a.mp4`);
  await waitForVideo(reopened, "a.mp4");
  await reopened.reload();
  await waitForVideo(reopened, "a.mp4");
  expect(await reopened.evaluate(() => localStorage.getItem("jwt"))).toBeNull();
});

test("token expiry renews without logging out or breaking media", async ({
  page,
  request,
}) => {
  await page.clock.install();
  const calls: string[] = [];
  page.on("request", (req) => {
    if (req.url().includes("/api/renew") || req.url().includes("/api/logout"))
      calls.push(req.url());
  });
  const token = await bootstrapSession(
    page,
    request,
    ALICE.username,
    ALICE.password
  );
  const claims = JSON.parse(
    Buffer.from(token.split(".")[1], "base64url").toString()
  );
  await page.goto(`/files/0/${FOLDER}/a.mp4`);
  await waitForVideo(page, "a.mp4");
  const renewed = page.waitForResponse(
    (res) => res.url().endsWith("/api/renew") && res.status() === 200
  );
  await page.clock.fastForward(
    Math.max(0, claims.exp * 1000 - Date.now() + 1000)
  );
  await renewed;
  await expect(page).not.toHaveURL(/\/login/);
  expect(calls.filter((url) => url.endsWith("/api/logout"))).toHaveLength(0);
  await page.getByRole("button", { name: "Next", exact: true }).click();
  await waitForVideo(page, "b.mp4");
});
