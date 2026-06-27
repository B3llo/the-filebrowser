<template>
  <div
    v-if="layoutStore.showSettings"
    class="fb-settings-overlay"
    @click.self="close"
  >
    <div class="fb-settings-modal">
      <div class="fb-settings-header">
        <span class="fb-settings-title">{{ t("settings.settings") }}</span>
        <button class="fb-settings-close" @click="close" :aria-label="t('buttons.close')">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="width:17px;height:17px">
            <path d="M18 6L6 18"/>
            <path d="M6 6l12 12"/>
          </svg>
        </button>
      </div>

      <div class="fb-settings-body">
        <!-- Appearance -->
        <div class="fb-settings-row">
          <div>
            <div class="fb-settings-row-label">{{ t("settings.appearance", "Appearance") }}</div>
            <div class="fb-settings-row-desc">{{ t("settings.appearanceDesc", "Light or dark interface") }}</div>
          </div>
          <div class="fb-settings-seg">
            <button
              class="fb-settings-seg-btn"
              :class="{ 'fb-settings-seg-btn--active': currentTheme === 'light' }"
              @click="applyTheme('light')"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round" style="width:15px;height:15px">
                <path d="M12 17a5 5 0 1 0 0-10 5 5 0 0 0 0 10Z"/>
                <path d="M12 1v2"/>
                <path d="M12 21v2"/>
                <path d="M4.2 4.2l1.4 1.4"/>
                <path d="M18.4 18.4l1.4 1.4"/>
                <path d="M1 12h2"/>
                <path d="M21 12h2"/>
                <path d="M4.2 19.8l1.4-1.4"/>
                <path d="M18.4 5.6l1.4-1.4"/>
              </svg>
              {{ t("settings.themes.light") }}
            </button>
            <button
              class="fb-settings-seg-btn"
              :class="{ 'fb-settings-seg-btn--active': currentTheme === 'dark' }"
              @click="applyTheme('dark')"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round" style="width:15px;height:15px">
                <path d="M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8Z"/>
              </svg>
              {{ t("settings.themes.dark") }}
            </button>
          </div>
        </div>

        <!-- Density -->
        <div class="fb-settings-row">
          <div>
            <div class="fb-settings-row-label">{{ t("settings.density", "Density") }}</div>
            <div class="fb-settings-row-desc">{{ t("settings.densityDesc", "Spacing between items") }}</div>
          </div>
          <div class="fb-settings-seg">
            <button
              class="fb-settings-seg-btn"
              :class="{ 'fb-settings-seg-btn--active': currentDensity === 'comfortable' }"
              @click="setDensity('comfortable')"
            >
              {{ t("settings.comfortable", "Comfortable") }}
            </button>
            <button
              class="fb-settings-seg-btn"
              :class="{ 'fb-settings-seg-btn--active': currentDensity === 'compact' }"
              @click="setDensity('compact')"
            >
              {{ t("settings.compact", "Compact") }}
            </button>
          </div>
        </div>

        <!-- Default view -->
        <div class="fb-settings-row">
          <div>
            <div class="fb-settings-row-label">{{ t("settings.defaultView", "Default view") }}</div>
            <div class="fb-settings-row-desc">{{ t("settings.defaultViewDesc", "How folders open") }}</div>
          </div>
          <div class="fb-settings-seg">
            <button
              class="fb-settings-seg-btn"
              :class="{ 'fb-settings-seg-btn--active': currentView === 'list' }"
              @click="setView('list')"
            >
              {{ t("settings.list", "List") }}
            </button>
            <button
              class="fb-settings-seg-btn"
              :class="{ 'fb-settings-seg-btn--active': currentView === 'mosaic' }"
              @click="setView('mosaic')"
            >
              {{ t("settings.grid", "Grid") }}
            </button>
          </div>
        </div>

        <!-- Show hidden files -->
        <div class="fb-settings-row">
          <div>
            <div class="fb-settings-row-label">{{ t("settings.showHidden", "Show hidden files") }}</div>
            <div class="fb-settings-row-desc">{{ t("settings.showHiddenDesc", "Files starting with a dot") }}</div>
          </div>
          <button
            class="fb-settings-toggle"
            :class="hideDotfiles ? 'fb-settings-toggle--off' : 'fb-settings-toggle--on'"
            @click="toggleHidden"
            :aria-label="t('settings.showHidden')"
          >
            <span class="fb-settings-toggle-knob"></span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useLayoutStore } from "@/stores/layout";
import { useAuthStore } from "@/stores/auth";
import { getTheme, setTheme } from "@/utils/theme";
import { users as api } from "@/api";
import { computed, inject, ref, onMounted } from "vue";
import { useI18n } from "vue-i18n";

const $showError = inject<IToastError>("$showError")!;

const layoutStore = useLayoutStore();
const authStore = useAuthStore();
const { t } = useI18n();

const currentTheme = ref<UserTheme>(getTheme());

const currentDensity = ref<"comfortable" | "compact">("comfortable");

const currentView = computed(() => {
  const mode = authStore.user?.viewMode ?? "list";
  return mode === "mosaic gallery" ? "mosaic" : mode;
});

const hideDotfiles = computed(() => authStore.user?.hideDotfiles ?? false);

onMounted(() => {
  const density = localStorage.getItem("fb-density");
  if (density === "compact" || density === "comfortable") {
    currentDensity.value = density;
    document.documentElement.setAttribute("data-density", density);
  }
});

function close() {
  layoutStore.closeSettings();
}

function applyTheme(theme: UserTheme) {
  currentTheme.value = theme;
  setTheme(theme);
}

function setDensity(density: "comfortable" | "compact") {
  currentDensity.value = density;
  localStorage.setItem("fb-density", density);
  document.documentElement.setAttribute("data-density", density);
}

async function setView(mode: "list" | "mosaic") {
  if (authStore.user === null) return;
  if (authStore.user.viewMode === mode) return;
  const data = { id: authStore.user.id, viewMode: mode };
  await api.update(data, ["viewMode"]).catch($showError);
  authStore.updateUser(data);
}

async function toggleHidden() {
  if (authStore.user === null) return;
  const newVal = !authStore.user.hideDotfiles;
  const data = { id: authStore.user.id, hideDotfiles: newVal };
  await api.update(data, ["hideDotfiles"]).catch($showError);
  authStore.updateUser(data);
}
</script>

<style scoped>
/* No scoped styles needed — all styles in settings-modal.css */
</style>
