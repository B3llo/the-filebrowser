import { useAuthStore } from "@/stores/auth";
import { users } from "@/api";

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

let syncTimer: ReturnType<typeof setTimeout> | null = null;

export function addRecent(file: Omit<RecentFile, "at">): void {
  if (!file.url || !file.name) return;
  try {
    const list = getRecents().filter((r) => r.url !== file.url);
    const updated = [{ ...file, at: Date.now() }, ...list].slice(0, MAX);
    localStorage.setItem(KEY, JSON.stringify(updated));

    // Debounced write-through to backend (batches rapid navigation)
    if (syncTimer) clearTimeout(syncTimer);
    syncTimer = setTimeout(() => {
      const authStore = useAuthStore();
      if (authStore.user?.id) {
        authStore.updateUser({ recents: updated });
        users
          .update({ id: authStore.user.id, recents: updated }, ["recents"])
          .catch(() => {/* ignore sync errors */});
      }
    }, 500);
  } catch {
    // Storage full or unavailable — recents are best-effort, ignore.
  }
}
