<template>
  <div
    ref="wrapperEl"
    class="fb-search-wrap"
    :class="{ 'fb-search-wrap--open': isMobileOpen }"
    @keydown.esc.stop="close"
    @focusout="onFocusout"
  >
    <FbIcon name="search" size="16px" class="fb-search-icon" />
    <input
      ref="inputEl"
      class="fb-search"
      type="text"
      v-model.trim="prompt"
      :placeholder="$t('search.search')"
      :aria-label="$t('search.search')"
      @focus="onFocus"
      @keydown.enter="onEnter"
      @keydown.down.prevent="moveActive(1)"
      @keydown.up.prevent="moveActive(-1)"
    />
    <button
      v-if="ongoing"
      class="fb-search-end"
      @mousedown.prevent="stopSearch"
      :title="$t('buttons.stopSearch')"
      :aria-label="$t('buttons.stopSearch')"
    >
      <FbIcon name="stop-circle" size="15px" />
    </button>
    <button
      v-else-if="prompt"
      class="fb-search-end"
      @mousedown.prevent="clearSearch"
    >
      <FbIcon name="x" size="15px" />
    </button>
    <button
      v-else
      class="fb-search-kbd"
      type="button"
      @mousedown.prevent="openPalette"
      :title="$t('commandPalette.placeholder')"
      :aria-label="$t('commandPalette.placeholder')"
    >
      {{ shortcutLabel }}
    </button>

    <div
      v-show="isOpen"
      ref="dropdownEl"
      class="fb-search-dropdown"
      @mousedown.prevent
    >
      <div v-if="isEmpty">
        <p class="fb-search-hint">{{ text }}</p>
        <template v-if="prompt === ''">
          <div class="fb-search-types">
            <h3>{{ $t("search.types") }}</h3>
            <div class="fb-search-type-grid">
              <button
                v-for="(v, k) in boxes"
                :key="k"
                class="fb-search-type-btn"
                @click="init('type:' + k)"
              >
                <FbIcon :name="v.icon" size="16px" />
                <span>{{ $t("search." + v.label) }}</span>
              </button>
            </div>
          </div>
        </template>
      </div>
      <div v-if="searchSections.length > 0" class="fb-search-results">
        <template v-for="section in searchSections" :key="section.id">
          <h3 class="fb-search-section">{{ sectionHeader(section.id) }}</h3>
          <ul>
            <li v-for="entry in section.entries" :key="entry.key">
              <router-link
                @click="close"
                @mousemove="hoverEntry(entry.index)"
                :to="entry.url"
                :class="{ active: entry.index === activeIndex }"
                :data-index="entry.index"
              >
                <FbIcon :name="entry.icon" size="15px" />
                <span
                  class="fb-search-name"
                  v-html="highlightMatch(entry.name, highlightTerm)"
                ></span>
                <span v-if="entry.dir" class="fb-search-sub">{{
                  entry.dir
                }}</span>
              </router-link>
            </li>
          </ul>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import FbIcon from "@/components/FbIcon.vue";
import type { IconName } from "@/utils/icons";
import {
  buildSearchSections,
  highlightMatch,
  highlightTermFromPrompt,
} from "@/utils/searchResults";
import url from "@/utils/url";
import { search } from "@/api";
import {
  computed,
  inject,
  nextTick,
  onMounted,
  onUnmounted,
  ref,
  watch,
} from "vue";
import { useI18n } from "vue-i18n";
import { useRoute, useRouter } from "vue-router";
import { StatusError } from "@/api/utils";

const boxes: Record<string, { label: string; icon: IconName }> = {
  image: { label: "images", icon: "image" },
  audio: { label: "music", icon: "audio" },
  video: { label: "video", icon: "video" },
  pdf: { label: "pdf", icon: "pdf" },
};

const fileStore = useFileStore();
const layoutStore = useLayoutStore();
let searchAbortController = new AbortController();

// OS-aware shortcut hint shown in the search bar (⌘K on macOS, Ctrl K else),
// mirroring the convention used by most modern apps. Clicking it opens the
// command palette (#33).
const isMac =
  typeof navigator !== "undefined" &&
  /Mac|iPhone|iPad|iPod/.test(navigator.platform || navigator.userAgent || "");
const shortcutLabel = isMac ? "⌘K" : "Ctrl K";

const openPalette = () => {
  layoutStore.toggleCommandPalette();
};

const prompt = ref<string>("");
const ongoing = ref<boolean>(false);
const results = ref<any[]>([]);
const resultsCount = ref<number>(50);
const isOpen = ref<boolean>(false);
const isMobileOpen = ref<boolean>(false);
// Keyboard navigation over the sectioned results (Ctrl+K parity).
const activeIndex = ref<number>(-1);
const kbNav = ref<boolean>(false);
let liveDebounce: number | null = null;

const scheduleLiveSearch = () => {
  kbNav.value = false;
  activeIndex.value = -1;
  if (liveDebounce) clearTimeout(liveDebounce);
  if (prompt.value === "") return;
  // Same debounce as the command palette: search while typing.
  liveDebounce = window.setTimeout(() => void runSearch(), 180);
};

const $showError = inject<IToastError>("$showError")!;

const inputEl = ref<HTMLInputElement | null>(null);
const dropdownEl = ref<HTMLElement | null>(null);
const wrapperEl = ref<HTMLElement | null>(null);

const { t } = useI18n();
const route = useRoute();
const router = useRouter();

watch(prompt, () => {
  reset();
  scheduleLiveSearch();
});

const isEmpty = computed(() => results.value.length === 0);

const text = computed(() => {
  if (ongoing.value) return "";
  return prompt.value === ""
    ? t("search.typeToSearch")
    : t("search.pressToSearch");
});

const filteredResults = computed(() =>
  results.value.slice(0, resultsCount.value)
);

/** Folders/Files sections like the Ctrl+K palette, with flat keyboard index. */
const searchSections = computed(() => {
  const sections = buildSearchSections(filteredResults.value);
  let i = 0;
  return sections.map((s) => ({
    ...s,
    entries: s.entries.map((e) => ({ ...e, index: i++ })),
  }));
});

const flatSearchEntries = computed(() =>
  searchSections.value.flatMap((s) => s.entries)
);

const highlightTerm = computed(() => highlightTermFromPrompt(prompt.value));

const sectionHeader = (id: string) =>
  id === "folders" ? t("files.folders") : t("files.files");

onMounted(() => {
  if (!dropdownEl.value) return;
  dropdownEl.value.addEventListener("scroll", (event: Event) => {
    const el = event.target as HTMLElement;
    if (el.offsetHeight + el.scrollTop >= el.scrollHeight - 100) {
      resultsCount.value += 50;
    }
  });
});

onUnmounted(() => {
  abortLastSearch();
  if (liveDebounce) clearTimeout(liveDebounce);
});

const onFocus = () => {
  isOpen.value = true;
};

const onFocusout = (e: FocusEvent) => {
  if (!wrapperEl.value?.contains(e.relatedTarget as Node)) {
    isOpen.value = false;
    isMobileOpen.value = false;
  }
};

const close = () => {
  isOpen.value = false;
  isMobileOpen.value = false;
  prompt.value = "";
  reset();
  inputEl.value?.blur();
};

const stopSearch = () => {
  abortLastSearch();
  ongoing.value = false;
};

const clearSearch = () => {
  prompt.value = "";
  reset();
  inputEl.value?.focus();
};

const init = (string: string) => {
  prompt.value = `${string} `;
  inputEl.value?.focus();
};

const reset = () => {
  abortLastSearch();
  if (liveDebounce) clearTimeout(liveDebounce);
  ongoing.value = false;
  resultsCount.value = 50;
  results.value = [];
};

const abortLastSearch = () => {
  searchAbortController.abort();
};

const runSearch = async () => {
  if (prompt.value === "") return;

  let path = route.path;
  if (!fileStore.isListing) {
    path = url.removeLastDir(path) + "/";
  }

  ongoing.value = true;

  try {
    abortLastSearch();
    searchAbortController = new AbortController();
    results.value = [];
    await search(path, prompt.value, searchAbortController.signal, (item) =>
      results.value.push(item)
    );
  } catch (error: any) {
    if (error instanceof StatusError && error.is_canceled) {
      return;
    }
    $showError(error);
  }

  ongoing.value = false;
};

const submit = (event: Event) => {
  event.preventDefault();
  if (liveDebounce) clearTimeout(liveDebounce);
  void runSearch();
};

const openEntry = (entryUrl: string) => {
  close();
  router.push(entryUrl);
};

/** Enter opens the keyboard/hover-highlighted row, else runs the search. */
const onEnter = (event: Event) => {
  const entry =
    kbNav.value && activeIndex.value >= 0
      ? flatSearchEntries.value[activeIndex.value]
      : undefined;
  if (entry) {
    event.preventDefault();
    openEntry(entry.url);
    return;
  }
  submit(event);
};

const moveActive = (delta: number) => {
  const n = flatSearchEntries.value.length;
  if (!n) return;
  kbNav.value = true;
  const start =
    activeIndex.value < 0 ? (delta > 0 ? -1 : 0) : activeIndex.value;
  activeIndex.value = (start + delta + n) % n;
  dropdownEl.value
    ?.querySelector(`[data-index="${activeIndex.value}"]`)
    ?.scrollIntoView({ block: "nearest" });
};

const hoverEntry = (index: number) => {
  activeIndex.value = index;
  kbNav.value = true;
};

const focus = () => {
  isMobileOpen.value = true;
  nextTick(() => inputEl.value?.focus());
};

defineExpose({ focus });
</script>
