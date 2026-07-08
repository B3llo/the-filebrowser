<template>
  <div
    v-show="isMobileActive"
    class="fb-sidebar-backdrop"
    @click="closeHovers"
  />
  <nav
    class="fb-sidebar"
    :class="{ active: isMobileActive, collapsed: sidebarCollapsed }"
  >
    <!-- Logo header — clicking the brand opens the first/default source -->
    <div class="fb-sidebar-header">
      <button
        class="fb-sidebar-brand"
        @click="toRoot"
        :aria-label="$t('files.home', 'Home')"
        :title="$t('files.home', 'Home')"
      >
        <div class="fb-sidebar-logo-chip">
          <img :src="logoURL" alt="" aria-hidden="true" />
        </div>
        <span class="fb-sidebar-wordmark">FileBrowser</span>
      </button>
      <button
        class="fb-sidebar-collapse-btn"
        @click="toggleSidebarCollapsed"
        :aria-label="$t('buttons.toggleSidebar')"
        :title="$t('buttons.toggleSidebar')"
      >
        <fb-icon name="panel-left" size="18px" />
      </button>
    </div>

    <!-- Navigation destinations -->
    <ul class="fb-sidebar-nav" v-if="isLoggedIn" role="list">
      <!-- Sources section: label + "+" for admins (only when expanded) -->
      <li v-if="!sidebarCollapsed && (sourceList.length > 1 || isAdmin)" class="fb-nav-section-header">
        <span class="fb-nav-section-label">{{ $t("sidebar.sources", "Sources") }}</span>
        <button
          v-if="isAdmin"
          class="fb-nav-section-add"
          @click.stop="openAddSource"
          :title="$t('sidebar.addSource', 'Add source')"
        >
          <fb-icon name="plus" size="14px" />
        </button>
      </li>

      <!-- Flat source list -->
      <li v-for="src in sourceList" :key="src.id">
        <button
          class="fb-nav-item"
          :class="{ 'is-active': isActiveSource(src.id) }"
          @click="toSource(src.id)"
          :aria-label="src.name"
          :title="src.name"
        >
          <fb-icon :name="sourceIcon(src.id)" size="18px" />
          <span class="fb-nav-item-label">{{ src.name }}</span>
        </button>
      </li>
      <li>
        <button
          class="fb-nav-item"
          :class="{ 'is-active': $route.path === '/recent' }"
          @click="toRecent"
          :aria-label="$t('sidebar.recent')"
          :title="$t('sidebar.recent')"
        >
          <fb-icon name="clock" size="18px" />
          <span>{{ $t("sidebar.recent", "Recent") }}</span>
        </button>
      </li>
      <li>
        <button
          class="fb-nav-item"
          :class="{ 'is-active': $route.path === '/starred' }"
          @click="toStarred"
          :aria-label="$t('sidebar.starred')"
          :title="$t('sidebar.starred')"
        >
          <fb-icon name="star" size="18px" />
          <span>{{ $t("sidebar.starred", "Starred") }}</span>
        </button>
      </li>
      <li>
        <button
          class="fb-nav-item"
          @click="toShares"
          :aria-label="$t('sidebar.shares', 'Shares')"
        >
          <fb-icon name="users" size="18px" />
          <span>{{ $t("sidebar.shares", "Shares") }}</span>
        </button>
      </li>
      <li>
        <button
          class="fb-nav-item"
          :class="{ 'is-active': $route.path === '/trash' }"
          @click="toTrash"
          :aria-label="$t('sidebar.trash')"
          :title="$t('sidebar.trash')"
        >
          <fb-icon name="trash" size="18px" />
          <span>{{ $t("sidebar.trash", "Trash") }}</span>
        </button>
      </li>
    </ul>

    <!-- Unauthenticated state -->
    <ul class="fb-sidebar-nav" v-else role="list">
      <li v-if="!hideLoginButton">
        <router-link class="fb-nav-item" to="/login">
          <fb-icon name="logout" size="18px" />
          <span>{{ $t("sidebar.login") }}</span>
        </router-link>
      </li>
      <li v-if="signup">
        <router-link class="fb-nav-item" to="/login">
          <fb-icon name="person" size="18px" />
          <span>{{ $t("sidebar.signup") }}</span>
        </router-link>
      </li>
    </ul>

    <!-- Storage meter -->
    <div v-if="isLoggedIn && !disableUsedPercentage" class="fb-sidebar-storage">
      <div class="fb-sidebar-storage-header">
        <span class="fb-sidebar-storage-label">{{
          $t("sidebar.storage")
        }}</span>
      </div>
      <div class="fb-sidebar-source-row">
        <span class="fb-sidebar-row-label">
          <fb-icon name="folder" size="14px" />
          {{ activeSourceName || $t("sidebar.myFiles") }}
        </span>
        <span class="fb-sidebar-row-value">{{ sourceSizeFormatted }}</span>
      </div>
      <div class="fb-sidebar-disk-row">
        <span class="fb-sidebar-row-label">
          <fb-icon name="home" size="14px" />
          {{ $t("sidebar.disk") }}
        </span>
        <span class="fb-sidebar-row-value">
          {{ usage.usedPercentage }}% · {{ usage.total }}
        </span>
      </div>
      <div class="fb-sidebar-storage-bar">
        <progress-bar
          :val="usage.usedPercentage"
          :size="3"
          :bg-color="'var(--hover)'"
          :bar-color="'var(--accent)'"
          :bar-border-radius="3"
        />
      </div>
    </div>

    <!-- User footer -->
    <div v-if="isLoggedIn" class="fb-sidebar-footer">
      <button
        class="fb-sidebar-user"
        @click="toProfile"
        :aria-label="$t('settings.profileSettings')"
        :title="$t('settings.profileSettings')"
      >
        <AvatarBadge :user="user" :size="30" />
        <div class="fb-sidebar-user-info">
          <span class="fb-sidebar-username">{{
            user.displayName || user.username
          }}</span>
          <span v-if="user.perm.admin" class="fb-sidebar-role">{{
            $t("settings.admin")
          }}</span>
        </div>
      </button>
      <div class="fb-sidebar-footer-actions">
        <button
          class="fb-icon-btn"
          @click="toggleTheme"
          :aria-label="$t('settings.toggleTheme')"
          :title="$t('settings.toggleTheme')"
        >
          <fb-icon :name="isDark ? 'sun' : 'moon'" size="16px" />
        </button>
        <button
          class="fb-icon-btn"
          @click="openSettingsModal"
          :aria-label="$t('sidebar.settings')"
          :title="$t('sidebar.settings')"
        >
          <fb-icon name="settings" size="16px" />
        </button>
        <button
          v-if="canLogout"
          class="fb-icon-btn"
          @click="logout"
          :aria-label="$t('sidebar.logout')"
          :title="$t('sidebar.logout')"
        >
          <fb-icon name="logout" size="16px" />
        </button>
      </div>
    </div>
  </nav>
</template>

<script>
import { reactive, ref } from "vue";
import { mapActions, mapState } from "pinia";
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { useSourceStore } from "@/stores/source";

import * as auth from "@/utils/auth";
import {
  version,
  signup,
  hideLoginButton,
  disableExternal,
  disableUsedPercentage,
  noAuth,
  logoutPage,
  loginPage,
  logoURL,
} from "@/utils/constants";
import { files as api, users } from "@/api";
import {
  toggleTheme as toggleThemeUtil,
  getTheme,
  getMediaPreference,
} from "@/utils/theme";
import ProgressBar from "@/components/ProgressBar.vue";
import FbIcon from "@/components/FbIcon.vue";
import AvatarBadge from "@/components/AvatarBadge.vue";
import prettyBytes from "pretty-bytes";

const USAGE_DEFAULT = { used: "0 B", total: "0 B", usedPercentage: 0 };

export default {
  name: "sidebar",
  setup() {
    const usage = reactive(USAGE_DEFAULT);
    const sourceDirSize = ref(null);
    const loadingSourceDirSize = ref(false);
    return {
      usage,
      sourceDirSize,
      loadingSourceDirSize,
      usageAbortController: new AbortController(),
    };
  },
  components: {
    ProgressBar,
    FbIcon,
    AvatarBadge,
  },
  inject: ["$showError"],
  computed: {
    ...mapState(useAuthStore, ["user", "isLoggedIn"]),
    ...mapState(useFileStore, ["isFiles", "reload"]),
    ...mapState(useLayoutStore, ["currentPromptName", "sidebarCollapsed"]),
    ...mapState(useSourceStore, ["sources", "hasMultiple", "activeId", "active"]),
    sourceList() {
      // Once sources are loaded they replace the navigation; while loading on
      // first paint, show the implicit "My Files" so the sidebar is never empty.
      return this.sources.length
        ? this.sources
        : [{ id: 0, name: this.$t("sidebar.myFiles") }];
    },
    isMobileActive() {
      return this.currentPromptName === "sidebar";
    },
    isMyFiles() {
      return this.$route.path.startsWith("/files");
    },
    activeRouteSourceId() {
      return this.$route.params?.sourceId;
    },
    activeSourceName() {
      return this.active?.name ?? "";
    },
    signup: () => signup,
    hideLoginButton: () => hideLoginButton,
    version: () => version,
    disableExternal: () => disableExternal,
    disableUsedPercentage: () => disableUsedPercentage,
    logoURL: () => logoURL,
    canLogout: () => !noAuth && (loginPage || logoutPage !== "/login"),
    isDark() {
      const t = this.user?.theme;
      if (t === "dark") return true;
      if (t === "light") return false;
      return getMediaPreference() === "dark";
    },
    isAdmin() {
      return this.user?.perm?.admin ?? false;
    },
    sourceSizeFormatted() {
      if (this.sourceDirSize === null) return "—";
      if (this.loadingSourceDirSize) return "…";
      return prettyBytes(this.sourceDirSize, { binary: true });
    },
  },
  methods: {
    ...mapActions(useLayoutStore, [
      "closeHovers",
      "showHover",
      "openSettings",
      "toggleSidebarCollapsed",
    ]),
    isActiveSource(id) {
      return (
        this.$route.path.startsWith("/files") &&
        String(this.activeRouteSourceId) === String(id)
      );
    },
    sourceIcon(id) {
      return Number(id) === 0 ? "folder" : "folder";
    },
    toSource(id) {
      this.$router.push({ path: `/files/${id}/` });
      this.closeHovers();
    },
    openAddSource() {
      this.showHover("addSource");
    },
    abortOngoingFetchUsage() {
      this.usageAbortController.abort();
    },
    async fetchUsage() {
      let path = this.$route.path;
      if (!path.includes("/files")) {
        path = `/files/${this.activeId}/`;
      }
      path = path.endsWith("/") ? path : path + "/";
      let usageStats = USAGE_DEFAULT;
      if (this.disableUsedPercentage) {
        return Object.assign(this.usage, usageStats);
      }
      try {
        this.abortOngoingFetchUsage();
        this.usageAbortController = new AbortController();
        const usage = await api.usage(path, this.usageAbortController.signal);
        usageStats = {
          used: prettyBytes(usage.used, { binary: true }),
          total: prettyBytes(usage.total, { binary: true }),
          usedPercentage: Math.round((usage.used / usage.total) * 100),
        };
      } finally {
        return Object.assign(this.usage, usageStats);
      }
    },
    async fetchSourceDirSize() {
      const source = this.active;
      // Skip the implicit source (id=0 / no explicit path): computing its size
      // would walk the user's entire default scope, which can be very slow.
      if (!source || source.path === undefined || source.path === "") {
        this.sourceDirSize = null;
        this.loadingSourceDirSize = false;
        return;
      }
      this.loadingSourceDirSize = true;
      try {
        const res = await api.dirSize(`/files/${source.id}/`);
        this.sourceDirSize = res.size;
      } catch {
        this.sourceDirSize = null;
      } finally {
        this.loadingSourceDirSize = false;
      }
    },
    toRoot() {
      const sourceStore = useSourceStore();
      this.$router.push({ path: `${sourceStore.filesBase}/` });
      this.closeHovers();
    },
    openSettingsModal() {
      this.openSettings();
      this.closeHovers();
    },
    toRecent() {
      this.$router.push({ path: "/recent" });
      this.closeHovers();
    },
    toStarred() {
      this.$router.push({ path: "/starred" });
      this.closeHovers();
    },
    toTrash() {
      this.$router.push({ path: "/trash" });
      this.closeHovers();
    },
    toShares() {
      this.$router.push({ path: "/settings/shares" });
      this.closeHovers();
    },
    toProfile() {
      this.$router.push({ path: "/settings/profile" });
      this.closeHovers();
    },
    logout: auth.logout,
    async toggleTheme() {
      toggleThemeUtil();
      const newTheme = getTheme();
      if (this.user) {
        const data = { id: this.user.id, theme: newTheme };
        await users.update(data, ["theme"]).catch(() => {});
        const authStore = useAuthStore();
        authStore.updateUser(data);
      }
    },
  },
  watch: {
    $route: {
      handler() {
        this.fetchUsage();
        this.closeHovers();
      },
      immediate: true,
    },
    active: {
      handler() {
        this.fetchSourceDirSize();
      },
      immediate: true,
    },
  },
  unmounted() {
    this.abortOngoingFetchUsage();
  },
};
</script>

<style scoped>
/* Backdrop (mobile only — hidden on desktop via CSS) */
.fb-sidebar-backdrop {
  display: none;
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.35);
  z-index: 99998;
  animation: fbfade 0.15s ease;
}

@media (max-width: 736px) {
  .fb-sidebar-backdrop {
    display: block;
  }
}

/* Sidebar shell */
.fb-sidebar {
  display: flex;
  flex-direction: column;
  width: 256px;
  flex-shrink: 0;
  height: 100%;
  overflow-y: auto;
  background: var(--sidebar);
  border-right: 1px solid var(--border);
  padding: 0 0 12px;
  transition: width 0.16s ease;
}

/* Logo header */
.fb-sidebar-header {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 12px 12px 8px;
  flex-shrink: 0;
}

.fb-sidebar-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
  min-width: 0;
  padding: 6px 8px;
  border: none;
  background: transparent;
  border-radius: 8px;
  cursor: pointer;
  text-align: left;
  transition: background 0.1s;
}

.fb-sidebar-brand:hover {
  background: var(--hover);
}

.fb-sidebar-collapse-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border-radius: 8px;
  border: none;
  background: transparent;
  color: var(--faint);
  cursor: pointer;
  flex-shrink: 0;
  transition:
    background 0.1s,
    color 0.1s;
}

.fb-sidebar-collapse-btn:hover {
  background: var(--hover);
  color: var(--text);
}

.fb-sidebar-logo-chip {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  overflow: hidden;
}

.fb-sidebar-logo-chip img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.fb-sidebar-wordmark {
  font-size: 16px;
  font-weight: 650;
  letter-spacing: -0.3px;
  color: var(--text);
  white-space: nowrap;
}

/* Nav list */
.fb-sidebar-nav {
  list-style: none;
  margin: 4px 0 0;
  padding: 0 8px;
  flex-shrink: 0;
}

.fb-sidebar-nav li {
  margin: 1px 0;
}

/* Nav item — shared base for button, router-link, and placeholder */
.fb-nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 8px 10px;
  border-radius: 8px;
  font-size: 13.5px;
  font-weight: 500;
  color: var(--dim);
  background: transparent;
  border: none;
  cursor: pointer;
  text-align: left;
  text-decoration: none;
  transition:
    background 0.1s,
    color 0.1s;
  line-height: 1.3;
}

.fb-nav-item:hover {
  background: var(--hover);
  color: var(--text);
}

.fb-nav-item.is-active,
.router-link-active.fb-nav-item {
  background: var(--sel);
  color: var(--accent);
}

.fb-nav-item-label {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Placeholder items (Coming soon) */
.fb-nav-item--placeholder {
  opacity: 0.38;
  cursor: not-allowed;
  pointer-events: none;
}

/* Storage meter */
.fb-sidebar-storage {
  margin: auto 12px 0;
  padding: 12px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 10px;
  flex-shrink: 0;
}

.fb-sidebar-storage-header {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 6px;
}

.fb-sidebar-storage-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text);
}

.fb-sidebar-source-row,
.fb-sidebar-disk-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 0;
  font-size: 12.5px;
  margin-top: 2px;
}

.fb-sidebar-row-label {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--text);
  font-weight: 500;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.fb-sidebar-row-value {
  color: var(--dim);
  font-weight: 500;
  flex-shrink: 0;
  margin-left: 8px;
  font-variant-numeric: tabular-nums;
}

.fb-sidebar-storage-bar {
  margin-top: 8px;
}

/* User footer */
.fb-sidebar-footer {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px 4px;
  margin-top: 10px;
  border-top: 1px solid var(--border);
  flex-shrink: 0;
}

.fb-sidebar-user {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
  border: none;
  background: transparent;
  cursor: pointer;
  padding: 4px 6px;
  border-radius: 8px;
  text-align: left;
  transition: background 0.1s;
}

.fb-sidebar-user:hover {
  background: var(--hover);
}

.fb-sidebar-user-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.fb-sidebar-username {
  font-size: 13px;
  font-weight: 500;
  color: var(--text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.fb-sidebar-role {
  font-size: 11px;
  color: var(--dim);
}

.fb-sidebar-footer-actions {
  display: flex;
  gap: 2px;
  flex-shrink: 0;
}

.fb-icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: var(--faint);
  cursor: pointer;
  transition:
    background 0.1s,
    color 0.1s;
  padding: 0;
}

.fb-icon-btn:hover {
  background: var(--hover);
  color: var(--text);
}

/* Source section header */
.fb-nav-section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 10px 2px;
  margin-top: 4px;
}

.fb-nav-section-label {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: var(--dim);
}

.fb-nav-section-add {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 4px;
  border: none;
  background: transparent;
  color: var(--faint);
  cursor: pointer;
  transition:
    background 0.1s,
    color 0.1s;
  padding: 0;
}

.fb-nav-section-add:hover {
  background: var(--hover);
  color: var(--accent);
}

/* Collapse toggle is a desktop-only affordance */
@media (max-width: 736px) {
  .fb-sidebar-collapse-btn {
    display: none;
  }
}

/* --- Collapsed rail (desktop only) --- */
@media (min-width: 737px) {
  .fb-sidebar.collapsed {
    width: 68px;
  }

  /* Header stacks the logo above the toggle, both centered */
  .fb-sidebar.collapsed .fb-sidebar-header {
    flex-direction: column;
    gap: 6px;
    padding: 12px 0 8px;
  }

  .fb-sidebar.collapsed .fb-sidebar-brand {
    flex: 0 0 auto;
    justify-content: center;
    padding: 6px;
  }

  .fb-sidebar.collapsed .fb-sidebar-wordmark {
    display: none;
  }

  /* Nav items become icon-only, centered (tooltips via title attr) */
  .fb-sidebar.collapsed .fb-sidebar-nav {
    padding: 0 10px;
  }

  .fb-sidebar.collapsed .fb-nav-item {
    justify-content: center;
    padding: 8px 0;
  }

  /* Hide only the text label, never the FbIcon (its root is span.fb-icon) */
  .fb-sidebar.collapsed .fb-nav-item span:not(.fb-icon) {
    display: none;
  }

  /* Hide the storage meter in the rail */
  .fb-sidebar.collapsed .fb-sidebar-storage {
    display: none;
  }

  /* Footer collapses to avatar + stacked action icons */
  .fb-sidebar.collapsed .fb-sidebar-footer {
    flex-direction: column;
    gap: 6px;
    padding: 10px 0 4px;
    margin-top: auto;
  }

  .fb-sidebar.collapsed .fb-sidebar-user {
    justify-content: center;
    padding: 4px;
  }

  .fb-sidebar.collapsed .fb-sidebar-user-info {
    display: none;
  }

  /* Collapsed rail: keep only the avatar, hide settings/logout actions */
  .fb-sidebar.collapsed .fb-sidebar-footer-actions {
    display: none;
  }
}
</style>
