<template>
  <div class="sheet-preview fb-scroll">
    <div v-if="sheets.length > 1" class="sheet-tabs">
      <button
        v-for="(s, i) in sheets"
        :key="i"
        class="sheet-tab"
        :class="{ active: i === activeTab }"
        @click="activeTab = i"
      >
        {{ s.name }}
      </button>
    </div>

    <div class="sheet-content">
      <div v-if="error" class="sheet-placeholder">
        <FbIcon name="file" size="56px" />
        <p class="sheet-placeholder-title">
          {{
            t("files.officeLoadFailed", "Unable to load spreadsheet preview")
          }}
        </p>
        <a
          v-if="downloadUrl"
          :href="downloadUrl"
          class="sheet-download-btn"
          target="_blank"
          rel="noopener"
        >
          {{ t("buttons.download") }}
        </a>
      </div>

      <div v-else-if="loading" class="sheet-loading">
        <div class="sheet-loading-spinner" aria-hidden="true"></div>
        <span>{{ t("files.loadingPreview", "Loading preview…") }}</span>
      </div>

      <div v-else class="sheet-table" v-html="sheets[activeTab]?.html"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onBeforeUnmount } from "vue";
import { useI18n } from "vue-i18n";
import DOMPurify from "dompurify";
import FbIcon from "@/components/FbIcon.vue";

interface Sheet {
  name: string;
  html: string;
}

const props = withDefaults(
  defineProps<{
    src?: string;
    filename?: string;
    downloadUrl?: string;
  }>(),
  { src: "", filename: "spreadsheet.xlsx", downloadUrl: "" }
);

const { t } = useI18n();

const loading = ref(true);
const error = ref(false);
const sheets = ref<Sheet[]>([]);
const activeTab = ref(0);

const load = async () => {
  loading.value = true;
  error.value = false;
  sheets.value = [];
  activeTab.value = 0;

  if (!props.src) {
    error.value = true;
    loading.value = false;
    return;
  }

  try {
    const res = await fetch(props.src);
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const arrayBuffer = await res.arrayBuffer();

    // Loaded on demand so the main bundle stays lightweight (#51).
    const mod: any = await import("xlsx");
    const XLSX = mod.default ?? mod;
    const wb = XLSX.read(arrayBuffer, { type: "array" });
    sheets.value = (wb.SheetNames as string[]).map((name) => ({
      name,
      html: DOMPurify.sanitize(
        XLSX.utils.sheet_to_html(wb.Sheets[name], { editable: false })
      ),
    }));
    loading.value = false;
  } catch (e) {
    console.error("Failed to load spreadsheet preview:", e);
    error.value = true;
    loading.value = false;
  }
};

watch(() => props.src, load, { immediate: true });

onBeforeUnmount(() => {
  sheets.value = [];
});
</script>
