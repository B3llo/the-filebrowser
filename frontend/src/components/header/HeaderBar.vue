<template>
  <header class="fb-toolbar">
    <!-- Mobile sidebar trigger (hamburger) — opens the sliding sidebar on <=736px -->
    <button
      class="fb-toolbar-btn fb-mobile-menu"
      :aria-label="t('buttons.toggleSidebar')"
      :title="t('buttons.toggleSidebar')"
      @click="layoutStore.showHover('sidebar')"
    >
      <FbIcon name="menu" size="20px" />
    </button>

    <!-- Logo (used by Preview/Editor which don't have a sidebar context on mobile) -->
    <img
      v-if="showLogo && !showBreadcrumb"
      :src="logoURL"
      class="fb-toolbar-logo"
      alt="FileBrowser"
    />

    <!-- Breadcrumb integrated into toolbar -->
    <div v-if="showBreadcrumb" class="fb-toolbar-breadcrumb">
      <Breadcrumbs :base="base ?? '/files'" />
    </div>

    <!-- Default slot content (search, title, etc.) -->
    <slot />

    <!-- Right-side actions -->
    <div
      id="dropdown"
      class="fb-toolbar-actions"
      :class="{ active: layoutStore.currentPromptName === 'more' }"
    >
      <slot name="actions" />
    </div>

    <!-- Mobile overflow trigger -->
    <button
      v-if="ifActionsSlot"
      id="more"
      class="fb-toolbar-btn"
      :aria-label="t('buttons.more')"
      @click="layoutStore.showHover('more')"
    >
      <FbIcon name="more-vertical" size="20px" />
    </button>

    <!-- Mobile overlay for dropdown -->
    <div
      class="overlay"
      v-show="layoutStore.currentPromptName == 'more'"
      @click="layoutStore.closeHovers"
    />
  </header>
</template>

<script setup lang="ts">
import { useLayoutStore } from "@/stores/layout";
import { logoURL } from "@/utils/constants";
import Breadcrumbs from "@/components/Breadcrumbs.vue";
import FbIcon from "@/components/FbIcon.vue";
import { computed, useSlots } from "vue";
import { useI18n } from "vue-i18n";

defineProps<{
  showLogo?: boolean;
  showBreadcrumb?: boolean;
  base?: string;
}>();

const layoutStore = useLayoutStore();
const slots = useSlots();
const { t } = useI18n();

const ifActionsSlot = computed(() => !!slots.actions);
</script>

<style scoped>
.fb-toolbar-logo {
  height: 2em;
  margin-right: 0.5em;
  flex-shrink: 0;
}

.fb-toolbar-breadcrumb {
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

.fb-toolbar-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 0 0 auto;
}

.fb-toolbar-title {
  font-size: 1.1em;
  font-weight: 600;
  color: var(--text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Inherit base toolbar-btn style for the overflow button */
.fb-toolbar-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--dim);
  cursor: pointer;
  padding: 0;
  transition:
    background 0.1s,
    color 0.1s;
}

.fb-toolbar-btn:hover {
  background: var(--hover);
  color: var(--text);
}

/* Mobile-only hamburger: opens the sliding sidebar. Hidden on desktop. */
.fb-mobile-menu {
  display: none;
}

@media (max-width: 736px) {
  .fb-mobile-menu {
    display: inline-flex;
  }
}
</style>
