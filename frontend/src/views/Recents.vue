<template>
  <div class="fb-files-root">
    <header-bar>
      <span class="fb-toolbar-title">{{ $t("sidebar.recent") }}</span>
      <template #actions>
        <!-- Local search -->
        <div class="fb-local-search">
          <FbIcon name="search" size="16px" />
          <input
            v-model.trim="searchQuery"
            :placeholder="$t('buttons.search')"
            type="text"
          />
          <button v-if="searchQuery" class="fb-local-search-clear" @click="searchQuery = ''">
            <FbIcon name="x" size="14px" />
          </button>
        </div>

        <!-- View toggle -->
        <div class="fb-view-toggle" :title="t('buttons.switchView')">
          <button
            class="fb-tbtn"
            :class="{ 'fb-tbtn--active': currentViewMode === 'list' }"
            :aria-label="t('buttons.listView', 'List view')"
            @click="setView('list')"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="width: 17px; height: 17px" aria-hidden="true">
              <path d="M8 6h13" /><path d="M8 12h13" /><path d="M8 18h13" />
              <path d="M3 6h.01" /><path d="M3 12h.01" /><path d="M3 18h.01" />
            </svg>
          </button>
          <button
            class="fb-tbtn"
            :class="{ 'fb-tbtn--active': currentViewMode === 'mosaic' }"
            :aria-label="t('buttons.gridView', 'Grid view')"
            @click="setView('mosaic')"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="width: 16px; height: 16px" aria-hidden="true">
              <path d="M3 3h7v7H3z" /><path d="M14 3h7v7h-7z" />
              <path d="M14 14h7v7h-7z" /><path d="M3 14h7v7H3z" />
            </svg>
          </button>
        </div>

        <!-- Sort dropdown -->
        <div style="position: relative">
          <button
            class="fb-tbtn fb-tbtn--bordered"
            :title="t('files.sort', 'Sort')"
            @click="sortMenuOpen = !sortMenuOpen"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="width: 17px; height: 17px" aria-hidden="true">
              <path d="M7 4v16" /><path d="M3 8l4-4 4 4" />
              <path d="M17 20V4" /><path d="M21 16l-4 4-4-4" />
            </svg>
          </button>
          <div v-show="sortMenuOpen" class="fb-sort-menu" @click.self="sortMenuOpen = false">
            <button
              class="fb-sort-item"
              :class="{ 'fb-sort-item--active': sortBy === 'name' }"
              @click="handleSortMenuClick('name')"
            >
              <span>{{ t("files.name") }}</span>
              <span v-if="sortBy === 'name'" class="fb-sort-arrow">{{ sortAsc ? "↑" : "↓" }}</span>
            </button>
            <button
              class="fb-sort-item"
              :class="{ 'fb-sort-item--active': sortBy === 'modified' }"
              @click="handleSortMenuClick('modified')"
            >
              <span>{{ t("files.lastModified") }}</span>
              <span v-if="sortBy === 'modified'" class="fb-sort-arrow">{{ sortAsc ? "↑" : "↓" }}</span>
            </button>
          </div>
        </div>

        <!-- Multi-select toggle -->
        <button
          class="fb-tbtn"
          :class="{ 'fb-tbtn--active': fileStore.multiple }"
          :aria-label="t('buttons.selectMultiple')"
          :title="t('buttons.selectMultiple')"
          @click="fileStore.toggleMultiple()"
        >
          <FbIcon name="select-all" size="18px" />
        </button>

        <!-- Details panel toggle -->
        <button
          class="fb-tbtn fb-tbtn--bordered"
          :class="{ 'fb-tbtn--active': layoutStore.showDetails }"
          :title="t('buttons.info')"
          @click="layoutStore.toggleDetails()"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round" style="width: 17px; height: 17px" aria-hidden="true">
            <path d="M3 5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
            <path d="M15 3v18" />
          </svg>
        </button>
      </template>
    </header-bar>

    <SelectionBar
      v-if="fileStore.multiple && fileStore.selectedCount > 0"
      :header-buttons="{ download: true, star: true }"
      :download="downloadSelected"
    />

    <div class="fb-content-row">
      <div class="fb-content-main">
        <div v-if="allItems.length === 0" class="fb-empty">
      <div class="fb-empty-icon">
        <fb-icon name="clock" size="40px" />
      </div>
      <h2 class="fb-empty-title">{{ $t("files.emptyRecentTitle") }}</h2>
      <p class="fb-empty-sub">{{ $t("files.emptyRecentSub") }}</p>
    </div>

    <div v-else-if="filteredItems.length === 0" class="fb-empty">
      <div class="fb-empty-icon">
        <fb-icon name="search" size="40px" />
      </div>
      <p class="fb-empty-sub">{{ $t("files.noResults", "No results found") }}</p>
    </div>

    <div
      v-else
      id="listing"
      class="file-icons no-size-col"
      :class="currentViewMode"
      data-clear-on-click="true"
      @click="handleEmptyAreaClick"
    >
      <div>
        <div class="fb-col-header">
          <div>
            <p class="name">{{ $t("files.name") }}</p>
            <p class="modified">{{ $t("files.lastModified") }}</p>
          </div>
        </div>
      </div>

      <h2 v-if="dirs.length > 0" data-clear-on-click="true">{{ $t("files.folders") }}</h2>
      <div v-if="dirs.length > 0" class="fb-items fb-items--folders" data-clear-on-click="true">
        <item
          v-for="(it, i) in dirs"
          :key="it.url"
          :index="i"
          :name="it.name"
          :isDir="true"
          :url="it.url"
          :type="it.type"
          :size="0"
          :modified="toISO(it.at)"
          :path="toPath(it.url)"
          hideSize
        />
      </div>

      <h2 v-if="files.length > 0" data-clear-on-click="true">{{ $t("files.files") }}</h2>
      <div v-if="files.length > 0" class="fb-items fb-items--files" data-clear-on-click="true">
        <item
          v-for="(it, i) in files"
          :key="it.url"
          :index="dirs.length + i"
          :name="it.name"
          :isDir="false"
          :url="it.url"
          :type="it.type"
          :size="0"
          :modified="toISO(it.at)"
          :path="toPath(it.url)"
          hideSize
        />
      </div>
    </div>

      </div>
      <DetailsPanel v-if="layoutStore.showDetails" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { getRecents, type RecentFile } from "@/utils/recents";
import { removePrefix } from "@/api/utils";
import { users, files as api } from "@/api";
import HeaderBar from "@/components/header/HeaderBar.vue";
import FbIcon from "@/components/FbIcon.vue";
import Item from "@/components/files/ListingItem.vue";
import SelectionBar from "@/components/SelectionBar.vue";
import DetailsPanel from "@/components/DetailsPanel.vue";

const { t } = useI18n();
const authStore = useAuthStore();
const fileStore = useFileStore();
const layoutStore = useLayoutStore();

const allItems = ref<RecentFile[]>([]);
const searchQuery = ref("");
const sortMenuOpen = ref(false);
const sortBy = ref("modified");
const sortAsc = ref(false);

onMounted(async () => {
  allItems.value = getRecents();
  try {
    const fresh = await users.get(authStore.user!.id);
    if (Array.isArray(fresh.recents)) {
      authStore.updateUser({ recents: fresh.recents });
      localStorage.setItem("fb-recent-files", JSON.stringify(fresh.recents));
      allItems.value = fresh.recents as RecentFile[];
    }
  } catch {
    // backend unavailable — keep localStorage data
  }
});

onUnmounted(() => {
  fileStore.updateRequest(null);
  fileStore.multiple = false;
});

const currentViewMode = computed(() => {
  const m = authStore.user?.viewMode ?? "list";
  return m === "mosaic gallery" ? "mosaic" : m;
});

const filteredItems = computed(() => {
  if (!searchQuery.value) return allItems.value;
  const q = searchQuery.value.toLowerCase();
  return allItems.value.filter((it) => it.name.toLowerCase().includes(q));
});

const sortedItems = computed(() => {
  const list = [...filteredItems.value];
  const dir = sortAsc.value ? 1 : -1;
  return list.sort((a, b) => {
    if (sortBy.value === "name") {
      return a.name.localeCompare(b.name) * dir;
    }
    return (a.at - b.at) * dir;
  });
});

const dirs = computed(() => sortedItems.value.filter((i) => i.type === "dir"));
const files = computed(() => sortedItems.value.filter((i) => i.type !== "dir"));

const displayItems = computed(() => [...dirs.value, ...files.value]);

watch(displayItems, () => syncFileStore(), { immediate: false });

function syncFileStore() {
  const items = displayItems.value.map((it, i) => ({
    index: i,
    name: it.name,
    isDir: it.type === "dir",
    url: it.url,
    type: it.type,
    size: 0,
    modified: toISO(it.at),
    path: toPath(it.url),
    extension: "",
    mode: 0,
    isSymlink: false,
  }));

  fileStore.updateRequest({
    path: "/",
    name: "Recent",
    size: 0,
    extension: "",
    modified: new Date().toISOString(),
    mode: 0,
    isDir: true,
    isSymlink: false,
    type: "dir",
    url: "/recent",
    items,
    numDirs: dirs.value.length,
    numFiles: files.value.length,
    sorting: { by: sortBy.value, asc: sortAsc.value },
  } as Resource);
}

const setView = (mode: "list" | "mosaic" | "mosaic gallery") => {
  if (authStore.user?.viewMode === mode) return;
  layoutStore.closeHovers();
  if (authStore.user?.id) {
    users.update({ id: authStore.user.id, viewMode: mode }, ["viewMode"]);
    authStore.updateUser({ viewMode: mode });
  }
};

const handleSortMenuClick = (by: string) => {
  if (sortBy.value === by) {
    sortAsc.value = !sortAsc.value;
  } else {
    sortBy.value = by;
    sortAsc.value = by === "name";
  }
  sortMenuOpen.value = false;
};

const downloadSelected = () => {
  const req = fileStore.req;
  if (!req) return;
  const urls: string[] = [];
  for (const idx of fileStore.selected) {
    const item = req.items[idx];
    if (item && !item.isDir) {
      urls.push(item.url);
    }
  }
  if (urls.length === 1) {
    api.download(null, urls[0]);
  } else if (urls.length > 1) {
    layoutStore.showHover({
      prompt: "download",
      confirm: (format: any) => {
        layoutStore.closeHovers();
        api.download(format, ...urls);
      },
    });
  }
};

const handleEmptyAreaClick = (e: MouseEvent) => {
  const target = e.target;
  if (!(target instanceof HTMLElement)) return;
  if (target.dataset.clearOnClick === "true") {
    fileStore.selected = [];
  }
};

const toISO = (ms: number) => new Date(ms).toISOString();
const toPath = (url: string) => removePrefix(decodeURIComponent(url));
</script>

<style scoped>
#listing.no-size-col.list .item > div:last-of-type,
#listing.no-size-col.list .fb-col-header > div {
  grid-template-columns: 1fr 168px;
}
#listing.no-size-col.list .size {
  display: none;
}
</style>
