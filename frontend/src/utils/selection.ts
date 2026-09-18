export interface SelectionBox {
  left: number;
  top: number;
  right: number;
  bottom: number;
}

const ITEM_SELECTOR = "#listing .item[data-index]";

/** Index of every rendered listing item, in visual (DOM) order. */
export function visualOrder(): number[] {
  return Array.from(document.querySelectorAll<HTMLElement>(ITEM_SELECTOR)).map(
    (el) => Number(el.dataset.index)
  );
}

/** Whether an item with this index is currently rendered in the listing. */
export function isIndexRendered(index: number): boolean {
  return visualOrder().indexOf(index) !== -1;
}

/**
 * Indices covered by the shift-click range between `anchor` and `target`.
 * Uses the VISUAL order of the rendered items, so it stays correct when
 * directories and files are laid out in separate sections (mosaic/list,
 * "modified" sorting), unlike a raw numeric range over `ResourceItem.index`.
 * Falls back to a numeric range when either end is not rendered (filtered
 * out or still paginated away).
 */
export function rangeIndices(anchor: number, target: number): number[] {
  const order = visualOrder();
  const a = order.indexOf(anchor);
  const b = order.indexOf(target);

  if (a === -1 || b === -1) {
    const lo = Math.min(anchor, target);
    const hi = Math.max(anchor, target);
    const out: number[] = [];
    for (let i = lo; i <= hi; i++) out.push(i);
    return out;
  }

  const [lo, hi] = a <= b ? [a, b] : [b, a];
  return order.slice(lo, hi + 1);
}

export function boxesIntersect(a: SelectionBox, b: SelectionBox): boolean {
  return !(
    a.right < b.left ||
    a.left > b.right ||
    a.bottom < b.top ||
    a.top > b.bottom
  );
}

/**
 * Indices of every rendered item whose box intersects `box`. Coordinates are
 * viewport-based (getBoundingClientRect), matching the marquee rectangle.
 */
export function hitIndices(container: ParentNode, box: SelectionBox): number[] {
  const hits: number[] = [];
  container.querySelectorAll<HTMLElement>(".item[data-index]").forEach((el) => {
    const rect = el.getBoundingClientRect();
    const itemBox: SelectionBox = {
      left: rect.left,
      top: rect.top,
      right: rect.right,
      bottom: rect.bottom,
    };
    if (!boxesIntersect(itemBox, box)) return;
    const index = Number(el.dataset.index);
    if (Number.isFinite(index)) hits.push(index);
  });
  return hits;
}
