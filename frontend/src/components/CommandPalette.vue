<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="fb-cmdk-backdrop"
      role="dialog"
      aria-modal="true"
      :aria-label="t('commandPalette.placeholder')"
      @click="close"
    >
      <div class="fb-cmdk" @click.stop @keydown="onKeydown">
        <div class="fb-cmdk-input-row">
          <FbIcon name="search" size="18px" class="fb-cmdk-search-icon" />
          <input
            ref="inputEl"
            v-model="query"
            class="fb-cmdk-input"
            type="text"
            autocomplete="off"
            spellcheck="false"
            :placeholder="t('commandPalette.placeholder')"
            :aria-label="t('commandPalette.placeholder')"
          />
          <span v-if="scopeName" class="fb-cmdk-scope">{{ scopeName }}</span>
        </div>

        <div class="fb-cmdk-filters">
          <button
            v-for="f in typeFilters"
            :key="f.key"
            type="button"
            class="fb-cmdk-filter"
            :class="{ active: activeType === f.key }"
            @click="toggleType(f.key)"
          >
            <FbIcon :name="f.icon" size="14px" />
            <span>{{ t("search." + f.label) }}</span>
          </button>
        </div>

        <div ref="listEl" class="fb-cmdk-list">
          <template v-if="flat.length">
            <template v-for="section in sections" :key="section.id">
              <div class="fb-cmdk-section">{{ section.header }}</div>
              <button
                v-for="item in section.items"
                :key="item.key"
                type="button"
                class="fb-cmdk-row"
                :class="{ active: item.index === activeIndex }"
                :data-index="item.index"
                @click="item.run()"
                @mousemove="activeIndex = item.index"
              >
                <FbIcon
                  :name="item.icon"
                  size="18px"
                  class="fb-cmdk-row-icon"
                />
                <span
                  class="fb-cmdk-row-label"
                  v-html="highlight(item.label, item.query)"
                ></span>
                <span v-if="item.sub" class="fb-cmdk-row-sub">{{
                  item.sub
                }}</span>
              </button>
            </template>
          </template>
          <div v-else-if="searching" class="fb-cmdk-status">
            {{ t("commandPalette.searching") }}
          </div>
          <div v-else-if="inSearchMode" class="fb-cmdk-status">
            {{ t("commandPalette.noResults") }}
          </div>
        </div>

        <div class="fb-cmdk-footer">
          <span
            ><kbd>↑</kbd><kbd>↓</kbd>
            {{ t("commandPalette.hintNavigate") }}</span
          >
          <span><kbd>↵</kbd> {{ t("commandPalette.hintOpen") }}</span>
          <span><kbd>esc</kbd> {{ t("commandPalette.hintClose") }}</span>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { useLayoutStore } from "@/stores/layout";
import { useSourceStore } from "@/stores/source";
import { search } from "@/api";
import FbIcon from "@/components/FbIcon.vue";
import type { IconName } from "@/utils/icons";
import { getRecents } from "@/utils/recents";

interface Entry {
  key: string;
  icon: IconName;
  label: string;
  sub?: string;
  // The current query, when this entry should highlight a match in its label.
  query?: string;
  run: () => void;
  // Flat index assigned at build time for keyboard navigation.
  index: number;
}

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const layoutStore = useLayoutStore();
const sourceStore = useSourceStore();

const open = computed(() => layoutStore.commandPalette);
const scopeName = computed(() => sourceStore.active?.name ?? "");
const inFiles = computed(() => route.path.startsWith("/files"));

// The backend search streams only { dir, path }; search.ts adds url. There is
// no name/type, so we derive them from the path and the extension.
interface SearchResult {
  dir: boolean;
  path: string;
  url: string;
}

// Type filter chips (Finder-style). Keys map to the backend `type:` query
// syntax: image/audio/video are mime-based, pdf is an extension condition.
const typeFilters: { key: string; label: string; icon: IconName }[] = [
  { key: "image", label: "images", icon: "image" },
  { key: "audio", label: "music", icon: "audio" },
  { key: "video", label: "video", icon: "video" },
  { key: "pdf", label: "pdf", icon: "pdf" },
];

const query = ref("");
const activeType = ref<string | null>(null);
const results = ref<SearchResult[]>([]);
const searching = ref(false);
const activeIndex = ref(0);

const inputEl = ref<HTMLInputElement | null>(null);
const listEl = ref<HTMLElement | null>(null);

let controller: AbortController | null = null;
let debounce: number | null = null;
let opener: HTMLElement | null = null;

const IMAGE_EXTS = new Set([
  ".jpg",
  ".jpeg",
  ".png",
  ".gif",
  ".webp",
  ".bmp",
  ".svg",
  ".ico",
  ".avif",
  ".heic",
  ".tiff",
]);
const VIDEO_EXTS = new Set([
  ".mp4",
  ".mov",
  ".webm",
  ".mkv",
  ".avi",
  ".m4v",
  ".wmv",
  ".flv",
]);
const AUDIO_EXTS = new Set([
  ".mp3",
  ".wav",
  ".flac",
  ".ogg",
  ".m4a",
  ".aac",
  ".opus",
]);
const CODE_EXTS = new Set([
  ".js",
  ".mjs",
  ".cjs",
  ".ts",
  ".tsx",
  ".jsx",
  ".vue",
  ".go",
  ".py",
  ".rb",
  ".rs",
  ".java",
  ".kt",
  ".c",
  ".cpp",
  ".h",
  ".cs",
  ".php",
  ".sh",
  ".html",
  ".css",
  ".scss",
  ".json",
  ".xml",
  ".yml",
  ".yaml",
  ".toml",
  ".sql",
]);
const ARCHIVE_EXTS = new Set([
  ".zip",
  ".rar",
  ".7z",
  ".tar",
  ".gz",
  ".bz2",
  ".xz",
  ".tgz",
]);

const extOf = (name: string) => {
  const i = name.lastIndexOf(".");
  return i <= 0 ? "" : name.slice(i).toLowerCase();
};

// Icon from filename extension (search results carry no resource type).
const iconForName = (name: string, isDir: boolean): IconName => {
  if (isDir) return "folder";
  const ext = extOf(name);
  if (IMAGE_EXTS.has(ext)) return "image";
  if (VIDEO_EXTS.has(ext)) return "video";
  if (AUDIO_EXTS.has(ext)) return "audio";
  if (ext === ".pdf") return "pdf";
  if (CODE_EXTS.has(ext)) return "code";
  if (ARCHIVE_EXTS.has(ext)) return "archive";
  return "file";
};

// Split a raw (already-decoded) FS path into name + parent directory. The
// search backend returns raw paths, so we must NOT decodeURIComponent here.
const splitPath = (path: string): { name: string; dir: string } => {
  const clean = path.replace(/\/$/, "");
  const slash = clean.lastIndexOf("/");
  return slash === -1
    ? { name: clean, dir: "" }
    : { name: clean.slice(slash + 1), dir: clean.slice(0, slash) };
};

const openUrl = (url: string) => {
  close();
  router.push({ path: url });
};

const runPrompt = (name: string) => {
  close();
  layoutStore.showHover(name);
};

const actionEntries = (): Omit<Entry, "index">[] => {
  const items: Omit<Entry, "index">[] = [];
  if (inFiles.value) {
    items.push({
      key: "a:newFolder",
      icon: "new-folder",
      label: t("sidebar.newFolder"),
      run: () => runPrompt("newDir"),
    });
    items.push({
      key: "a:newFile",
      icon: "new-file",
      label: t("sidebar.newFile"),
      run: () => runPrompt("newFile"),
    });
    items.push({
      key: "a:upload",
      icon: "upload",
      label: t("buttons.upload"),
      run: () => runPrompt("upload"),
    });
  }
  items.push({
    key: "a:settings",
    icon: "settings",
    label: t("commandPalette.goToSettings"),
    run: () => openUrl("/settings"),
  });
  return items;
};

interface Section {
  id: string;
  header: string;
  items: Entry[];
}

// The query in "search mode" — a typed query or an active type chip.
const inSearchMode = computed(
  () => query.value.trim() !== "" || activeType.value !== null
);

// Substring to highlight in result names (the typed text, minus a leading dot
// used for the extension shortcut).
const highlightTerm = computed(() => {
  const raw = query.value.trim();
  return raw.startsWith(".") ? raw.slice(1) : raw;
});

const sections = computed<Section[]>(() => {
  const out: { id: string; header: string; items: Omit<Entry, "index">[] }[] =
    [];

  if (!inSearchMode.value) {
    const recents = getRecents();
    if (recents.length) {
      out.push({
        id: "recent",
        header: t("commandPalette.recent"),
        items: recents.slice(0, 8).map((r) => {
          const { dir } = splitPath(r.url.replace(/^\/files\/[^/]+\//, ""));
          return {
            key: "recent:" + r.url,
            icon: iconForName(r.name, false),
            label: r.name,
            sub: dir,
            run: () => openUrl(r.url),
          };
        }),
      });
    }
    out.push({
      id: "actions",
      header: t("commandPalette.actions"),
      items: actionEntries(),
    });
  } else {
    const toEntry = (item: SearchResult): Omit<Entry, "index"> => {
      const { name, dir } = splitPath(item.path);
      return {
        key: "res:" + item.url,
        icon: iconForName(name, item.dir),
        label: name,
        sub: dir,
        query: highlightTerm.value,
        run: () => openUrl(item.url),
      };
    };
    const folders = results.value.filter((r) => r.dir);
    const files = results.value.filter((r) => !r.dir);
    if (folders.length) {
      out.push({
        id: "folders",
        header: t("files.folders"),
        items: folders.map(toEntry),
      });
    }
    if (files.length) {
      out.push({
        id: "files",
        header: t("files.files"),
        items: files.map(toEntry),
      });
    }
  }

  // Assign flat indices in render order for keyboard navigation.
  let i = 0;
  return out.map((s) => ({
    ...s,
    items: s.items.map((it) => ({ ...it, index: i++ })),
  }));
});

const flat = computed<Entry[]>(() => sections.value.flatMap((s) => s.items));

const escapeHtml = (s: string) =>
  s.replace(
    /[&<>"]/g,
    (c) =>
      ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;" })[c] as string
  );

const highlight = (text: string, q?: string): string => {
  if (!q) return escapeHtml(text);
  const idx = text.toLowerCase().indexOf(q.toLowerCase());
  if (idx === -1) return escapeHtml(text);
  return (
    escapeHtml(text.slice(0, idx)) +
    "<mark>" +
    escapeHtml(text.slice(idx, idx + q.length)) +
    "</mark>" +
    escapeHtml(text.slice(idx + q.length))
  );
};

// Build the backend query string from the active type chip and the typed text.
//   - active chip      -> `type:<key>` (image/audio/video mime, pdf extension)
//   - text ".gitignore" -> `type:gitignore` (exact extension shortcut)
//   - text "report"     -> substring name match
const buildQuery = (): string => {
  const raw = query.value.trim();
  const parts: string[] = [];
  if (activeType.value) parts.push(`type:${activeType.value}`);
  if (/^\.[a-z0-9]+$/i.test(raw)) {
    parts.push(`type:${raw.slice(1)}`);
  } else if (raw) {
    parts.push(raw);
  }
  return parts.join(" ").trim();
};

const runSearch = async () => {
  const q = buildQuery();
  if (!q) return;
  if (controller) controller.abort();
  const mine = new AbortController();
  controller = mine;
  searching.value = true;
  results.value = [];
  const base = `${sourceStore.filesBase}/`;
  try {
    await search(base, q, mine.signal, (item) => {
      if (controller === mine) {
        results.value.push(item as unknown as SearchResult);
      }
    });
  } catch {
    // aborted or failed search — leave whatever streamed in
  } finally {
    if (controller === mine) searching.value = false;
  }
};

// Re-run whenever the typed query or the active type filter changes.
watch([query, activeType], () => {
  activeIndex.value = 0;
  if (debounce) clearTimeout(debounce);
  if (controller) controller.abort();
  results.value = [];
  searching.value = false;
  if (!inSearchMode.value) return;
  debounce = window.setTimeout(runSearch, 180);
});

const toggleType = (key: string) => {
  activeType.value = activeType.value === key ? null : key;
  inputEl.value?.focus();
};

const scrollActiveIntoView = () => {
  nextTick(() => {
    listEl.value
      ?.querySelector(`[data-index="${activeIndex.value}"]`)
      ?.scrollIntoView({ block: "nearest" });
  });
};

const move = (delta: number) => {
  const n = flat.value.length;
  if (!n) return;
  activeIndex.value = (activeIndex.value + delta + n) % n;
  scrollActiveIntoView();
};

const execActive = () => {
  flat.value[activeIndex.value]?.run();
};

const close = () => {
  layoutStore.closeCommandPalette();
};

const onKeydown = (e: KeyboardEvent) => {
  // Toggle shortcut while open closes it; keep it from reaching the global
  // window handler (which would immediately re-open it).
  if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
    e.preventDefault();
    e.stopPropagation();
    close();
    return;
  }
  // Stop nav keys from bubbling to window — the Preview view listens there for
  // arrows/Enter/Esc and would drive prev/next behind the palette.
  switch (e.key) {
    case "Escape":
      e.preventDefault();
      e.stopPropagation();
      close();
      break;
    case "ArrowDown":
      e.preventDefault();
      e.stopPropagation();
      move(1);
      break;
    case "ArrowUp":
      e.preventDefault();
      e.stopPropagation();
      move(-1);
      break;
    case "Enter":
      e.preventDefault();
      e.stopPropagation();
      execActive();
      break;
    case "Tab":
      // Focus trap: keep focus on the input so Tab can't walk out of the
      // palette into the view behind it (spec requirement).
      e.preventDefault();
      e.stopPropagation();
      inputEl.value?.focus();
      break;
  }
};

watch(open, (isOpen) => {
  if (isOpen) {
    opener = document.activeElement as HTMLElement | null;
    query.value = "";
    activeType.value = null;
    results.value = [];
    activeIndex.value = 0;
    searching.value = false;
    nextTick(() => inputEl.value?.focus());
  } else {
    if (controller) controller.abort();
    if (debounce) clearTimeout(debounce);
    // Restore focus to whatever opened the palette (accessibility).
    opener?.focus?.();
    opener = null;
  }
});
</script>

<style scoped>
.fb-cmdk-backdrop {
  position: fixed;
  inset: 0;
  z-index: 20000;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding-top: 12vh;
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(2px);
}

.fb-cmdk {
  width: min(640px, calc(100vw - 32px));
  max-height: 70vh;
  display: flex;
  flex-direction: column;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 14px;
  box-shadow: 0 24px 60px rgba(0, 0, 0, 0.35);
  overflow: hidden;
}

.fb-cmdk-input-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 16px;
  border-bottom: 1px solid var(--border);
}

.fb-cmdk-search-icon {
  color: var(--dim);
  flex: 0 0 auto;
}

.fb-cmdk-input {
  flex: 1 1 auto;
  border: none;
  background: transparent;
  outline: none;
  font-size: 1.05rem;
  color: var(--text);
}

.fb-cmdk-input::placeholder {
  color: var(--faint);
}

.fb-cmdk-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
  padding: 10px 14px;
  border-bottom: 1px solid var(--border);
}

.fb-cmdk-filter {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  height: 28px;
  padding: 0 10px;
  border: 1px solid var(--border);
  border-radius: 999px;
  background: transparent;
  color: var(--dim);
  font-size: 12.5px;
  font-weight: 500;
  font-family: inherit;
  cursor: pointer;
  transition:
    background 0.1s,
    color 0.1s,
    border-color 0.1s;
}

.fb-cmdk-filter:hover {
  background: var(--hover);
  color: var(--text);
}

.fb-cmdk-filter.active {
  background: var(--accent);
  border-color: var(--accent);
  color: var(--on-accent, #fff);
}

.fb-cmdk-scope {
  flex: 0 0 auto;
  padding: 3px 9px;
  border-radius: 6px;
  background: var(--hover);
  color: var(--dim);
  font-size: 0.72rem;
  font-weight: 600;
  white-space: nowrap;
}

.fb-cmdk-list {
  flex: 1 1 auto;
  overflow-y: auto;
  padding: 6px;
}

.fb-cmdk-section {
  padding: 10px 12px 4px;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--faint);
}

.fb-cmdk-row {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 9px 12px;
  border: none;
  border-radius: 9px;
  background: transparent;
  color: var(--text);
  cursor: pointer;
  text-align: left;
}

.fb-cmdk-row.active {
  background: var(--hover);
}

.fb-cmdk-row-icon {
  color: var(--dim);
  flex: 0 0 auto;
}

.fb-cmdk-row-label {
  flex: 0 1 auto;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.fb-cmdk-row-label :deep(mark) {
  background: transparent;
  color: var(--accent);
  font-weight: 700;
}

.fb-cmdk-row-sub {
  flex: 1 1 auto;
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  text-align: right;
  font-size: 0.8rem;
  color: var(--faint);
}

.fb-cmdk-status {
  padding: 28px 12px;
  text-align: center;
  color: var(--faint);
}

.fb-cmdk-footer {
  display: flex;
  gap: 16px;
  padding: 9px 16px;
  border-top: 1px solid var(--border);
  font-size: 0.72rem;
  color: var(--faint);
}

.fb-cmdk-footer kbd {
  display: inline-block;
  min-width: 16px;
  padding: 1px 5px;
  margin-right: 2px;
  border: 1px solid var(--border);
  border-radius: 4px;
  background: var(--hover);
  font-family: inherit;
  font-size: 0.7rem;
  color: var(--dim);
}

@media (max-width: 736px) {
  .fb-cmdk-backdrop {
    padding: 0;
  }
  .fb-cmdk {
    width: 100vw;
    max-height: 100vh;
    height: 100%;
    border: none;
    border-radius: 0;
  }
}
</style>
