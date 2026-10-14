import { test, expect, type Page } from "@playwright/test";
import { BACKEND, apiLogin, bootstrapSession } from "./helpers";
import { ADMIN } from "./global-setup";

const FOLDER = "e2e-listing-scroll";

async function visibleRegion(page: Page) {
  return page.locator(".fb-content-main").evaluate((scroller) => {
    const top = scroller.getBoundingClientRect().top;
    const item = Array.from(
      scroller.querySelectorAll<HTMLElement>("#listing .item")
    ).find((el) => el.getBoundingClientRect().bottom > top);
    return {
      scrollTop: scroller.scrollTop,
      path: item?.dataset.path,
      offset: item ? item.getBoundingClientRect().top - top : 0,
    };
  });
}

for (const mode of ["list", "mosaic"]) {
  test.describe(mode, () => {
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
                viewMode: mode,
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
        Array.from({ length: 180 }, async (_, i) => {
          const filename = `${String(i).padStart(3, "0")}.txt`;
          expect(
            (
              await request.post(
                `${BACKEND}/api/resources/${FOLDER}/${filename}`,
                {
                  headers: fresh,
                  data: filename,
                }
              )
            ).ok()
          ).toBeTruthy();
        })
      );
      await bootstrapSession(page, request, ADMIN.username, ADMIN.password);
      await page.goto(`/files/${FOLDER}/`);
      await expect(page.locator("#listing")).toBeVisible();
      // Load multiple incremental batches, not just the initial 50 files.
      for (let i = 0; i < 4; i++) {
        const count = await page.locator("#listing .item").count();
        await page.locator(".fb-content-main").evaluate((el) => {
          el.scrollTop = el.scrollHeight;
        });
        if (count < 180) {
          await expect
            .poll(() => page.locator("#listing .item").count())
            .toBeGreaterThan(count);
        }
      }
      await page.locator(".fb-content-main").evaluate((el) => {
        el.scrollTop = el.scrollHeight * 0.65;
      });
      await expect
        .poll(async () => (await visibleRegion(page)).scrollTop)
        .toBeGreaterThan(500);
      // Click a fully visible row/card so Playwright does not scroll it into view.
      await page.locator(".fb-content-main").evaluate((el) => {
        const top = el.getBoundingClientRect().top;
        const first = Array.from(
          el.querySelectorAll<HTMLElement>("#listing .item")
        ).find((item) => item.getBoundingClientRect().bottom > top);
        if (first) el.scrollTop += first.getBoundingClientRect().top - top;
      });
    });

    test("closing a preview restores the same region and rendered batch", async ({
      page,
    }) => {
      const before = await visibleRegion(page);
      const count = await page.locator("#listing .item").count();
      const target = await page
        .locator("#listing")
        .evaluate(
          (el) =>
            Array.from(
              el.querySelectorAll<HTMLElement>(".item[data-path]")
            ).find((item) => item.getBoundingClientRect().top > 120)?.dataset
              .path
        );
      await page.locator(`#listing .item[data-path="${target}"]`).dblclick();
      await expect(page.locator(".fb-preview-name")).toBeVisible();
      await page.keyboard.press("Escape");
      await expect(page.locator("#listing")).toBeVisible();
      await expect
        .poll(async () => (await visibleRegion(page)).path)
        .toBe(before.path);
      await expect
        .poll(async () =>
          Math.abs((await visibleRegion(page)).offset - before.offset)
        )
        .toBeLessThan(2);
      expect(
        await page.locator("#listing .item").count()
      ).toBeGreaterThanOrEqual(count);
    });

    test("deleting the first visible file keeps its surviving neighbor in place", async ({
      page,
    }) => {
      const before = await visibleRegion(page);
      const item = page.locator(`#listing .item[data-path="${before.path}"]`);
      const neighbor = await item.evaluate((el) => {
        const items = Array.from(
          document.querySelectorAll<HTMLElement>("#listing .item")
        );
        const next = items[items.indexOf(el as HTMLElement) + 1];
        return {
          path: next.dataset.path!,
          top: next.getBoundingClientRect().top,
        };
      });
      // The first partially visible row can sit underneath the sticky toolbar.
      // Select without the test runner's automatic scroll-into-view.
      await item.dispatchEvent("click");
      await page.keyboard.press("Delete");
      await expect(page.locator("#focus-prompt")).toBeVisible();
      await page.locator("#focus-prompt").click();
      await expect(item).toHaveCount(0);
      const surviving = page.locator(
        `#listing .item[data-path="${neighbor.path}"]`
      );
      await expect(surviving).toBeVisible();
      await expect
        .poll(async () =>
          Math.abs((await surviving.boundingBox())!.y - neighbor.top)
        )
        .toBeLessThan(2);
      expect((await visibleRegion(page)).scrollTop).toBeGreaterThan(500);
    });

    test("renaming a file does not scroll to its new sorted position", async ({
      page,
    }) => {
      const before = await visibleRegion(page);
      const target = await page
        .locator("#listing")
        .evaluate(
          (el) =>
            Array.from(
              el.querySelectorAll<HTMLElement>(".item[data-path]")
            ).find((item) => item.getBoundingClientRect().top > 120)?.dataset
              .path
        );
      await page.locator(`#listing .item[data-path="${target}"]`).click();
      await page.keyboard.press("F2");
      await expect(page.locator("input#focus-prompt")).toBeVisible();
      await page.locator("input#focus-prompt").fill("zzz-renamed.txt");
      await page.locator("input#focus-prompt").press("Enter");
      await expect(page.locator("input#focus-prompt")).toHaveCount(0);
      await expect(
        page.locator(`#listing .item[data-path="${target}"]`)
      ).toHaveCount(0);
      await expect
        .poll(async () => (await visibleRegion(page)).path)
        .toBe(before.path);
      await expect
        .poll(async () =>
          Math.abs((await visibleRegion(page)).offset - before.offset)
        )
        .toBeLessThan(2);
    });
  });
}
