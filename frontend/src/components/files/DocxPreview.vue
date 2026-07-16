<template>
  <div class="doc-preview fb-scroll">
    <div class="doc-content">
      <div v-if="error" class="doc-placeholder">
        <FbIcon name="file" size="56px" />
        <p class="doc-placeholder-title">
          {{ t("files.officeLoadFailed", "Unable to load document preview") }}
        </p>
        <a
          v-if="downloadUrl"
          :href="downloadUrl"
          class="doc-download-btn"
          target="_blank"
          rel="noopener"
        >
          {{ t("buttons.download") }}
        </a>
      </div>

      <div v-else-if="loading" class="doc-loading">
        <div class="doc-loading-spinner" aria-hidden="true"></div>
        <span>{{ t("files.loadingPreview", "Loading preview…") }}</span>
      </div>

      <div v-else class="doc-page" v-html="html"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onBeforeUnmount } from "vue";
import { useI18n } from "vue-i18n";
import DOMPurify from "dompurify";
import FbIcon from "@/components/FbIcon.vue";

const props = withDefaults(
  defineProps<{
    src?: string;
    filename?: string;
    downloadUrl?: string;
  }>(),
  { src: "", filename: "document.docx", downloadUrl: "" }
);

const { t } = useI18n();

const loading = ref(true);
const error = ref(false);
const html = ref("");

const load = async () => {
  loading.value = true;
  error.value = false;
  html.value = "";

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
    const mod: any = await import("mammoth");
    const mammoth = mod.default ?? mod;
    const result = await mammoth.convertToHtml({ arrayBuffer });
    html.value = DOMPurify.sanitize(result.value);
    loading.value = false;
  } catch (e) {
    console.error("Failed to load DOCX preview:", e);
    error.value = true;
    loading.value = false;
  }
};

watch(() => props.src, load, { immediate: true });

onBeforeUnmount(() => {
  html.value = "";
});
</script>
