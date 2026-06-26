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
    <button
      v-if="headerButtons.download"
      class="fb-act"
      @click="download"
    >
      <FbIcon name="download" size="16px" />
      <span>{{ t('buttons.download') }}</span>
    </button>
    <button
      v-if="headerButtons.share"
      class="fb-act"
      @click="layoutStore.showHover('share')"
    >
      <FbIcon name="share" size="16px" />
      <span>{{ t('buttons.share') }}</span>
    </button>
    <button
      v-if="headerButtons.move"
      class="fb-act"
      @click="layoutStore.showHover('move')"
    >
      <FbIcon name="move" size="16px" />
      <span>{{ t('buttons.moveFile') }}</span>
    </button>
    <button
      v-if="headerButtons.rename"
      class="fb-act"
      @click="layoutStore.showHover('rename')"
    >
      <FbIcon name="rename" size="16px" />
      <span>{{ t('buttons.rename') }}</span>
    </button>
    <button
      class="fb-act"
      aria-disabled="true"
      style="opacity:0.38;pointer-events:none"
    >
      <FbIcon name="star" size="16px" />
      <span>{{ t('buttons.star', 'Star') }}</span>
    </button>
    <div class="fb-sel-spacer" />
    <button
      v-if="headerButtons.delete"
      class="fb-act fb-act--danger"
      data-danger="true"
      @click="layoutStore.showHover('delete')"
    >
      <FbIcon name="delete" size="16px" />
      <span>{{ t('buttons.delete') }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import FbIcon from "@/components/FbIcon.vue";

const props = defineProps<{
  headerButtons: {
    download?: boolean;
    share?: boolean;
    move?: boolean;
    rename?: boolean;
    delete?: boolean;
  };
  download: () => void;
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
</script>
