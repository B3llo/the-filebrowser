<template>
  <div>
    <header-bar>
      <span class="fb-toolbar-title">{{ $t("sidebar.starred") }}</span>
    </header-bar>

    <div v-if="items.length === 0" class="fb-empty">
      <div class="fb-empty-icon">
        <fb-icon name="star" size="40px" />
      </div>
      <h2 class="fb-empty-title">{{ $t("files.emptyStarredTitle") }}</h2>
      <p class="fb-empty-sub">{{ $t("files.emptyStarredSub") }}</p>
    </div>

    <div v-else id="listing" class="file-icons no-size-col" :class="viewMode">
      <div>
        <div class="fb-col-header">
          <div>
            <p class="name">{{ $t("files.name") }}</p>
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
          :modified="toISO(it.starredAt)"
          :path="toPath(it.url)"
          hideSize
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
          :modified="toISO(it.starredAt)"
          :path="toPath(it.url)"
          hideSize
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { useAuthStore } from "@/stores/auth";
import { getStarred, starVersion, type StarredFile } from "@/utils/starred";
import { removePrefix } from "@/api/utils";
import { users } from "@/api";
import HeaderBar from "@/components/header/HeaderBar.vue";
import FbIcon from "@/components/FbIcon.vue";
import Item from "@/components/files/ListingItem.vue";

const authStore = useAuthStore();
const items = ref<StarredFile[]>([]);

const loadItems = () => {
  items.value = getStarred();
};

onMounted(async () => {
  // Show localStorage immediately, then sync with backend for cross-device data
  loadItems();
  try {
    const fresh = await users.get(authStore.user!.id);
    if (Array.isArray(fresh.starred) && fresh.starred.length >= 0) {
      authStore.updateUser({ starred: fresh.starred });
      localStorage.setItem("fb-starred-files", JSON.stringify(fresh.starred));
      items.value = fresh.starred as StarredFile[];
    }
  } catch {
    // backend unavailable — keep localStorage data
  }
});

// Re-sync the list whenever a star/unstar happens (e.g. via badge click)
watch(starVersion, loadItems);

const viewMode = computed(() => {
  const m = authStore.user?.viewMode ?? "list";
  return m === "mosaic gallery" ? "mosaic" : m;
});

const dirs = computed(() => items.value.filter((i) => i.type === "dir"));
const files = computed(() => items.value.filter((i) => i.type !== "dir"));

const toISO = (ms: number) => new Date(ms).toISOString();
const toPath = (url: string) => removePrefix(decodeURIComponent(url));
</script>

<style scoped>
#listing.no-size-col.list .item > div:last-of-type,
#listing.no-size-col.list .fb-col-header > div {
  grid-template-columns: 1fr 168px;
}
#listing.no-size-col.list .size {
  display: none;
}
</style>
