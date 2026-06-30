// Thumbnail concurrency limiter.
// Caps simultaneous image thumbnail requests to MAX_CONCURRENT to prevent
// CPU/network spikes when browsing folders with many images.

const MAX_CONCURRENT = 6;

let active = 0;
const queue: Array<() => void> = [];

// Deduplicates in-flight requests for the same URL so multiple grid items
// waiting on the same thumbnail only fire one HTTP request.
const pending = new Map<string, Promise<void>>();

function processNext() {
  if (active >= MAX_CONCURRENT || queue.length === 0) return;
  active++;
  queue.shift()!();
}

export function loadThumbnail(img: HTMLImageElement, url: string): void {
  if (!url) return;
  // Already loaded with this URL — nothing to do.
  if (img.src === url || img.getAttribute("data-thumb-url") === url) return;

  img.setAttribute("data-thumb-url", url);

  if (pending.has(url)) {
    // Another request for this URL is already in flight — piggyback on it.
    pending.get(url)!.then(() => {
      if (img.getAttribute("data-thumb-url") === url) img.src = url;
    });
    return;
  }

  const p = new Promise<void>((resolve) => {
    queue.push(() => {
      const done = () => {
        active--;
        pending.delete(url);
        resolve();
        processNext();
      };
      img.onload = done;
      img.onerror = done;
      img.src = url;
    });
  });

  pending.set(url, p);
  processNext();
}
