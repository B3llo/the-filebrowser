// @vitest-environment jsdom
import { describe, it, expect, beforeEach } from "vitest";
import {
  visualOrder,
  isIndexRendered,
  rangeIndices,
  boxesIntersect,
  hitIndices,
  type SelectionBox,
} from "@/utils/selection";

const rect = (left: number, top: number, right: number, bottom: number) =>
  ({ left, top, right, bottom }) as SelectionBox;

function stubRect(el: Element, box: SelectionBox) {
  el.getBoundingClientRect = () =>
    ({
      ...box,
      x: box.left,
      y: box.top,
      width: box.right - box.left,
      height: box.bottom - box.top,
      toJSON: () => box,
    }) as DOMRect;
}

// Mirrors a "modified" sort in FileListing: the dirs and files sections are
// laid out separately, so DOM order (0, 2, 1, 3) differs from index order.
const FIXTURE = `
  <div id="listing">
    <div class="fb-items--folders">
      <div class="item" data-index="0" data-dir="true"></div>
      <div class="item" data-index="2" data-dir="true"></div>
    </div>
    <div class="fb-items--files">
      <div class="item" data-index="1"></div>
      <div class="item" data-index="3"></div>
    </div>
    <div class="item">load more</div>
  </div>
`;

describe("visualOrder", () => {
  beforeEach(() => {
    document.body.innerHTML = FIXTURE;
  });

  it("reads rendered item indices in DOM order", () => {
    expect(visualOrder()).toEqual([0, 2, 1, 3]);
  });

  it("ignores items without a data-index", () => {
    expect(visualOrder()).toHaveLength(4);
  });
});

describe("isIndexRendered", () => {
  beforeEach(() => {
    document.body.innerHTML = FIXTURE;
  });

  it("is true for rendered indices and false otherwise", () => {
    expect(isIndexRendered(2)).toBe(true);
    expect(isIndexRendered(9)).toBe(false);
  });
});

describe("rangeIndices", () => {
  beforeEach(() => {
    document.body.innerHTML = FIXTURE;
  });

  it("follows visual order instead of raw index order", () => {
    expect(rangeIndices(0, 3)).toEqual([0, 2, 1, 3]);
  });

  it("works in reverse (dragging backwards)", () => {
    expect(rangeIndices(3, 0)).toEqual([0, 2, 1, 3]);
  });

  it("returns the two-element range between adjacent items", () => {
    expect(rangeIndices(2, 1)).toEqual([2, 1]);
  });

  it("falls back to a numeric range when the anchor is not rendered", () => {
    expect(rangeIndices(8, 10)).toEqual([8, 9, 10]);
  });

  it("returns a single index for the anchor itself", () => {
    expect(rangeIndices(1, 1)).toEqual([1]);
  });
});

describe("boxesIntersect", () => {
  it("detects overlap and rejects separation", () => {
    expect(boxesIntersect(rect(0, 0, 10, 10), rect(5, 5, 15, 15))).toBe(true);
    expect(boxesIntersect(rect(0, 0, 10, 10), rect(20, 0, 30, 10))).toBe(false);
  });

  it("treats touching edges as intersecting", () => {
    expect(boxesIntersect(rect(0, 0, 10, 10), rect(10, 10, 20, 20))).toBe(true);
  });
});

describe("hitIndices", () => {
  beforeEach(() => {
    document.body.innerHTML = FIXTURE;
    const items = document.querySelectorAll<HTMLElement>("#listing .item");
    const boxes: Record<number, SelectionBox> = {
      0: rect(0, 0, 100, 50),
      2: rect(0, 60, 100, 110),
      1: rect(0, 120, 100, 170),
      3: rect(0, 180, 100, 230),
    };
    items.forEach((el) => {
      const index = Number(el.dataset.index);
      if (Number.isFinite(index) && boxes[index]) stubRect(el, boxes[index]);
    });
  });

  it("returns every item intersecting the marquee, in DOM order", () => {
    expect(hitIndices(document, rect(10, 10, 90, 130))).toEqual([0, 2, 1]);
  });

  it("returns an empty list when nothing is covered", () => {
    expect(hitIndices(document, rect(500, 500, 600, 600))).toEqual([]);
  });
});
