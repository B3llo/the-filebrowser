<template>
  <div class="dashboard">
    <header-bar showMenu>
      <span class="fb-toolbar-title">{{ sectionTitle }}</span>
    </header-bar>

    <div class="fb-settings-content">
      <nav class="fb-settings-nav">
        <ul class="fb-settings-nav-list">
          <li>
            <router-link
              to="/settings/profile"
              class="fb-settings-nav-item"
              :class="{ 'is-active': $route.path === '/settings/profile' }"
            >
              <svg
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                style="width: 18px; height: 18px"
              >
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
                <circle cx="12" cy="7" r="4" />
              </svg>
              {{ t("settings.profileSettings") }}
            </router-link>
          </li>
          <li v-if="user?.perm.share">
            <router-link
              to="/settings/shares"
              class="fb-settings-nav-item"
              :class="{ 'is-active': $route.path === '/settings/shares' }"
            >
              <svg
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                style="width: 18px; height: 18px"
              >
                <circle cx="18" cy="5" r="3" />
                <circle cx="6" cy="12" r="3" />
                <circle cx="18" cy="19" r="3" />
                <line x1="8.59" y1="13.51" x2="15.42" y2="17.49" />
                <line x1="15.41" y1="6.51" x2="8.59" y2="10.49" />
              </svg>
              {{ t("settings.shareManagement") }}
            </router-link>
          </li>
          <li v-if="user?.perm.admin">
            <router-link
              to="/settings/global"
              class="fb-settings-nav-item"
              :class="{ 'is-active': $route.path === '/settings/global' }"
            >
              <svg
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                style="width: 18px; height: 18px"
              >
                <circle cx="12" cy="12" r="3" />
                <path
                  d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"
                />
              </svg>
              {{ t("settings.globalSettings") }}
            </router-link>
          </li>
          <li v-if="user?.perm.admin">
            <router-link
              to="/settings/sources"
              class="fb-settings-nav-item"
              :class="{
                'is-active':
                  $route.path === '/settings/sources' ||
                  $route.name === 'Source',
              }"
            >
              <svg
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                style="width: 18px; height: 18px"
              >
                <path
                  d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"
                />
              </svg>
              {{ t("settings.sourceManagement") }}
            </router-link>
          </li>
          <li v-if="user?.perm.admin">
            <router-link
              to="/settings/users"
              class="fb-settings-nav-item"
              :class="{
                'is-active':
                  $route.path === '/settings/users' || $route.name === 'User',
              }"
            >
              <svg
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                style="width: 18px; height: 18px"
              >
                <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
                <circle cx="9" cy="7" r="4" />
                <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
                <path d="M16 3.13a4 4 0 0 1 0 7.75" />
              </svg>
              {{ t("settings.userManagement") }}
            </router-link>
          </li>
        </ul>
      </nav>

      <main class="fb-settings-main">
        <div v-if="loading">
          <h2 class="message delayed">
            <div class="spinner">
              <div class="bounce1"></div>
              <div class="bounce2"></div>
              <div class="bounce3"></div>
            </div>
            <span>{{ t("files.loading") }}</span>
          </h2>
        </div>

        <router-view></router-view>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import HeaderBar from "@/components/header/HeaderBar.vue";
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";

const { t } = useI18n();
const route = useRoute();

const authStore = useAuthStore();
const layoutStore = useLayoutStore();

const user = computed(() => authStore.user);
const loading = computed(() => layoutStore.loading);

const sectionTitles: Record<string, string> = {
  ProfileSettings: "settings.profileSettings",
  Shares: "settings.shareManagement",
  GlobalSettings: "settings.globalSettings",
  Users: "settings.users",
  User: "settings.user",
  Sources: "settings.sources",
  Source: "settings.source",
};

const sectionTitle = computed(() => {
  const name = route.name as string;
  return t(sectionTitles[name] || "sidebar.settings");
});
</script>
