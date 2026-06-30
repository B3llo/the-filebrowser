import { defineStore } from "pinia";
import { sources as sourcesApi } from "@/api";
import { setActiveSourceId } from "@/api/utils";

function writeSourceCookie(id: string) {
  document.cookie = `source=${id}; Path=/; SameSite=Strict;`;
}

function clearSourceCookie() {
  document.cookie = "source=; Max-Age=0; Path=/; SameSite=Strict;";
}

export const useSourceStore = defineStore("source", {
  state: (): {
    sources: ISource[];
    activeId: string;
    loaded: boolean;
  } => ({
    sources: [],
    // "0" is the implicit legacy source (the classic user scope).
    activeId: "0",
    loaded: false,
  }),
  getters: {
    active(state): ISource | undefined {
      return (
        state.sources.find((s) => String(s.id) === state.activeId) ??
        state.sources[0]
      );
    },
    defaultId(state): string {
      return String(state.sources[0]?.id ?? "0");
    },
    hasMultiple(state): boolean {
      return state.sources.length > 1;
    },
    // The /files base path for the active source, used by breadcrumbs and the
    // header toolbar.
    filesBase(state): string {
      const id = String(this.active?.id ?? state.sources[0]?.id ?? "0");
      return `/files/${id}`;
    },
  },
  actions: {
    async load() {
      try {
        this.sources = await sourcesApi.list();
      } catch (error) {
        // If the sources endpoint is unreachable, fall back to the single
        // implicit legacy source so navigation never gets stuck waiting. The
        // backend resolves id 0 to the classic user scope.
        console.error("failed to load sources, using implicit source", error);
        this.sources = [{ id: 0, name: "My Files" }];
      }
      this.loaded = true;
      // Pick the active source from the cookie if it is still valid, otherwise
      // the default (first) source.
      const cookieId = readSourceCookie();
      const valid =
        cookieId !== "" && this.sources.some((s) => String(s.id) === cookieId);
      const id = valid ? cookieId! : this.defaultId;
      this.setActive(id);
    },
    setActive(id: string) {
      this.activeId = id;
      setActiveSourceId(id);
      writeSourceCookie(id);
    },
    isKnown(idStr: string): boolean {
      // The implicit legacy source (id 0) is always valid, even before sources
      // have been loaded from the backend.
      if (idStr === "0") return true;
      return this.sources.some((s) => String(s.id) === idStr);
    },
    reset() {
      this.sources = [];
      this.activeId = "0";
      this.loaded = false;
      setActiveSourceId("0");
      clearSourceCookie();
    },
  },
});

function readSourceCookie(): string {
  const match = document.cookie.match(/(?:^|;\s*)source\s*=\s*([^;]+)/);
  return match ? match[1] : "";
}
