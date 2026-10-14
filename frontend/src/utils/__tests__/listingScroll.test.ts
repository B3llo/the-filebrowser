// @vitest-environment jsdom
import { afterEach, describe, expect, it } from "vitest";
import { captureListingScroll, restoreListingScroll } from "../listingScroll";

function rect(top: number, height: number): DOMRect {
  return {
    top,
    bottom: top + height,
    height,
    left: 0,
    right: 300,
    width: 300,
    x: 0,
    y: top,
    toJSON() {},
  };
}

function fixture(paths = ["/a.txt", "/b.txt", "/c.txt"]) {
  const scroller = document.createElement("div");
  scroller.className = "fb-content-main";
  scroller.getBoundingClientRect = () => rect(100, 100);
  const listing = document.createElement("div");
  scroller.append(listing);
  document.body.append(scroller);
  const items = paths.map((path, index) => {
    const item = document.createElement("div");
    item.className = "item";
    item.dataset.path = path;
    item.getBoundingClientRect = () => rect(90 + index * 40, 40);
    listing.append(item);
    return item;
  });
  scroller.scrollTop = 900;
  return { scroller, listing, items };
}

afterEach(() => document.body.replaceChildren());

describe("listing scroll position", () => {
  it("captures partially visible rows and the loaded batch", () => {
    const { listing } = fixture();
    expect(captureListingScroll(listing, 180)).toEqual({
      scrollTop: 900,
      showLimit: 180,
      anchors: [
        { path: "/a.txt", offset: -10 },
        { path: "/b.txt", offset: 30 },
        { path: "/c.txt", offset: 70 },
      ],
    });
  });

  it("uses a surviving neighbor when the first visible file is removed", () => {
    const { scroller, listing, items } = fixture();
    const position = captureListingScroll(listing, 180);
    items[0].remove();
    items[1].getBoundingClientRect = () => rect(490, 40);
    scroller.scrollTop = 0;
    restoreListingScroll(listing, position);
    expect(scroller.scrollTop).toBe(360);
  });

  it("falls back to the saved vertical offset when all anchors disappear", () => {
    const { scroller, listing } = fixture();
    const position = captureListingScroll(listing, 180);
    listing.replaceChildren();
    scroller.scrollTop = 0;
    restoreListingScroll(listing, position);
    expect(scroller.scrollTop).toBe(900);
  });

  it("matches paths literally, including quotes and selector characters", () => {
    const { scroller, listing, items } = fixture(['/[photo]".png']);
    const position = captureListingScroll(listing, 50);
    items[0].getBoundingClientRect = () => rect(140, 40);
    scroller.scrollTop = 0;
    restoreListingScroll(listing, position);
    expect(scroller.scrollTop).toBe(50);
  });
});
