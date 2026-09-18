import { onBeforeUnmount, watch, type Ref } from "vue";
import { useFileStore } from "@/stores/file";
import { hitIndices, type SelectionBox } from "@/utils/selection";

const DRAG_THRESHOLD_PX = 5;
const SCROLL_ZONE_PX = 48;
const SCROLL_MAX_SPEED = 24;

// Interactive areas that must keep their own click/drag behavior.
const IGNORE_SELECTOR =
  ".item, button, a, input, textarea, select, label, h2, .fb-col-header, .fb-trash-group-row";

/**
 * Desktop-style rubber-band selection for a file listing. Attach the listing
 * element ref; dragging from empty space draws a marquee and selects every
 * rendered item it touches. Ctrl/Cmd adds to the selection captured on
 * mousedown, plain drag replaces it. Auto-scrolls near the top/bottom edges.
 */
export function useMarqueeSelection(listingRef: Ref<HTMLElement | null>) {
  const fileStore = useFileStore();

  let startX = 0;
  let startY = 0;
  let curX = 0;
  let curY = 0;
  let additive = false;
  let baseSelection: number[] = [];
  let dragging = false;
  let marqueeEl: HTMLElement | null = null;
  let rafId: number | null = null;
  let swallowHandler: ((event: MouseEvent) => void) | null = null;

  const scroller = (): HTMLElement => {
    const main = document.querySelector("main");
    if (main && main.scrollHeight > main.clientHeight + 4) {
      return main as HTMLElement;
    }
    return document.documentElement;
  };

  const marqueeBox = (): SelectionBox => ({
    left: Math.min(startX, curX),
    top: Math.min(startY, curY),
    right: Math.max(startX, curX),
    bottom: Math.max(startY, curY),
  });

  /** Position the overlay and sync the selection with the items it covers. */
  const paint = () => {
    const listing = listingRef.value;
    if (!listing || !marqueeEl) return;

    const rect = listing.getBoundingClientRect();
    const box = marqueeBox();
    marqueeEl.style.left = `${box.left - rect.left}px`;
    marqueeEl.style.top = `${box.top - rect.top}px`;
    marqueeEl.style.width = `${box.right - box.left}px`;
    marqueeEl.style.height = `${box.bottom - box.top}px`;

    const hits = hitIndices(listing, box);
    const next = additive
      ? Array.from(new Set([...baseSelection, ...hits])).sort((a, b) => a - b)
      : hits;

    const same =
      next.length === fileStore.selected.length &&
      next.every((value, i) => value === fileStore.selected[i]);
    if (!same) fileStore.selected = next;
  };

  /** Edge auto-scroll loop: keeps scrolling while the pointer rests in a zone. */
  const tick = () => {
    rafId = null;
    if (!dragging) return;

    const el = scroller();
    const rect = el.getBoundingClientRect();
    const viewH = el.clientHeight || window.innerHeight;
    const zone = Math.min(SCROLL_ZONE_PX, viewH / 3);

    const fromTop = curY - rect.top;
    const fromBottom = rect.bottom - curY;
    let speed = 0;
    if (fromTop >= 0 && fromTop < zone) {
      speed = -Math.max(4, Math.round(SCROLL_MAX_SPEED * (1 - fromTop / zone)));
    } else if (fromBottom >= 0 && fromBottom < zone) {
      speed = Math.max(
        4,
        Math.round(SCROLL_MAX_SPEED * (1 - fromBottom / zone))
      );
    }

    if (speed !== 0) {
      el.scrollTop += speed;
      paint();
    }
    rafId = requestAnimationFrame(tick);
  };

  const start = () => {
    const listing = listingRef.value;
    if (!listing) return;
    dragging = true;
    fileStore.selectionAnchor = null;

    const el = document.createElement("div");
    el.className = "fb-marquee";
    listing.appendChild(el);
    marqueeEl = el;

    document.body.classList.add("fb-marquee-dragging");
    paint();
    rafId = requestAnimationFrame(tick);
  };

  const stop = () => {
    dragging = false;
    if (rafId !== null) {
      cancelAnimationFrame(rafId);
      rafId = null;
    }
    marqueeEl?.remove();
    marqueeEl = null;
    document.body.classList.remove("fb-marquee-dragging");
  };

  const removeSwallow = () => {
    if (swallowHandler) {
      document.removeEventListener("click", swallowHandler, true);
      swallowHandler = null;
    }
  };

  /**
   * A marquee drag still fires a `click` on mouseup. Swallow exactly that one
   * click so the listing's empty-area handler doesn't wipe the new selection.
   */
  const suppressNextClick = () => {
    removeSwallow();
    swallowHandler = (event: MouseEvent) => {
      event.preventDefault();
      event.stopPropagation();
      removeSwallow();
    };
    document.addEventListener("click", swallowHandler, true);
  };

  const detachDocument = () => {
    document.removeEventListener("mousemove", onMouseMove, true);
    document.removeEventListener("mouseup", onMouseUp, true);
    document.removeEventListener("keydown", onKeyDown, true);
    window.removeEventListener("blur", onCancel);
  };

  const onMouseMove = (event: MouseEvent) => {
    curX = event.clientX;
    curY = event.clientY;

    if (!dragging) {
      const dx = Math.abs(curX - startX);
      const dy = Math.abs(curY - startY);
      if (dx < DRAG_THRESHOLD_PX && dy < DRAG_THRESHOLD_PX) return;
      start();
      return;
    }

    event.preventDefault();
    paint();
  };

  const onMouseUp = () => {
    detachDocument();
    if (!dragging) return;

    // Shift-click after a marquee extends from its first (top-left) item.
    const listing = listingRef.value;
    const hits = listing ? hitIndices(listing, marqueeBox()) : [];
    fileStore.selectionAnchor = hits.length > 0 ? hits[0] : null;

    suppressNextClick();
    stop();
  };

  const onCancel = () => {
    detachDocument();
    stop();
  };

  /** Esc aborts the drag without swallowing the trailing click. */
  const onKeyDown = (event: KeyboardEvent) => {
    if (event.key !== "Escape" || !dragging) return;
    detachDocument();
    removeSwallow();
    stop();
  };

  const onMouseDown = (event: MouseEvent) => {
    if (event.button !== 0 || event.defaultPrevented) return;

    const target = event.target as Element | null;
    if (!target || !target.closest) return;
    if (target.closest(IGNORE_SELECTOR)) return;

    removeSwallow();
    startX = event.clientX;
    startY = event.clientY;
    curX = startX;
    curY = startY;
    additive = event.ctrlKey || event.metaKey;
    baseSelection = additive ? [...fileStore.selected] : [];
    dragging = false;

    event.preventDefault();
    document.addEventListener("mousemove", onMouseMove, true);
    document.addEventListener("mouseup", onMouseUp, true);
    document.addEventListener("keydown", onKeyDown, true);
    window.addEventListener("blur", onCancel);
  };

  const attach = (el: HTMLElement | null) => {
    el?.addEventListener("mousedown", onMouseDown);
  };

  const detach = (el: HTMLElement | null) => {
    el?.removeEventListener("mousedown", onMouseDown);
  };

  watch(
    listingRef,
    (el, oldEl) => {
      detach(oldEl ?? null);
      attach(el);
    },
    { immediate: true }
  );

  onBeforeUnmount(() => {
    detach(listingRef.value);
    detachDocument();
    removeSwallow();
    stop();
  });
}
