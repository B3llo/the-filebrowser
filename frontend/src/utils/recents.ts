// Lightweight "recent files" list backed by localStorage. There is no backend
// recents API yet (the sidebar "Recent" item is still a placeholder), so the
// command palette records files as they are previewed and reads the most
// recent ones for its empty state.

export interface RecentFile {
  // Router path used to navigate back to the file (already source-prefixed).
  url: string;
  name: string;
  // Resource type, used to pick an icon.
  type: string;
  // Epoch millis of the last visit, for ordering.
  at: number;
}

const KEY = "fb-recent-files";
const MAX = 12;

export function getRecents(): RecentFile[] {
  try {
    const raw = localStorage.getItem(KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed.filter(
      (r) => r && typeof r.url === "string" && typeof r.name === "string"
    );
  } catch {
    return [];
  }
}

export function addRecent(file: Omit<RecentFile, "at">): void {
  if (!file.url || !file.name) return;
  try {
    const list = getRecents().filter((r) => r.url !== file.url);
    list.unshift({ ...file, at: Date.now() });
    localStorage.setItem(KEY, JSON.stringify(list.slice(0, MAX)));
  } catch {
    // Storage full or unavailable — recents are best-effort, ignore.
  }
}
