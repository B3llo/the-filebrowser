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
        <span
          class="fb-nav-item fb-nav-item--placeholder"
          aria-disabled="true"
          :title="$t('sidebar.comingSoon', 'Coming soon')"
        >
          <fb-icon name="clock" size="18px" />
          <span>{{ $t("sidebar.recent", "Recent") }}</span>
        </span>
      </li>
      <li>
        <span
          class="fb-nav-item fb-nav-item--placeholder"
          aria-disabled="true"
          :title="$t('sidebar.comingSoon', 'Coming soon')"
        >
          <fb-icon name="star" size="18px" />
          <span>{{ $t("sidebar.starred", "Starred") }}</span>
        </span>
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
        <span
          class="fb-nav-item fb-nav-item--placeholder"
          aria-disabled="true"
          :title="$t('sidebar.comingSoon', 'Coming soon')"
        >
          <fb-icon name="trash" size="18px" />
          <span>{{ $t("sidebar.trash", "Trash") }}</span>
        </span>
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
    <div
      v-if="isLoggedIn && isFiles && !disableUsedPercentage"
      class="fb-sidebar-storage"
    >
      <div class="fb-sidebar-storage-header">
        <span class="fb-sidebar-storage-label">Storage</span>
        <span class="fb-sidebar-storage-pct">{{ usage.usedPercentage }}%</span>
      </div>
      <progress-bar :val="usage.usedPercentage" size="small" />
      <div class="fb-sidebar-storage-detail">
        {{ $t("sidebar.diskUsed", { used: usage.used, total: usage.total }) }}
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
        <div class="fb-sidebar-avatar" aria-hidden="true">
          {{ userInitial }}
        </div>
        <div class="fb-sidebar-user-info">
          <span class="fb-sidebar-username">{{ user.username }}</span>
          <span v-if="user.perm.admin" class="fb-sidebar-role">Admin</span>
        </div>
      </button>
      <div class="fb-sidebar-footer-actions">
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
import { reactive } from "vue";
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
import { files as api } from "@/api";
import ProgressBar from "@/components/ProgressBar.vue";
import FbIcon from "@/components/FbIcon.vue";
import prettyBytes from "pretty-bytes";

const USAGE_DEFAULT = { used: "0 B", total: "0 B", usedPercentage: 0 };

export default {
  name: "sidebar",
  setup() {
    const usage = reactive(USAGE_DEFAULT);
    return { usage, usageAbortController: new AbortController() };
  },
  components: {
    ProgressBar,
    FbIcon,
  },
  inject: ["$showError"],
  computed: {
    ...mapState(useAuthStore, ["user", "isLoggedIn"]),
    ...mapState(useFileStore, ["isFiles", "reload"]),
    ...mapState(useLayoutStore, ["currentPromptName", "sidebarCollapsed"]),
    ...mapState(useSourceStore, ["sources", "hasMultiple"]),
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
    userInitial() {
      return this.user?.username?.[0]?.toUpperCase() ?? "?";
    },
    signup: () => signup,
    hideLoginButton: () => hideLoginButton,
    version: () => version,
    disableExternal: () => disableExternal,
    disableUsedPercentage: () => disableUsedPercentage,
    logoURL: () => logoURL,
    canLogout: () => !noAuth && (loginPage || logoutPage !== "/login"),
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
      // The implicit legacy source keeps the generic folder icon; real sources
      // read as mounted volumes.
      return Number(id) === 0 ? "folder" : "folder";
    },
    toSource(id) {
      this.$router.push({ path: `/files/${id}/` });
      this.closeHovers();
    },
    abortOngoingFetchUsage() {
      this.usageAbortController.abort();
    },
    async fetchUsage() {
      const path = this.$route.path.endsWith("/")
        ? this.$route.path
        : this.$route.path + "/";
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
    toRoot() {
      const sourceStore = useSourceStore();
      this.$router.push({ path: `${sourceStore.filesBase}/` });
      this.closeHovers();
    },
    openSettingsModal() {
      this.openSettings();
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
  },
  watch: {
    $route: {
      handler(to) {
        if (to.path.includes("/files")) {
          this.fetchUsage();
        }
        this.closeHovers();
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

.fb-sidebar-storage-pct {
  font-size: 11px;
  color: var(--dim);
}

.fb-sidebar-storage-detail {
  margin-top: 6px;
  font-size: 11.5px;
  color: var(--dim);
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

.fb-sidebar-avatar {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: var(--accent);
  color: var(--on-accent);
  font-size: 13px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  user-select: none;
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
}
</style>
