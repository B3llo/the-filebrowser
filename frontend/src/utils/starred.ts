import { ref } from "vue";

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
    if (idx !== -1) {
      list.splice(idx, 1);
      localStorage.setItem(KEY, JSON.stringify(list));
      starVersion.value++;
      return false;
    }
    list.unshift({ ...file, starredAt: Date.now() });
    localStorage.setItem(KEY, JSON.stringify(list.slice(0, MAX)));
    starVersion.value++;
    return true;
  } catch {
    return false;
  }
}
