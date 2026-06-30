<template>
  <div>
    <header-bar>
      <span class="fb-toolbar-title">{{ $t("sidebar.recent") }}</span>
    </header-bar>

    <div v-if="items.length === 0" class="fb-empty">
      <div class="fb-empty-icon">
        <fb-icon name="clock" size="40px" />
      </div>
      <h2 class="fb-empty-title">{{ $t("files.emptyRecentTitle") }}</h2>
      <p class="fb-empty-sub">{{ $t("files.emptyRecentSub") }}</p>
    </div>

    <div v-else id="listing" class="file-icons" :class="viewMode">
      <div>
        <div class="fb-col-header">
          <div>
            <p class="name">{{ $t("files.name") }}</p>
            <p class="size">{{ $t("files.size") }}</p>
            <p class="modified">{{ $t("files.lastModified") }}</p>
          </div>
        </div>
      </div>

      <h2 v-if="dirs.length > 0">{{ $t("files.folders") }}</h2>
      <div v-if="dirs.length > 0" class="fb-items fb-items--folders">
        <item
          v-for="(it, i) in dirs"
          :key="it.url"
          :index="i"
          :name="it.name"
          :isDir="true"
          :url="it.url"
          :type="it.type"
          :size="0"
          :modified="toISO(it.at)"
          :path="toPath(it.url)"
        />
      </div>

      <h2 v-if="files.length > 0">{{ $t("files.files") }}</h2>
      <div v-if="files.length > 0" class="fb-items fb-items--files">
        <item
          v-for="(it, i) in files"
          :key="it.url"
          :index="dirs.length + i"
          :name="it.name"
          :isDir="false"
          :url="it.url"
          :type="it.type"
          :size="0"
          :modified="toISO(it.at)"
          :path="toPath(it.url)"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useAuthStore } from "@/stores/auth";
import { getRecents, type RecentFile } from "@/utils/recents";
import { removePrefix } from "@/api/utils";
import { users } from "@/api";
import HeaderBar from "@/components/header/HeaderBar.vue";
import FbIcon from "@/components/FbIcon.vue";
import Item from "@/components/files/ListingItem.vue";

const authStore = useAuthStore();
const items = ref<RecentFile[]>([]);

onMounted(async () => {
  // Show localStorage immediately, then sync with backend for cross-device data
  items.value = getRecents();
  try {
    const fresh = await users.get(authStore.user!.id);
    if (Array.isArray(fresh.recents) && fresh.recents.length >= 0) {
      authStore.updateUser({ recents: fresh.recents });
      localStorage.setItem("fb-recent-files", JSON.stringify(fresh.recents));
      items.value = fresh.recents as RecentFile[];
    }
  } catch {
    // backend unavailable — keep localStorage data
  }
});

const viewMode = computed(() => {
  const m = authStore.user?.viewMode ?? "list";
  return m === "mosaic gallery" ? "mosaic" : m;
});

const dirs = computed(() => items.value.filter((i) => i.type === "dir"));
const files = computed(() => items.value.filter((i) => i.type !== "dir"));

const toISO = (ms: number) => new Date(ms).toISOString();
const toPath = (url: string) => removePrefix(url);
</script>
