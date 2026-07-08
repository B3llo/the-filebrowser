<template>
  <div class="pdf-preview fb-scroll">
    <div class="pdf-content">
      <!-- Error / fallback: offer a direct download -->
      <div v-if="error" class="pdf-placeholder">
        <FbIcon name="file" size="56px" />
        <p class="pdf-placeholder-title">{{ t("files.previewUnavailable", "Unable to load PDF preview") }}</p>
        <a v-if="downloadUrl" :href="downloadUrl" class="pdf-download-btn" target="_blank" rel="noopener">
          {{ t("buttons.download") }}
        </a>
      </div>

      <!-- Loading state -->
      <div v-else-if="loading" class="pdf-loading">
        <div class="pdf-loading-spinner" aria-hidden="true"></div>
        <span>{{ t("files.loadingPreview", "Loading preview…") }}</span>
      </div>

      <!-- Rendered pages -->
      <div v-else ref="pagesRef" class="pdf-pages">
        <canvas
          v-for="(canvas, i) in pageCanvases"
          :key="i"
          :ref="(el) => setCanvasRef(el as HTMLCanvasElement | null, i)"
          class="pdf-page-canvas"
        ></canvas>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onBeforeUnmount, nextTick } from "vue";
import * as pdfjsLib from "pdfjs-dist";
import type { PDFDocumentProxy, PDFDocumentLoadingTask, RenderTask } from "pdfjs-dist";
import { useI18n } from "vue-i18n";
import FbIcon from "@/components/FbIcon.vue";

pdfjsLib.GlobalWorkerOptions.workerSrc = new URL(
  "pdfjs-dist/build/pdf.worker.min.mjs",
  import.meta.url
).toString();

const props = withDefaults(
  defineProps<{
    src?: string;
    filename?: string;
    downloadUrl?: string;
  }>(),
  { src: "", filename: "document.pdf", downloadUrl: "" }
);

const { t } = useI18n();

const loading = ref(true);
const error = ref(false);
const pageCanvases = ref<number[]>([]);
const pagesRef = ref<HTMLElement | null>(null);

// Keep a live map of the canvas elements so the render loop can address them
// after Vue has stamped them into the DOM.
const canvasEls = new Map<number, HTMLCanvasElement | null>();
const setCanvasRef = (el: HTMLCanvasElement | null, i: number) => {
  if (el) canvasEls.set(i, el);
  else canvasEls.delete(i);
};

let pdfDoc: PDFDocumentProxy | null = null;
let loadingTask: PDFDocumentLoadingTask | null = null;
const renderTasks: RenderTask[] = [];

const cancelRender = () => {
  for (const task of renderTasks) {
    try {
      task.cancel();
    } catch {
      /* already finished */
    }
  }
  renderTasks.length = 0;
};

const cleanup = () => {
  cancelRender();
  loadingTask?.destroy().catch(() => {});
  loadingTask = null;
  pdfDoc = null;
  canvasEls.clear();
};

const renderAllPages = async () => {
  if (!pdfDoc || !pagesRef.value) return;
  const container = pagesRef.value;
  const containerWidth = Math.max(container.clientWidth, 280);
  const dpr = Math.min(window.devicePixelRatio || 1, 2);

  for (let i = 1; i <= pdfDoc.numPages; i++) {
    const page = await pdfDoc.getPage(i);
    const viewport1 = page.getViewport({ scale: 1 });
    const scale = containerWidth / viewport1.width;
    const viewport = page.getViewport({ scale });

    const canvas = canvasEls.get(i - 1);
    if (!canvas) continue;

    canvas.width = Math.floor(viewport.width * dpr);
    canvas.height = Math.floor(viewport.height * dpr);
    canvas.style.width = `${Math.floor(viewport.width)}px`;
    canvas.style.height = `${Math.floor(viewport.height)}px`;

    const ctx = canvas.getContext("2d");
    if (!ctx) continue;
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);

    const task = page.render({ canvasContext: ctx, viewport, canvas });
    renderTasks.push(task);
    try {
      await task.promise;
    } catch (err: any) {
      // Cancelled renders are expected when the component unmounts mid-render.
      if (err?.name !== "RenderingCancelledException") {
        throw err;
      }
      return;
    }
  }
};

const load = async () => {
  cleanup();
  if (!props.src) {
    error.value = true;
    loading.value = false;
    return;
  }
  loading.value = true;
  error.value = false;
  try {
    loadingTask = pdfjsLib.getDocument({ url: props.src });
    pdfDoc = await loadingTask.promise;
    pageCanvases.value = Array.from({ length: pdfDoc.numPages }, (_, i) => i);
    loading.value = false;
    // Wait for Vue to mount the canvas elements before rendering into them.
    await nextTick();
    await renderAllPages();
  } catch (e) {
    console.error("Failed to load PDF preview:", e);
    error.value = true;
    loading.value = false;
  }
};

watch(
  () => props.src,
  () => {
    load();
  },
  { immediate: true }
);

// Re-render on viewport resize so pages stay crisp and fit the width.
let resizeRaf = 0;
const onResize = () => {
  cancelAnimationFrame(resizeRaf);
  resizeRaf = requestAnimationFrame(() => {
    if (!loading.value && !error.value) renderAllPages();
  });
};
window.addEventListener("resize", onResize);

onBeforeUnmount(() => {
  window.removeEventListener("resize", onResize);
  cancelAnimationFrame(resizeRaf);
  cleanup();
});
</script>

<style scoped>
.pdf-preview {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--pdf-bg, #525659);
  overflow: auto;
}

.pdf-content {
  flex: 1;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  min-height: 0;
}

.pdf-pages {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 16px;
  width: 100%;
  max-width: 900px;
}

.pdf-page-canvas {
  display: block;
  max-width: 100%;
  background: #fff;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.25);
  border-radius: 2px;
}

.pdf-loading,
.pdf-placeholder {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
  color: rgba(255, 255, 255, 0.85);
  padding: 32px;
  text-align: center;
}

.pdf-loading-spinner {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  border: 3px solid rgba(255, 255, 255, 0.25);
  border-top-color: #fff;
  animation: fbspin 0.9s linear infinite;
}

.pdf-placeholder-title {
  margin: 0;
  font-size: 15px;
  font-weight: 500;
}

.pdf-download-btn {
  display: inline-block;
  padding: 10px 20px;
  background: var(--accent);
  color: var(--on-accent, #fff);
  text-decoration: none;
  border-radius: 9px;
  font-size: 14px;
  font-weight: 500;
}

.pdf-download-btn:hover {
  filter: brightness(1.05);
}
</style>
