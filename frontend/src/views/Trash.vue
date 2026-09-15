<template>
  <div class="fb-files-root">
    <div v-if="loading" class="loading">
      <div class="spinner">
        <div class="bounce1"></div>
        <div class="bounce2"></div>
        <div class="bounce3"></div>
      </div>
    </div>
    <div v-else-if="trashDisabled" class="fb-trash-disabled">
      <p>{{ t("trash.disabled") }}</p>
    </div>
    <template v-else>
      <div class="fb-content-row">
        <div class="fb-content-main">
          <FileListing
            isTrash
            :source-id="activeSourceId"
            :search-query="searchQuery"
            :sort-by="sortBy"
            :sort-asc="sortAsc"
            :trash-sub-path="trashSubPath"
            :breadcrumb="breadcrumbSegments"
            :flat-view="flatView"
            @update:search-query="searchQuery = $event"
            @update:sort-by="sortBy = $event"
            @update:sort-asc="sortAsc = $event"
            @restore="restoreAllSelected"
            @delete-permanent="deletePermanently"
            @empty-trash="confirmEmptyTrash"
            @switch-source="switchSource"
            @navigate-to="navigateTo"
            @toggle-flat="setFlatView"
          />
        </div>
        <DetailsPanel v-if="layoutStore.showDetails" />
      </div>
    </template>

    <ConfirmDialog
      v-if="showDeleteDialog"
      :message="
        t('prompts.deleteMessageMultiple', { count: pendingDeleteItems.length })
      "
      :confirm-text="t('buttons.delete')"
      :cancel-text="t('buttons.cancel')"
      :danger="true"
      @confirm="confirmDelete"
      @cancel="cancelDelete"
    />

    <ConfirmDialog
      v-if="showEmptyDialog"
      :message="t('trash.confirmEmpty')"
      :confirm-text="t('trash.emptyTrash')"
      :cancel-text="t('buttons.cancel')"
      :danger="true"
      @confirm="executeEmptyTrash"
      @cancel="cancelEmptyTrash"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, inject, onMounted, onUnmounted, watch, computed } from "vue";
import { useI18n } from "vue-i18n";
import { files as api, settings as settingsApi } from "@/api";
import { fetchURL, removePrefix, StatusError } from "@/api/utils";
import { useSourceStore } from "@/stores/source";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { trashEnabled as bootTrashEnabled } from "@/utils/constants";
import FileListing from "@/views/files/FileListing.vue";
import DetailsPanel from "@/components/DetailsPanel.vue";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import {
  baseName,
  buildFlatTrashResource,
  cleanTrashName,
  isMirroredTrashPath,
  isTrashEnabled,
  legacyTrashRoot,
  mergeTrashResources,
  originalPathFromTrash,
  parentDir,
  parseTrashInfo,
  shouldShowInTrash,
  trashFilesRoot,
  trashInfoRoot,
  trashInfoUrl,
  withVersionSuffix,
} from "@/utils/trash";

const $showError = inject<IToastError>("$showError")!;
const { t } = useI18n();
const sourceStore = useSourceStore();
const fileStore = useFileStore();
const layoutStore = useLayoutStore();

const loading = ref(false);
// Bootstrap default (public index.html flag); refined via /api/settings
// when the user is allowed to read it (admins).
const trashDisabled = ref(!bootTrashEnabled);
const activeSourceId = ref(String(sourceStore.activeId));
const trashSubPath = ref("");
const flatView = ref(false);

const showDeleteDialog = ref(false);
const showEmptyDialog = ref(false);
const pendingDeleteItems = ref<any[]>([]);

const searchQuery = ref("");
const sortBy = ref("name");
const sortAsc = ref(true);

const filesBase = computed(() => `/files/${activeSourceId.value}`);
/** Mirrored trash tree (new layout). */
const trashFilesPath = computed(() => `${trashFilesRoot(filesBase.value)}/`);
/** Sidecar tree (never listed). */
const trashInfoPath = computed(() => `${trashInfoRoot(filesBase.value)}/`);
/** Legacy flat trash root (pre-mirror items). */
const legacyTrashPath = computed(() => `${legacyTrashRoot(filesBase.value)}/`);

const breadcrumbSegments = computed(() => {
  const segs: { label: string; url: string }[] = [
    { label: t("sidebar.trash"), url: trashFilesPath.value },
  ];
  if (flatView.value) {
    segs.push({ label: t("trash.flatView"), url: trashFilesPath.value });
    return segs;
  }
  if (!trashSubPath.value) return segs;
  const parts = trashSubPath.value.replace(/\/$/, "").split("/");
  let accumulated = "";
  for (const part of parts) {
    accumulated += part + "/";
    segs.push({
      label: cleanTrashName(part),
      url: trashFilesPath.value + accumulated,
    });
  }
  return segs;
});

onMounted(() => initTrash());
onUnmounted(() => {
  fileStore.updateRequest(null);
});
watch(activeSourceId, () => {
  trashSubPath.value = "";
  flatView.value = false;
  loadTrash();
});

const switchSource = (id: string) => {
  // Keep the global source (cookie + api urls) in sync with the trash tabs:
  // each source owns its own .Trash, so the backend Fs must follow the tab.
  sourceStore.setActive(id);
  activeSourceId.value = id;
};

const setFlatView = (value: boolean) => {
  flatView.value = value;
  loadTrash();
};

/** Refresh the enabled flag (admins), then load unless disabled. */
const initTrash = async () => {
  try {
    const s = await settingsApi.get();
    trashDisabled.value = !isTrashEnabled(s.trash);
  } catch {
    // Non-admins cannot read settings: keep the bootstrap value.
  }
  if (!trashDisabled.value) await loadTrash();
};

const loadTrash = async () => {
  if (trashDisabled.value) return;
  loading.value = true;
  fileStore.updateRequest(null);
  try {
    if (flatView.value) {
      await loadFlatTrash();
    } else {
      await loadTreeTrash();
    }
  } catch (e: any) {
    $showError(e);
  } finally {
    loading.value = false;
  }
};

/** Tree view: mirrored dir listing, merged with legacy flat items at the root. */
const loadTreeTrash = async () => {
  let mirrored: Resource | null = null;
  try {
    mirrored = await api.fetch(trashFilesPath.value + trashSubPath.value);
  } catch (e: any) {
    if (!(e instanceof StatusError && e.status === 404)) throw e;
  }
  if (trashSubPath.value !== "") {
    fileStore.updateRequest(mirrored);
    return;
  }
  // Root level: also surface pre-mirror `<epoch>_<name>` items.
  let legacy: Resource | null = null;
  try {
    legacy = await api.fetch(legacyTrashPath.value);
  } catch (e: any) {
    if (!(e instanceof StatusError && e.status === 404)) throw e;
  }
  if (legacy) {
    legacy = {
      ...legacy,
      items: legacy.items.filter(
        (item) =>
          item.name !== "files" &&
          item.name !== "info" &&
          shouldShowInTrash(item.name)
      ),
    };
  }
  fileStore.updateRequest(mergeTrashResources(mirrored, legacy));
};

/** Flat view: every trashed file/folder in one list with its original location. */
const loadFlatTrash = async () => {
  let entries: RecursiveEntry[] = [];
  try {
    entries = await api.fetchAll(trashFilesPath.value);
  } catch (e: any) {
    if (!(e instanceof StatusError && e.status === 404)) throw e;
  }
  try {
    const legacy = await api.fetch(legacyTrashPath.value);
    for (const item of legacy.items ?? []) {
      if (item.name === "files" || item.name === "info") continue;
      if (!shouldShowInTrash(item.name)) continue;
      entries.push({
        path: item.path,
        name: item.name,
        size: item.size,
        modified: item.modified,
        isDir: item.isDir,
      });
    }
  } catch (e: any) {
    if (!(e instanceof StatusError && e.status === 404)) throw e;
  }
  if (entries.length === 0) {
    fileStore.updateRequest(null);
    return;
  }
  fileStore.updateRequest(
    buildFlatTrashResource(entries, filesBase.value, activeSourceId.value)
  );
};

/** Read the sidecar for a mirrored item; null when missing/unreadable. */
const readTrashInfo = async (trashSourcePath: string) => {
  const infoUrl = trashInfoUrl(filesBase.value, trashSourcePath);
  if (infoUrl === null) return null;
  try {
    const res = await fetchURL(
      `/api/raw${removePrefix(infoUrl)}?inline=true`,
      {}
    );
    return parseTrashInfo(await res.text());
  } catch {
    return null;
  }
};

/** Resolve where an item must be restored to (sidecar wins, mirror mirrors, legacy → root). */
const resolveRestorePath = async (item: {
  path: string;
  name: string;
  isDir: boolean;
}): Promise<string> => {
  if (isMirroredTrashPath(item.path)) {
    const info = await readTrashInfo(item.path);
    if (info !== null) return info.originalPath;
    const mirrored = originalPathFromTrash(item.path);
    if (mirrored !== null) return mirrored;
  }
  return `/${cleanTrashName(item.name)}`;
};

const moveWithUniqueName = async (
  fromUrl: string,
  destDirUrl: string,
  name: string,
  isDir: boolean
): Promise<void> => {
  for (let attempt = 0; attempt < 1000; attempt++) {
    const candidate = withVersionSuffix(name, attempt);
    try {
      await api.move([
        { from: fromUrl, to: `${destDirUrl}${candidate}${isDir ? "/" : ""}` },
      ]);
      return;
    } catch (e: any) {
      if (e instanceof StatusError && e.status === 409) continue;
      throw e;
    }
  }
  throw new Error(`could not find a free name for ${name}`);
};

const navigateTo = (url: string) => {
  // Coming from a flat-view folder: go back to tree mode at that folder.
  flatView.value = false;
  const base = trashFilesPath.value;
  if (url.startsWith(base)) {
    const rel = url.slice(base.length);
    trashSubPath.value = rel.startsWith("/") ? rel.slice(1) : rel;
  } else {
    trashSubPath.value = "";
  }
  loadTrash();
};

const restoreAllSelected = async () => {
  const items = fileStore.req?.items ?? [];
  const selectedItems = items.filter((it) =>
    fileStore.selected.includes(it.index)
  );
  if (selectedItems.length === 0) return;
  // Shallowest first; skip children of an already-restored folder (one move
  // restores the whole subtree).
  const ordered = [...selectedItems].sort(
    (a, b) => a.path.split("/").length - b.path.split("/").length
  );
  const restoredRoots: string[] = [];
  try {
    for (const item of ordered) {
      if (restoredRoots.some((root) => item.path.startsWith(`${root}/`)))
        continue;
      const destSourcePath = await resolveRestorePath(item);
      const destParent = parentDir(destSourcePath);
      const destParentUrl =
        destParent === "/"
          ? `${filesBase.value}/`
          : `${filesBase.value}${destParent}/`;
      try {
        await api.post(destParentUrl);
      } catch {
        // Parent may already exist.
      }
      await moveWithUniqueName(
        item.url,
        destParentUrl,
        baseName(destSourcePath),
        item.isDir
      );
      // Sidecar cleanup is best effort — a leftover .trashinfo never surfaces in listings.
      const infoUrl = trashInfoUrl(filesBase.value, item.path);
      if (infoUrl !== null) {
        try {
          await api.remove(infoUrl);
        } catch {
          // Ignore: restore already succeeded.
        }
      }
      if (item.isDir) restoredRoots.push(item.path.replace(/\/$/, ""));
    }
    fileStore.selected = [];
    await loadTrash();
  } catch (e: any) {
    $showError(e);
  }
};

const deletePermanently = () => {
  const items = fileStore.req?.items ?? [];
  pendingDeleteItems.value = items.filter((item) =>
    fileStore.selected.includes(item.index)
  );
  showDeleteDialog.value = true;
};

const confirmDelete = async () => {
  showDeleteDialog.value = false;
  try {
    await Promise.all(
      pendingDeleteItems.value.map((item) => api.remove(item.url))
    );
    fileStore.selected = [];
    pendingDeleteItems.value = [];
    await loadTrash();
  } catch (e: any) {
    $showError(e);
  }
};

const cancelDelete = () => {
  showDeleteDialog.value = false;
  pendingDeleteItems.value = [];
};

const confirmEmptyTrash = () => {
  showEmptyDialog.value = true;
};

/** Remove every top-level child of a trash root (directory removes are recursive). */
const removeAllChildren = async (rootUrl: string) => {
  let listing: Resource | null = null;
  try {
    listing = await api.fetch(rootUrl);
  } catch (e: any) {
    if (e instanceof StatusError && e.status === 404) return;
    throw e;
  }
  await Promise.all((listing?.items ?? []).map((item) => api.remove(item.url)));
};

const executeEmptyTrash = async () => {
  showEmptyDialog.value = false;
  try {
    // Empties the whole source trash regardless of the current subfolder or view.
    await removeAllChildren(trashFilesPath.value);
    await removeAllChildren(trashInfoPath.value);
    try {
      const legacy = await api.fetch(legacyTrashPath.value);
      const leftovers = (legacy.items ?? []).filter(
        (item) => item.name !== "files" && item.name !== "info"
      );
      await Promise.all(leftovers.map((item) => api.remove(item.url)));
    } catch (e: any) {
      if (!(e instanceof StatusError && e.status === 404)) throw e;
    }
    fileStore.selected = [];
    await loadTrash();
  } catch (e: any) {
    $showError(e);
  }
};

const cancelEmptyTrash = () => {
  showEmptyDialog.value = false;
};
</script>
