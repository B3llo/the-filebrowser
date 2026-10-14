import type { InjectionKey } from "vue";

export interface ListingScrollPosition {
  scrollTop: number;
  showLimit: number;
  anchors: { path: string; offset: number }[];
}

// Owned by the Files view: survives preview/editor switches, but not logout.
export const listingScrollKey: InjectionKey<
  Map<string, ListingScrollPosition>
> = Symbol("listingScroll");

export function listingScroller(listing: HTMLElement): HTMLElement {
  return (
    listing.closest<HTMLElement>(".fb-content-main") ??
    listing.closest<HTMLElement>("main") ??
    (document.scrollingElement as HTMLElement | null) ??
    document.documentElement
  );
}

export function captureListingScroll(
  listing: HTMLElement,
  showLimit: number
): ListingScrollPosition {
  const scroller = listingScroller(listing);
  const rect = scroller.getBoundingClientRect();
  const anchors: ListingScrollPosition["anchors"] = [];
  for (const item of listing.querySelectorAll<HTMLElement>(
    ".item[data-path]"
  )) {
    const itemRect = item.getBoundingClientRect();
    if (itemRect.bottom <= rect.top || itemRect.top >= rect.bottom) continue;
    // Keep neighbors too: the first visible file may be deleted or moved.
    anchors.push({ path: item.dataset.path!, offset: itemRect.top - rect.top });
  }
  return { scrollTop: scroller.scrollTop, showLimit, anchors };
}

export function restoreListingScroll(
  listing: HTMLElement,
  position: ListingScrollPosition
): void {
  const scroller = listingScroller(listing);
  const items = new Map(
    Array.from(
      listing.querySelectorAll<HTMLElement>(".item[data-path]"),
      (item) => [item.dataset.path!, item]
    )
  );
  for (const anchor of position.anchors) {
    const item = items.get(anchor.path);
    if (!item) continue;
    scroller.scrollTop +=
      item.getBoundingClientRect().top -
      scroller.getBoundingClientRect().top -
      anchor.offset;
    return;
  }
  // No surviving anchor (e.g. a whole selection removed): retain the offset,
  // naturally clamped by the browser when the directory has become shorter.
  scroller.scrollTop = position.scrollTop;
}
