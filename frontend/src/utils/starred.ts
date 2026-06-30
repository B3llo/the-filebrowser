import { ref } from "vue";
import { useAuthStore } from "@/stores/auth";
import { users } from "@/api";

export interface StarredFile {
  url: string;
  name: string;
  type: string;
  starredAt: number;
}

// Global reactive counter — bump it inside toggleStarred so any computed
// that reads starVersion.value automatically re-evaluates.
export const starVersion = ref(0);

const KEY = "fb-starred-files";
const MAX = 500;

export function getStarred(): StarredFile[] {
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

export function isStarred(url: string): boolean {
  return getStarred().some((r) => r.url === url);
}

/** Toggles the starred state. Returns the new starred state (true = now starred). */
export function toggleStarred(file: Omit<StarredFile, "starredAt">): boolean {
  if (!file.url || !file.name) return false;
  try {
    const list = getStarred();
    const idx = list.findIndex((r) => r.url === file.url);
    let updated: StarredFile[];
    let nowStarred: boolean;

    if (idx !== -1) {
      updated = [...list.slice(0, idx), ...list.slice(idx + 1)];
      nowStarred = false;
    } else {
      updated = [{ ...file, starredAt: Date.now() }, ...list].slice(0, MAX);
      nowStarred = true;
    }

    localStorage.setItem(KEY, JSON.stringify(updated));
    starVersion.value++;

    // Write-through to backend (fire-and-forget — local state is authoritative for UX)
    const authStore = useAuthStore();
    if (authStore.user?.id) {
      authStore.updateUser({ starred: updated });
      users
        .update({ id: authStore.user.id, starred: updated }, ["starred"])
        .catch(() => {/* ignore sync errors */});
    }

    return nowStarred;
  } catch {
    return false;
  }
}
