<template>
  <div id="fb-layout">
    <div v-if="uploadStore.totalBytes" class="progress">
      <div
        v-bind:style="{
          width: sentPercent + '%',
        }"
      ></div>
    </div>
    <sidebar></sidebar>
    <main>
      <router-view></router-view>
      <shell
        v-if="
          enableExec && authStore.isLoggedIn && authStore.user?.perm.execute
        "
      />
    </main>
    <prompts></prompts>
    <upload-files></upload-files>
    <settings-modal></settings-modal>
    <command-palette></command-palette>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import { useFileStore } from "@/stores/file";
import { useUploadStore } from "@/stores/upload";
import Sidebar from "@/components/Sidebar.vue";
import Prompts from "@/components/prompts/Prompts.vue";
import Shell from "@/components/Shell.vue";
import UploadFiles from "@/components/prompts/UploadFiles.vue";
import SettingsModal from "@/components/SettingsModal.vue";
import CommandPalette from "@/components/CommandPalette.vue";
import { enableExec } from "@/utils/constants";
import { computed, onBeforeUnmount, onMounted, watch } from "vue";
import { useRoute } from "vue-router";

const layoutStore = useLayoutStore();
const authStore = useAuthStore();
const fileStore = useFileStore();
const uploadStore = useUploadStore();
const route = useRoute();

const sentPercent = computed(() =>
  ((uploadStore.sentBytes / uploadStore.totalBytes) * 100).toFixed(2)
);

// Global Cmd/Ctrl+K toggles the command palette (#33). When the palette is
// open it handles the shortcut itself (and stops propagation), so this only
// fires to open it.
const onGlobalKeydown = (e: KeyboardEvent) => {
  if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
    e.preventDefault();
    layoutStore.toggleCommandPalette();
  }
};

onMounted(() => window.addEventListener("keydown", onGlobalKeydown));
onBeforeUnmount(() => window.removeEventListener("keydown", onGlobalKeydown));

watch(route, () => {
  fileStore.selected = [];
  fileStore.multiple = false;
  if (layoutStore.currentPromptName !== "success") {
    layoutStore.closeHovers();
  }
  // Close the palette on navigation so it never lingers over a new view.
  layoutStore.closeCommandPalette();
});
</script>
