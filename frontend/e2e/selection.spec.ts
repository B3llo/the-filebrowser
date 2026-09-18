import {
  test,
  expect,
  type Page,
  type APIRequestContext,
} from "@playwright/test";
import { BACKEND, apiLogin } from "./helpers";
import { ADMIN } from "./global-setup";

const FOLDER = "e2e-selection";
const NAMES = ["a.txt", "b.txt", "c.txt", "d.txt", "e.txt"];

async function prepareFixture(request: APIRequestContext, token: string) {
  const headers = { "X-Auth": token };

  await request.delete(`${BACKEND}/api/resources/${FOLDER}/`, {
    headers,
    failOnStatusCode: false,
  });
  const dir = await request.post(`${BACKEND}/api/resources/${FOLDER}/`, {
    headers,
  });
  expect(dir.ok()).toBeTruthy();

  for (const name of NAMES) {
    const res = await request.post(
      `${BACKEND}/api/resources/${FOLDER}/${name}`,
      { headers, data: "x" }
    );
    expect(res.ok()).toBeTruthy();
  }
}

/**
 * Keep the admin user deterministic: name sorting, double-click to open,
 * mosaic view. Applied once, before the shared token is issued: a user update
 * bumps `LastUpdate` and revokes previously issued tokens.
 */
async function prepareAdmin(request: APIRequestContext, token: string) {
  const headers = { "X-Auth": token };
  const users = await (
    await request.get(`${BACKEND}/api/users`, { headers })
  ).json();
  const admin = users.find((u: { username: string }) => u.username === "admin");
  const res = await request.put(`${BACKEND}/api/users/${admin.id}`, {
    headers,
    data: {
      what: "user",
      which: ["singleClick", "sorting", "viewMode"],
      data: {
        id: admin.id,
        singleClick: false,
        viewMode: "mosaic",
        sorting: { by: "name", asc: true },
      },
    },
  });
  expect(res.ok()).toBeTruthy();
}

async function openFixture(page: Page) {
  await page.goto(`/files/${FOLDER}`);
  await expect(page.locator("#listing .item[data-index]")).toHaveCount(
    NAMES.length
  );
}

function itemByName(page: Page, name: string) {
  return page.locator("#listing .item[data-index]", { hasText: name }).first();
}

async function selectedNames(page: Page): Promise<string[]> {
  const raw = await page
    .locator('#listing .item[data-selected="true"] .name')
    .allTextContents();
  return raw.map((n) => n.trim()).sort();
}

interface Card {
  box: { x: number; y: number; width: number; height: number };
  name: string;
}

/** Switch to mosaic and return every card fully visible right of the sidebar. */
async function visibleMosaicCards(page: Page): Promise<Card[]> {
  await page.setViewportSize({ width: 800, height: 900 });
  await page.locator(".fb-view-toggle button").nth(1).click();
  await expect(page.locator("#listing.mosaic")).toBeVisible();

  const viewport = page.viewportSize()!;
  const nav = await page.locator("nav").boundingBox();
  const items = page.locator("#listing .item[data-index]");
  const count = await items.count();
  const cards: Card[] = [];

  for (let i = 0; i < count; i++) {
    const box = await items.nth(i).boundingBox();
    if (!box) continue;
    const visibleLeft = box.x >= (nav?.x ?? 0) + (nav?.width ?? 0) + 4;
    const visibleRight = box.x + box.width <= viewport.width - 4;
    if (!visibleLeft || !visibleRight) continue;
    const name = (await items.nth(i).locator(".name").textContent())!.trim();
    cards.push({ box, name });
  }
  return cards;
}

/**
 * Drag a vertical marquee confined to the card's own column: starts in the
 * grid gap below the card and covers only that card.
 */
async function dragWithinCard(page: Page, card: Card) {
  const x = card.box.x + card.box.width / 2;
  const startY = card.box.y + card.box.height + 6;
  const endY = card.box.y + 6;

  const startsOnItem = await page.evaluate(
    ({ x, y }) => document.elementFromPoint(x, y)?.closest(".item") !== null,
    { x, y: startY }
  );
  expect(startsOnItem).toBe(false);

  await page.mouse.move(x, startY);
  await page.mouse.down();
  await page.mouse.move(x + 2, startY, { steps: 2 });
  await page.mouse.move(x + 2, endY, { steps: 8 });
  await page.mouse.up();
}

// One login per phase: updating a user bumps `LastUpdate`, which revokes every
// token issued before it (see the IssuedAt check in http/auth.go). So the
// deterministic admin settings are applied first, then a fresh token is issued
// and reused; the per-test fixture only touches files, never the user record.
let adminToken = "";

test.beforeAll(async ({ request }) => {
  const bootstrapToken = await apiLogin(
    request,
    ADMIN.username,
    ADMIN.password
  );
  await prepareAdmin(request, bootstrapToken);
  adminToken = await apiLogin(request, ADMIN.username, ADMIN.password);
});

test.beforeEach(async ({ page, request }) => {
  await page
    .context()
    .addCookies([
      { name: "auth", value: adminToken, domain: "127.0.0.1", path: "/" },
    ]);
  // Seeds localStorage before the app boots, so tests need a single page load.
  await page.addInitScript((t) => localStorage.setItem("jwt", t), adminToken);
  await prepareFixture(request, adminToken);
});

test("shift-click selects the visual range and re-clicking shrinks it", async ({
  page,
}) => {
  await openFixture(page);

  await itemByName(page, "a.txt").click();
  await itemByName(page, "d.txt").click({ modifiers: ["Shift"] });
  expect(await selectedNames(page)).toEqual([
    "a.txt",
    "b.txt",
    "c.txt",
    "d.txt",
  ]);

  // Shift-clicking inside the range must shrink it, not collapse or freeze it.
  await itemByName(page, "b.txt").click({ modifiers: ["Shift"] });
  expect(await selectedNames(page)).toEqual(["a.txt", "b.txt"]);

  // Shift-clicking past the end grows it again from the same anchor.
  await itemByName(page, "e.txt").click({ modifiers: ["Shift"] });
  expect(await selectedNames(page)).toEqual([
    "a.txt",
    "b.txt",
    "c.txt",
    "d.txt",
    "e.txt",
  ]);
});

test("ctrl+shift-click adds a range to the current selection", async ({
  page,
}) => {
  await openFixture(page);

  await itemByName(page, "a.txt").click();
  await itemByName(page, "b.txt").click({ modifiers: ["Shift"] });
  await itemByName(page, "d.txt").click({
    modifiers: ["ControlOrMeta", "Shift"],
  });
  expect(await selectedNames(page)).toEqual([
    "a.txt",
    "b.txt",
    "c.txt",
    "d.txt",
  ]);
});

test("marquee drag selects the covered item and highlights it", async ({
  page,
}) => {
  await openFixture(page);
  const cards = await visibleMosaicCards(page);
  expect(cards.length).toBeGreaterThan(0);

  await dragWithinCard(page, cards[0]);

  expect(await selectedNames(page)).toEqual([cards[0].name]);
  const selected = page.locator("#listing .item[data-selected='true']");
  await expect(selected).toHaveCount(1);

  const selectedBackground = await selected.evaluate(
    (el) => getComputedStyle(el).backgroundColor
  );
  expect(selectedBackground).not.toBe("rgba(0, 0, 0, 0)");

  // The trailing click must not clear what the marquee just selected.
  await expect(selected).toHaveCount(1);
});

test("ctrl+marquee adds to the current selection", async ({ page }) => {
  await openFixture(page);
  const cards = await visibleMosaicCards(page);
  expect(cards.length).toBeGreaterThan(1);

  await itemByName(page, cards[0].name).click();
  await page.keyboard.down("Control");
  await dragWithinCard(page, cards[1]);
  await page.keyboard.up("Control");

  expect(await selectedNames(page)).toEqual(
    [cards[0].name, cards[1].name].sort()
  );
});
