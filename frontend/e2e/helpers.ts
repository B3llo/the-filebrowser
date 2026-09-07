import { expect, type Page, type APIRequestContext } from "@playwright/test";

export const BACKEND = "http://127.0.0.1:8080";

export async function apiLogin(
  request: APIRequestContext,
  username: string,
  password: string
): Promise<string> {
  const res = await request.post(`${BACKEND}/api/login`, {
    data: { username, password },
  });
  expect(res.ok()).toBeTruthy();
  return (await res.text()).trim();
}

/** Bootstrap an authenticated browser session via API (fast, stable). */
export async function bootstrapSession(
  page: Page,
  request: APIRequestContext,
  username: string,
  password: string
) {
  const token = await apiLogin(request, username, password);
  await page.goto("/login");
  await page.evaluate((t) => localStorage.setItem("jwt", t), token);
  await page.context().addCookies([
    { name: "auth", value: token, domain: "127.0.0.1", path: "/" },
  ]);
  await page.goto("/files");
  await expect(page).not.toHaveURL(/\/login/);
  return token;
}

/** Full UI login through the form (covers the real login UX). */
export async function uiLogin(page: Page, username: string, password: string) {
  await page.goto("/login");
  await page.locator("#login input[type=text]").fill(username);
  await page.locator("#login input[type=password]").fill(password);
  await page.locator("#login input[type=submit]").click();
  await expect(page).not.toHaveURL(/\/login/, { timeout: 15_000 });
}
