<template>
  <div>
    <header-bar>
      <span class="fb-toolbar-title">{{ $t("sidebar.trash") }}</span>
      <template #actions>
        <div v-if="sourceStore.sources.length > 1" class="fb-source-tabs">
          <button
            v-for="src in sourceStore.sources"
            :key="src.id"
            class="fb-tbtn"
            :class="{ 'fb-tbtn--active': String(src.id) === activeSourceId }"
            @click="switchSource(String(src.id))"
          >
            {{ src.name }}
          </button>
        </div>
        <button
          v-if="fileStore.req && fileStore.req.items.length > 0"
          class="fb-tbtn fb-tbtn--danger"
          :title="$t('trash.emptyTrash')"
          @click="confirmEmptyTrash"
        >
          {{ $t("trash.emptyTrash") }}
        </button>
      </template>
    </header-bar>

    <SelectionBar
      v-if="fileStore.selected.length > 0"
      :header-buttons="{
        download: false,
        share: false,
        move: false,
        rename: false,
        delete: true,
        star: false,
      }"
      :download="deletePermanently"
      :delete-action="deletePermanently"
    />

    <div v-if="loading" class="loading">
      <div class="spinner">
        <div class="bounce1"></div>
        <div class="bounce2"></div>
        <div class="bounce3"></div>
      </div>
    </div>

    <div
      v-else-if="!fileStore.req || fileStore.req.items.length === 0"
      class="fb-empty"
    >
      <div class="fb-empty-icon">
        <fb-icon name="trash" size="40px" />
      </div>
      <h2 class="fb-empty-title">{{ $t("files.emptyTrashTitle") }}</h2>
      <p class="fb-empty-sub">{{ $t("files.emptyTrashSub") }}</p>
    </div>

    <div
      v-else
      id="listing"
      class="file-icons"
      :class="viewMode"
      @contextmenu="onContextMenu"
    >
      <div>
        <div class="fb-col-header">
          <div>
            <p class="name">{{ $t("files.name") }}</p>
            <p class="size">{{ $t("files.size") }}</p>
            <p class="modified">{{ $t("trash.deletedAt") }}</p>
          </div>
        </div>
      </div>

      <h2 v-if="trashDirs.length > 0">{{ $t("files.folders") }}</h2>
      <div v-if="trashDirs.length > 0" class="fb-items fb-items--folders">
        <item
          v-for="it in trashDirs"
          :key="it.url"
          :index="it.index"
          :name="cleanName(it.name)"
          :isDir="it.isDir"
          :url="it.url"
          :type="it.type"
          :size="it.size"
          :modified="it.modified"
          :path="it.path"
        />
      </div>

      <h2 v-if="trashFiles.length > 0">{{ $t("files.files") }}</h2>
      <div v-if="trashFiles.length > 0" class="fb-items fb-items--files">
        <item
          v-for="it in trashFiles"
          :key="it.url"
          :index="it.index"
          :name="cleanName(it.name)"
          :isDir="it.isDir"
          :url="it.url"
          :type="it.type"
          :size="it.size"
          :modified="it.modified"
          :path="it.path"
          :preview="it.preview"
        />
      </div>
    </div>

    <context-menu :show="ctxVisible" :pos="ctxPos" @hide="ctxVisible = false">
      <action
        :fbIcon="'arrow-back'"
        :label="$t('trash.restore')"
        @action="restoreSelected"
      />
      <hr class="fb-menu-divider" />
      <action
        :fbIcon="'delete'"
        :label="$t('trash.deletePermanent')"
        @action="deleteSelected"
      />
    </context-menu>

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
import { ref, computed, inject, onMounted, onUnmounted, watch } from "vue";
import { useI18n } from "vue-i18n";
import { files as api } from "@/api";
import { useSourceStore } from "@/stores/source";
import { useFileStore } from "@/stores/file";
import { useAuthStore } from "@/stores/auth";
import { StatusError } from "@/api/utils";
import HeaderBar from "@/components/header/HeaderBar.vue";
import FbIcon from "@/components/FbIcon.vue";
import Item from "@/components/files/ListingItem.vue";
import ContextMenu from "@/components/ContextMenu.vue";
import Action from "@/components/header/Action.vue";
import SelectionBar from "@/components/SelectionBar.vue";
import ConfirmDialog from "@/components/ConfirmDialog.vue";

const $showError = inject<IToastError>("$showError")!;
const { t } = useI18n();
const sourceStore = useSourceStore();
const fileStore = useFileStore();
const authStore = useAuthStore();

const loading = ref(false);
const activeSourceId = ref(String(sourceStore.activeId));

const showDeleteDialog = ref(false);
const showEmptyDialog = ref(false);
const pendingDeleteItems = ref<any[]>([]);

const filesBase = computed(() => `/files/${activeSourceId.value}`);
const trashPath = computed(() => `${filesBase.value}/.Trash/`);

const viewMode = computed(() => {
  const m = authStore.user?.viewMode ?? "list";
  return m === "mosaic gallery" ? "mosaic" : m;
});

const trashDirs = computed(() =>
  (fileStore.req?.items ?? []).filter((i) => i.isDir)
);
const trashFiles = computed(() =>
  (fileStore.req?.items ?? []).filter((i) => !i.isDir)
);

onMounted(() => loadTrash());
onUnmounted(() => fileStore.updateRequest(null));

watch(activeSourceId, () => loadTrash());

const switchSource = (id: string) => {
  activeSourceId.value = id;
};

const loadTrash = async () => {
  loading.value = true;
  fileStore.updateRequest(null);
  try {
    const res = await api.fetch(trashPath.value);
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

/** Strip the timestamp prefix that "Move to Trash" added. */
const cleanName = (trashName: string): string => trashName.replace(/^\d+_/, "");

// ---- Context menu ----
const ctxVisible = ref(false);
const ctxPos = ref({ x: 0, y: 0 });

const onContextMenu = (event: MouseEvent) => {
  event.preventDefault();
  const target = event.target as HTMLElement;
  const item = target.closest(".item") as HTMLElement | null;
  if (!item) {
    ctxVisible.value = false;
    return;
  }

  // Select the right-clicked item by finding its index in the listing.
  const items = fileStore.req?.items ?? [];
  const itemUrl = item.getAttribute("data-path")
    ? items.find((it) => it.path === item.getAttribute("data-path"))
    : null;
  if (itemUrl) {
    fileStore.selected = [itemUrl.index];
  }

  ctxPos.value = { x: event.clientX, y: event.clientY };
  ctxVisible.value = true;
};

const selectedItem = computed(() => {
  const items = fileStore.req?.items ?? [];
  if (fileStore.selected.length === 0) return null;
  return items.find((it) => it.index === fileStore.selected[0]) ?? null;
});

const restoreSelected = async () => {
  ctxVisible.value = false;
  const item = selectedItem.value;
  if (!item) return;
  const destName = cleanName(item.name);
  const destPath = `${filesBase.value}/${destName}${item.isDir ? "/" : ""}`;
  try {
    await api.move([{ from: item.url, to: destPath }], false, true);
    await loadTrash();
  } catch (e: any) {
    $showError(e);
  }
};

const deleteSelected = async () => {
  ctxVisible.value = false;
  const item = selectedItem.value;
  if (!item) return;
  try {
    await api.remove(item.url);
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
    await Promise.all(pendingDeleteItems.value.map((item) => api.remove(item.url)));
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
