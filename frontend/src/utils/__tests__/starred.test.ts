import { describe, it, expect, beforeEach, vi } from "vitest";
import type { StarredFile } from "@/utils/starred";

vi.mock("@/stores/auth", () => ({
  useAuthStore: () => ({}),
}));

vi.mock("@/api", () => ({
  users: { update: vi.fn() },
}));

const KEY = "fb-starred-files";

class MemoryStorage implements Storage {
  private store = new Map<string, string>();
  get length() {
    return this.store.size;
  }
  clear() {
    this.store.clear();
  }
  getItem(key: string) {
    return this.store.get(key) ?? null;
  }
  key(index: number) {
    return Array.from(this.store.keys())[index] ?? null;
  }
  removeItem(key: string) {
    this.store.delete(key);
  }
  setItem(key: string, value: string) {
    this.store.set(key, value);
  }
}

beforeEach(() => {
  vi.stubGlobal("localStorage", new MemoryStorage());
  vi.stubGlobal(
    "window",
    Object.assign({
      FileBrowser: {
        Name: "File Browser",
        BaseURL: "",
        StaticURL: "/static",
        TusSettings: {},
      },
      location: { origin: "http://localhost" },
    })
  );
});

function star(url: string, name: string) {
  const list: StarredFile[] = JSON.parse(localStorage.getItem(KEY) ?? "[]");
  list.push({ url, name, type: "blob", starredAt: Date.now() });
  localStorage.setItem(KEY, JSON.stringify(list));
}

type Item = { url: string; name: string };

const items: Item[] = [
  { url: "/files/0/beta.txt", name: "beta.txt" },
  { url: "/files/0/alpha.txt", name: "alpha.txt" },
  { url: "/files/0/zeta.txt", name: "zeta.txt" },
  { url: "/files/0/mid.txt", name: "mid.txt" },
];

describe("sortStarredFirst", () => {
  it("keeps order unchanged when nothing is starred", async () => {
    const { sortStarredFirst } = await import("@/utils/starred");
    expect(sortStarredFirst(items)).toEqual(items);
  });

  it("pins starred items to the front sorted alphabetically", async () => {
    const { sortStarredFirst } = await import("@/utils/starred");
    star("/files/0/zeta.txt", "zeta.txt");
    star("/files/0/beta.txt", "beta.txt");

    expect(sortStarredFirst(items).map((i) => i.name)).toEqual([
      "beta.txt",
      "zeta.txt",
      "alpha.txt",
      "mid.txt",
    ]);
  });

  it("keeps non-starred items in their original relative order", async () => {
    const { sortStarredFirst } = await import("@/utils/starred");
    star("/files/0/mid.txt", "mid.txt");

    expect(sortStarredFirst(items).map((i) => i.name)).toEqual([
      "mid.txt",
      "beta.txt",
      "alpha.txt",
      "zeta.txt",
    ]);
  });

  it("returns original order once items are unstarred", async () => {
    const { sortStarredFirst } = await import("@/utils/starred");
    star("/files/0/alpha.txt", "alpha.txt");
    expect(sortStarredFirst(items).map((i) => i.name)[0]).toBe("alpha.txt");

    localStorage.removeItem(KEY);
    expect(sortStarredFirst(items)).toEqual(items);
  });
});
