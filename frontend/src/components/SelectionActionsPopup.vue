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
    <div
      class="fb-sel-actions-menu"
      v-show="isOpen"
      @click.stop
      :style="menuStyle"
    >
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
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from "vue";
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
const menuTop = ref(0);
const menuLeft = ref(0);
const menuTransform = ref("");

const togglePopup = async () => {
  isOpen.value = !isOpen.value;
  if (isOpen.value) {
    await nextTick();
    calculatePosition();
  }
};

const closePopup = () => {
  isOpen.value = false;
};

const menuStyle = computed(() => ({
  top: `${menuTop.value}px`,
  left: `${menuLeft.value}px`,
  transform: menuTransform.value,
}));

const calculatePosition = () => {
  if (!popupRef.value) return;

  const trigger = popupRef.value.querySelector(".fb-sel-actions-trigger") as HTMLElement;
  const menu = popupRef.value.querySelector(".fb-sel-actions-menu") as HTMLElement;
  if (!trigger || !menu) return;

  const triggerRect = trigger.getBoundingClientRect();
  const menuRect = menu.getBoundingClientRect();
  const viewportWidth = window.innerWidth;
  const viewportHeight = window.innerHeight;
  const gap = 8;

  let top = triggerRect.bottom + gap;
  let left = triggerRect.right - menuRect.width;
  let transform = "";

  // If menu would go off the bottom, show above the trigger
  if (top + menuRect.height > viewportHeight) {
    top = triggerRect.top - menuRect.height - gap;
    transform = "none";
  }

  // If menu would go off the top, pin to top with some padding
  if (top < gap) {
    top = gap;
  }

  // If menu would go off the right edge, align to right edge of trigger
  if (left + menuRect.width > viewportWidth - gap) {
    left = viewportWidth - menuRect.width - gap;
  }

  // If menu would go off the left edge, pin to left with some padding
  if (left < gap) {
    left = gap;
  }

  menuTop.value = top;
  menuLeft.value = left;
  menuTransform.value = transform;
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

const handleScroll = () => {
  if (isOpen.value) {
    calculatePosition();
  }
};

onMounted(() => {
  document.addEventListener("click", handleClickOutside);
  document.addEventListener("scroll", handleScroll, true);
  window.addEventListener("resize", handleScroll);
});

onUnmounted(() => {
  document.removeEventListener("click", handleClickOutside);
  document.removeEventListener("scroll", handleScroll, true);
  window.removeEventListener("resize", handleScroll);
});
</script>
