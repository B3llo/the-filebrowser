<template>
  <div class="fb-selection-bar">
    <button
      class="fb-act fb-sel-clear"
      :title="t('buttons.clear')"
      :aria-label="t('buttons.clear')"
      @click="clearSelection"
    >
      <FbIcon name="x" size="17px" />
    </button>
    <span class="fb-sel-label">
      {{ selectionLabel }}
    </span>
    <button v-if="headerButtons.download" class="fb-act" @click="download">
      <FbIcon name="download" size="16px" />
      <span>{{ t("buttons.download") }}</span>
    </button>
    <button
      v-if="headerButtons.share"
      class="fb-act"
      @click="layoutStore.showHover('share')"
    >
      <FbIcon name="share" size="16px" />
      <span>{{ t("buttons.share") }}</span>
    </button>
    <button
      v-if="headerButtons.move"
      class="fb-act"
      @click="layoutStore.showHover('move')"
    >
      <FbIcon name="move" size="16px" />
      <span>{{ t("buttons.moveFile") }}</span>
    </button>
    <button
      v-if="headerButtons.rename"
      class="fb-act"
      @click="layoutStore.showHover('rename')"
    >
      <FbIcon name="rename" size="16px" />
      <span>{{ t("buttons.rename") }}</span>
    </button>
    <button
      v-if="headerButtons.star"
      class="fb-act"
      @click="handleStar"
    >
      <FbIcon name="star" size="16px" />
      <span>{{ allSelectedStarred ? t("files.unstar", "Unstar") : t("buttons.star", "Star") }}</span>
    </button>
    <button
      v-if="headerButtons.restore"
      class="fb-act"
      @click="onRestore"
    >
      <FbIcon name="arrow-back" size="16px" />
      <span>{{ t("trash.restore") }}</span>
    </button>
    <div class="fb-sel-spacer" />
    <button
      v-if="headerButtons.delete"
      class="fb-act fb-act--danger"
      data-danger="true"
      @click="onDelete"
    >
      <FbIcon name="delete" size="16px" />
      <span>{{ t("buttons.delete") }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { toggleStarred, isStarred, starVersion } from "@/utils/starred";
import FbIcon from "@/components/FbIcon.vue";

const props = defineProps<{
  headerButtons: {
    download?: boolean;
    share?: boolean;
    move?: boolean;
    rename?: boolean;
    delete?: boolean;
    star?: boolean;
    restore?: boolean;
  };
  download: () => void;
  deleteAction?: () => void;
  restoreAction?: () => void;
}>();

const { t } = useI18n();
const fileStore = useFileStore();
const layoutStore = useLayoutStore();

const selectionLabel = computed(() => {
  const count = fileStore.selectedCount;
  return count === 1
    ? t("prompts.filesSelected", 1)
    : t("prompts.filesSelected", count);
});

const clearSelection = () => {
  fileStore.selected = [];
};

const allSelectedStarred = computed(() => {
  starVersion.value; // reactive dependency
  const items = fileStore.req?.items;
  if (!items || fileStore.selected.length === 0) return false;
  return fileStore.selected.every((idx) => {
    const item = items[idx];
    return item ? isStarred(item.url) : false;
  });
});

const handleStar = () => {
  const items = fileStore.req?.items;
  if (!items || fileStore.selected.length === 0) return;
  for (const idx of fileStore.selected) {
    const item = items[idx];
    if (item) {
      toggleStarred({ url: item.url, name: item.name, type: item.type });
    }
  }
};

const onDelete = () => {
  if (props.deleteAction) {
    props.deleteAction();
  } else {
    layoutStore.showHover("delete");
  }
};

const onRestore = () => {
  if (props.restoreAction) {
    props.restoreAction();
  }
};
</script>
