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
      :style="menuStyle"
      @click.stop
      ref="menuRef"
    >
      <button class="fb-sel-action-item" @click="handleInfo">
        <FbIcon name="info" size="16px" />
        <span>{{ t("buttons.info") }}</span>
      </button>
      <button
        v-if="headerButtons.download"
        class="fb-sel-action-item"
        @click="handleDownload"
      >
        <FbIcon name="download" size="16px" />
        <span>{{ t("buttons.download") }}</span>
      </button>
      <button
        v-if="headerButtons.share"
        class="fb-sel-action-item"
        @click="handleShare"
      >
        <FbIcon name="share" size="16px" />
        <span>{{ t("buttons.share") }}</span>
      </button>
      <button
        class="fb-sel-action-item"
        @click="handleStar"
      >
        <FbIcon name="star" size="16px" />
        <span>{{ t("buttons.star", "Star") }}</span>
      </button>
      <div class="fb-menu-divider"></div>
      <button
        v-if="headerButtons.rename"
        class="fb-sel-action-item"
        @click="handleRename"
      >
        <FbIcon name="rename" size="16px" />
        <span>{{ t("buttons.rename") }}</span>
      </button>
      <button
        v-if="headerButtons.copy"
        class="fb-sel-action-item"
        @click="handleCopy"
      >
        <FbIcon name="copy" size="16px" />
        <span>{{ t("buttons.copyFile") }}</span>
      </button>
      <button
        v-if="headerButtons.move"
        class="fb-sel-action-item"
        @click="handleMove"
      >
        <FbIcon name="move" size="16px" />
        <span>{{ t("buttons.moveFile") }}</span>
      </button>
      <div class="fb-menu-divider"></div>
      <button class="fb-sel-action-item" @click="handleSelectAll">
        <FbIcon name="check" size="16px" />
        <span>{{ t("buttons.selectAll", "Select All") }}</span>
      </button>
      <div class="fb-menu-divider"></div>
      <button
        v-if="headerButtons.delete"
        class="fb-sel-action-item fb-sel-action--danger"
        @click="handleDelete"
      >
        <FbIcon name="delete" size="16px" />
        <span>{{ t("buttons.delete") }}</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from "vue";
import { useI18n } from "vue-i18n";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { toggleStarred } from "@/utils/starred";
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
const menuRef = ref<HTMLElement | null>(null);
// Kept hidden until measured & positioned so it never flashes at (0,0)
const menuStyle = ref<Record<string, string>>({ visibility: "hidden" });

const togglePopup = async () => {
  isOpen.value = !isOpen.value;
  if (isOpen.value) {
    menuStyle.value = { visibility: "hidden" };
    await nextTick();
    // rAF guarantees the browser has laid the menu out before we measure
    requestAnimationFrame(positionMenu);
  }
};

const closePopup = () => {
  isOpen.value = false;
  menuStyle.value = { visibility: "hidden" };
};

const positionMenu = () => {
  const trigger = popupRef.value?.querySelector(
    ".fb-sel-actions-trigger"
  ) as HTMLElement | null;
  const menu = menuRef.value;
  if (!trigger || !menu) return;

  const triggerRect = trigger.getBoundingClientRect();
  // offsetWidth/Height are reliable layout metrics; the menu is position:fixed
  // in CSS so it is already shrink-wrapped to its content, not the toolbar.
  const menuWidth = menu.offsetWidth;
  const menuHeight = menu.offsetHeight;
  const gap = 8;
  const vw = window.innerWidth;
  const vh = window.innerHeight;

  // Default: hang below the trigger, right edge aligned to the trigger's right
  let left = triggerRect.right - menuWidth;
  let top = triggerRect.bottom + gap;

  // Clamp horizontally inside the viewport (handles right- and left-edge cases)
  if (left + menuWidth > vw - gap) left = vw - menuWidth - gap;
  if (left < gap) left = gap;

  // Flip above the trigger if it would overflow the bottom, then clamp top
  if (top + menuHeight > vh - gap) top = triggerRect.top - gap - menuHeight;
  if (top < gap) top = gap;

  menuStyle.value = {
    visibility: "visible",
    top: `${top}px`,
    left: `${left}px`,
  };
};

const handleInfo = () => {
  layoutStore.showDetails = true;
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

const handleStar = () => {
  const items = fileStore.req?.items;
  if (!items || fileStore.selected.length === 0) return;
  const item = items[fileStore.selected[0]];
  if (!item) return;
  toggleStarred({ url: item.url, name: item.name, type: item.type });
  closePopup();
};

const handleDelete = () => {
  layoutStore.showHover("delete");
  closePopup();
};

const handleClickOutside = (event: MouseEvent) => {
  if (
    popupRef.value &&
    !popupRef.value.contains(event.target as Node) &&
    menuRef.value &&
    !menuRef.value.contains(event.target as Node)
  ) {
    closePopup();
  }
};

// Close (don't try to reposition) when the viewport changes while open — a
// position:fixed menu would otherwise float at stale coordinates.
const handleViewportChange = () => {
  if (isOpen.value) closePopup();
};

onMounted(() => {
  document.addEventListener("click", handleClickOutside);
  window.addEventListener("resize", handleViewportChange);
  window.addEventListener("scroll", handleViewportChange, true);
});

onUnmounted(() => {
  document.removeEventListener("click", handleClickOutside);
  window.removeEventListener("resize", handleViewportChange);
  window.removeEventListener("scroll", handleViewportChange, true);
});
</script>
