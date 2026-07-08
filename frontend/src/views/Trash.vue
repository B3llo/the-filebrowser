<template>
  <div class="fb-files-root">
    <div v-if="loading" class="loading">
      <div class="spinner">
        <div class="bounce1"></div>
        <div class="bounce2"></div>
        <div class="bounce3"></div>
      </div>
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
            @update:search-query="searchQuery = $event"
            @update:sort-by="sortBy = $event"
            @update:sort-asc="sortAsc = $event"
            @restore="restoreAllSelected"
            @delete-permanent="deletePermanently"
            @empty-trash="confirmEmptyTrash"
            @switch-source="switchSource"
            @navigate-to="navigateTo"
          />
        </div>
        <DetailsPanel v-if="layoutStore.showDetails" />
      </div>
    </template>

    <ConfirmDialog
      v-if="showDeleteDialog"
      :message="t('prompts.deleteMessageMultiple', { count: pendingDeleteItems.length })"
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
import { files as api } from "@/api";
import { useSourceStore } from "@/stores/source";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { StatusError } from "@/api/utils";
import FileListing from "@/views/files/FileListing.vue";
import DetailsPanel from "@/components/DetailsPanel.vue";
import ConfirmDialog from "@/components/ConfirmDialog.vue";

const $showError = inject<IToastError>("$showError")!;
const { t } = useI18n();
const sourceStore = useSourceStore();
const fileStore = useFileStore();
const layoutStore = useLayoutStore();

const loading = ref(false);
const activeSourceId = ref(String(sourceStore.activeId));
const trashSubPath = ref("");

const showDeleteDialog = ref(false);
const showEmptyDialog = ref(false);
const pendingDeleteItems = ref<any[]>([]);

const searchQuery = ref("");
const sortBy = ref("name");
const sortAsc = ref(true);

const filesBase = computed(() => `/files/${activeSourceId.value}`);
const trashPath = computed(() => `${filesBase.value}/.Trash/`);

const breadcrumbSegments = computed(() => {
  const segs: { label: string; url: string }[] = [
    { label: t("sidebar.trash"), url: trashPath.value },
  ];
  if (!trashSubPath.value) return segs;
  const parts = trashSubPath.value.replace(/\/$/, "").split("/");
  let accumulated = "";
  for (const part of parts) {
    accumulated += part + "/";
    segs.push({ label: cleanName(part), url: trashPath.value + accumulated });
  }
  return segs;
});

onMounted(() => loadTrash());
onUnmounted(() => { fileStore.updateRequest(null); });
watch(activeSourceId, () => {
  trashSubPath.value = "";
  loadTrash();
});

const switchSource = (id: string) => {
  activeSourceId.value = id;
};

const loadTrash = async () => {
  loading.value = true;
  fileStore.updateRequest(null);
  try {
    const res = await api.fetch(trashPath.value + trashSubPath.value);
    fileStore.updateRequest(res);
  } catch (e: any) {
    if (e instanceof StatusError && e.status === 404) {
      fileStore.updateRequest(null);
    } else {
      $showError(e);
    }
  } finally {
    loading.value = false;
  }
};

const cleanName = (trashName: string): string => trashName.replace(/^\d+_/, "");

const navigateTo = (url: string) => {
  const base = trashPath.value;
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
  try {
    await Promise.all(
      selectedItems.map((item) => {
        const destName = cleanName(item.name);
        const destPath = `${filesBase.value}/${destName}${item.isDir ? "/" : ""}`;
        return api.move([{ from: item.url, to: destPath }], false, true);
      })
    );
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

const executeEmptyTrash = async () => {
  showEmptyDialog.value = false;
  const items = fileStore.req?.items ?? [];
  try {
    await Promise.all(items.map((item) => api.remove(item.url)));
    fileStore.updateRequest(null);
  } catch (e: any) {
    $showError(e);
  }
};

const cancelEmptyTrash = () => {
  showEmptyDialog.value = false;
};
</script>
