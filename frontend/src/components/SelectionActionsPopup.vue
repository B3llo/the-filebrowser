<template>
  <div
    class="fb-sel-actions"
    v-show="fileStore.selectedCount > 0"
    ref="popupRef"
  >
    <button
      class="fb-sel-actions-trigger"
      :title="t('files.selectedItems', fileStore.selectedCount)"
      @click.stop="togglePopup"
    >
      <FbIcon name="more-vertical" size="18px" />
      <span class="fb-sel-actions-count">{{ fileStore.selectedCount }}</span>
    </button>
    <div class="fb-sel-actions-menu" v-show="isOpen" @click.stop>
      <button class="fb-sel-action-item" @click="handleInfo">
        <FbIcon name="info" size="16px" />
        <span>{{ t('buttons.info') }}</span>
      </button>
      <button
        v-if="headerButtons.download"
        class="fb-sel-action-item"
        @click="handleDownload"
      >
        <FbIcon name="download" size="16px" />
        <span>{{ t('buttons.download') }}</span>
      </button>
      <button
        v-if="headerButtons.share"
        class="fb-sel-action-item"
        @click="handleShare"
      >
        <FbIcon name="share" size="16px" />
        <span>{{ t('buttons.share') }}</span>
      </button>
      <button class="fb-sel-action-item fb-sel-action--disabled" aria-disabled="true">
        <FbIcon name="star" size="16px" />
        <span>{{ t('buttons.star', 'Star') }}</span>
      </button>
      <div class="fb-menu-divider"></div>
      <button
        v-if="headerButtons.rename"
        class="fb-sel-action-item"
        @click="handleRename"
      >
        <FbIcon name="rename" size="16px" />
        <span>{{ t('buttons.rename') }}</span>
      </button>
      <button
        v-if="headerButtons.copy"
        class="fb-sel-action-item"
        @click="handleCopy"
      >
        <FbIcon name="copy" size="16px" />
        <span>{{ t('buttons.copyFile') }}</span>
      </button>
      <button
        v-if="headerButtons.move"
        class="fb-sel-action-item"
        @click="handleMove"
      >
        <FbIcon name="move" size="16px" />
        <span>{{ t('buttons.moveFile') }}</span>
      </button>
      <div class="fb-menu-divider"></div>
      <button class="fb-sel-action-item" @click="handleSelectAll">
        <FbIcon name="check" size="16px" />
        <span>{{ t('buttons.selectAll', 'Select All') }}</span>
      </button>
      <div class="fb-menu-divider"></div>
      <button
        v-if="headerButtons.delete"
        class="fb-sel-action-item fb-sel-action--danger"
        @click="handleDelete"
      >
        <FbIcon name="delete" size="16px" />
        <span>{{ t('buttons.delete') }}</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
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
    copy?: boolean;
    delete?: boolean;
  };
  download: () => void;
}>();

const { t } = useI18n();
const fileStore = useFileStore();
const layoutStore = useLayoutStore();

const isOpen = ref(false);
const popupRef = ref<HTMLElement | null>(null);

const togglePopup = () => {
  isOpen.value = !isOpen.value;
};

const closePopup = () => {
  isOpen.value = false;
};

const handleInfo = () => {
  layoutStore.toggleDetails();
  closePopup();
};

const handleDownload = () => {
  props.download();
  closePopup();
};

const handleShare = () => {
  layoutStore.showHover("share");
  closePopup();
};

const handleRename = () => {
  layoutStore.showHover("rename");
  closePopup();
};

const handleCopy = () => {
  layoutStore.showHover("copy");
  closePopup();
};

const handleMove = () => {
  layoutStore.showHover("move");
  closePopup();
};

const handleSelectAll = () => {
  if (fileStore.req?.items) {
    fileStore.selected = fileStore.req.items.map((_, i) => i);
  }
  closePopup();
};

const handleDelete = () => {
  layoutStore.showHover("delete");
  closePopup();
};

const handleClickOutside = (event: MouseEvent) => {
  if (popupRef.value && !popupRef.value.contains(event.target as Node)) {
    closePopup();
  }
};

onMounted(() => {
  document.addEventListener("click", handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener("click", handleClickOutside);
});
</script>
