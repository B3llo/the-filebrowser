<template>
  <header class="fb-toolbar">
    <!-- Mobile hamburger -->
    <button
      v-if="showMenu"
      class="fb-toolbar-btn fb-toolbar-menu"
      :aria-label="t('buttons.toggleSidebar')"
      @click="layoutStore.showHover('sidebar')"
    >
      <i class="material-icons">menu</i>
    </button>

    <!-- Logo (used by Preview/Editor which don't have a sidebar context on mobile) -->
    <img v-if="showLogo && !showBreadcrumb" :src="logoURL" class="fb-toolbar-logo" alt="FileBrowser" />

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
      <i class="material-icons">more_vert</i>
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
import { computed, useSlots } from "vue";
import { useI18n } from "vue-i18n";

defineProps<{
  showLogo?: boolean;
  showMenu?: boolean;
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

.fb-toolbar-menu {
  margin-right: 4px;
  display: none;
}

@media (max-width: 736px) {
  .fb-toolbar-menu {
    display: flex;
  }
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
  transition: background 0.1s, color 0.1s;
}

.fb-toolbar-btn:hover {
  background: var(--hover);
  color: var(--text);
}
</style>
