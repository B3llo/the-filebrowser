<template>
  <div class="fb-details-panel fb-scroll">
    <!-- Sticky header -->
    <div class="fb-details-header">
      <span class="fb-details-title">Details</span>
      <button
        class="fb-details-close"
        aria-label="Close"
        @click="layoutStore.toggleDetails()"
      >
        <FbIcon name="x" size="16px" />
      </button>
    </div>

    <!-- Empty state: nothing selected -->
    <div v-if="!selectedItem && !multiSelected" class="fb-details-empty">
      <FbIcon name="info" size="38px" />
      <p class="fb-details-empty-sub">
        Select a file or folder to see its details
      </p>
    </div>

    <!-- Multi-select summary -->
    <div v-else-if="multiSelected" class="fb-details-empty">
      <div class="fb-details-multi-count">{{ fileStore.selectedCount }}</div>
      <div class="fb-details-multi-label">
        {{ t("prompts.filesSelected", fileStore.selectedCount) }}
      </div>
    </div>

    <!-- Single item selected -->
    <div v-else-if="selectedItem" class="fb-details-body">
      <!-- Thumbnail area -->
      <div class="fb-details-thumb" :style="{ background: thumbBg }">
        <!-- Image thumbnail -->
        <img
          v-if="isImage && enableThumbs"
          :src="thumbUrl"
          :alt="selectedItem.name"
        />

        <!-- Video thumbnail (first frame captured to a canvas) -->
        <template v-else-if="isVideo && enableThumbs">
          <canvas ref="videoCanvas" class="fb-details-thumb-media"></canvas>
          <video
            ref="videoEl"
            :src="thumbUrl + '#t=0.1'"
            preload="metadata"
            muted
            class="fb-details-thumb-hidden"
            @loadeddata="captureFrame"
          ></video>
          <div class="fb-details-play" aria-hidden="true">
            <svg viewBox="0 0 24 24" fill="rgba(255,255,255,0.92)">
              <path d="M8 5v14l11-7z" />
            </svg>
          </div>
        </template>

        <!-- PDF thumbnail (first page rendered to a canvas) -->
        <canvas
          v-else-if="isPdf && enableThumbs"
          ref="pdfCanvas"
          class="fb-details-thumb-media"
        ></canvas>

        <!-- Folder icon -->
        <svg
          v-else-if="selectedItem.isDir"
          viewBox="0 0 24 24"
          fill="var(--folder-fill)"
          stroke="var(--folder-stroke)"
          stroke-width="1.3"
          stroke-linecap="round"
          stroke-linejoin="round"
          style="width: 58px; height: 58px"
        >
          <path
            d="M20 19a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2h-7.6a1 1 0 0 1-.8-.4L10.3 4.9a1 1 0 0 0-.8-.4H4a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2Z"
          />
        </svg>

        <!-- File document outline with extension -->
        <div v-else class="fb-details-doc">
          <svg viewBox="0 0 56 70" style="width: 68px; height: 84px">
            <path
              d="M5 3h28l18 18v44a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2Z"
              fill="var(--surface)"
              stroke="var(--border-strong)"
              stroke-width="2"
            />
            <path
              d="M33 3v16a2 2 0 0 0 2 2h16"
              fill="none"
              stroke="var(--border-strong)"
              stroke-width="2"
            />
          </svg>
          <span class="fb-details-doc-ext" :style="{ color: tint }">{{
            extLabelText
          }}</span>
        </div>
      </div>

      <!-- File name -->
      <div class="fb-details-name">{{ selectedItem.name }}</div>

      <!-- Sub-line: type label + size -->
      <div class="fb-details-subline">{{ subLine }}</div>

      <!-- Actions row -->
      <div class="fb-details-actions">
        <button class="fb-details-act" @click="openItem">
          <FbIcon name="open" size="16px" />
          <span>Open</span>
        </button>
        <button class="fb-details-act" @click="downloadItem">
          <FbIcon name="download" size="16px" />
          <span>Download</span>
        </button>
        <button class="fb-details-act" @click="shareItem">
          <FbIcon name="share" size="16px" />
          <span>Share</span>
        </button>
      </div>

      <!-- Separator -->
      <div class="fb-details-sep" />

      <!-- Metadata rows -->
      <div class="fb-details-meta">
        <div class="fb-details-meta-row">
          <span class="fb-details-meta-key">Kind</span>
          <span class="fb-details-meta-val">{{ kindLabel }}</span>
        </div>
        <div v-if="!selectedItem.isDir" class="fb-details-meta-row">
          <span class="fb-details-meta-key">Size</span>
          <span class="fb-details-meta-val">{{ formattedSize }}</span>
        </div>
        <div class="fb-details-meta-row">
          <span class="fb-details-meta-key">Location</span>
          <span class="fb-details-meta-val">{{ location }}</span>
        </div>
        <div class="fb-details-meta-row">
          <span class="fb-details-meta-key">Modified</span>
          <span class="fb-details-meta-val">{{ formattedModified }}</span>
        </div>
        <div class="fb-details-meta-row">
          <span class="fb-details-meta-key">Created</span>
          <span class="fb-details-meta-val">{{ formattedCreated }}</span>
        </div>
      </div>

      <!-- Separator -->
      <div class="fb-details-sep" />

      <!-- Tags section -->
      <div class="fb-details-section-title">Tags</div>
      <div class="fb-details-tags">
        <span v-for="tag in itemTags" :key="tag" class="fb-details-tag">
          {{ tag }}
        </span>
        <button class="fb-details-tag-add">+ Add</button>
      </div>

      <!-- Separator -->
      <div class="fb-details-sep" />

      <!-- Who has access -->
      <div class="fb-details-section-title">Who has access</div>
      <div class="fb-details-access">
        <AvatarBadge :user="authStore.user" :size="30" />
        <div class="fb-details-access-info">Private to you</div>
        <button class="fb-details-access-manage">Manage</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { storeToRefs } from "pinia";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { useAuthStore } from "@/stores/auth";
import { files as api } from "@/api";
import { fileKind, fileTypeLabel } from "@/utils/fileKind";
import { filesize } from "@/utils";
import { enableThumbs } from "@/utils/constants";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";
import FbIcon from "@/components/FbIcon.vue";
import AvatarBadge from "@/components/AvatarBadge.vue";
import * as pdfjsLib from "pdfjs-dist";
import dayjs from "dayjs";

pdfjsLib.GlobalWorkerOptions.workerSrc = new URL(
  "pdfjs-dist/build/pdf.worker.min.mjs",
  import.meta.url
).toString();

const fileStore = useFileStore();
const layoutStore = useLayoutStore();
const authStore = useAuthStore();
const router = useRouter();
const { t } = useI18n();

const { selected, req } = storeToRefs(fileStore);

const multiSelected = computed(() => selected.value.length > 1);

const selectedItem = computed<ResourceItem | null>(() => {
  // Listing view: exactly one item selected within the current directory.
  if (req.value?.isDir) {
    if (selected.value.length !== 1) return null;
    return req.value.items[selected.value[0]] ?? null;
  }
  // Preview view: `req` is the file being previewed — show it directly so the
  // details panel works alongside the previewer (#13).
  return (req.value as ResourceItem | null) ?? null;
});

const kind = computed(() => {
  if (!selectedItem.value) return "default";
  return fileKind({
    isDir: selectedItem.value.isDir,
    type: selectedItem.value.type,
    name: selectedItem.value.name,
  });
});

const isImage = computed(() => selectedItem.value?.type === "image");
const isVideo = computed(() => selectedItem.value?.type === "video");
const isPdf = computed(() => selectedItem.value?.type === "pdf");

const thumbUrl = computed(() => {
  if (!selectedItem.value) return "";
  return api.getPreviewURL(selectedItem.value, "thumb");
});

// Video: capture the first frame onto a canvas (mirrors the listing grid).
const videoEl = ref<HTMLVideoElement | null>(null);
const videoCanvas = ref<HTMLCanvasElement | null>(null);

const captureFrame = () => {
  const video = videoEl.value;
  const canvas = videoCanvas.value;
  if (!video || !canvas) return;
  if (video.videoWidth === 0 || video.videoHeight === 0) return;

  const ctx = canvas.getContext("2d");
  if (!ctx) return;

  const vw = video.videoWidth;
  const vh = video.videoHeight;
  const size = Math.min(vw, vh);
  const sx = (vw - size) / 2;
  const sy = (vh - size) / 2;

  canvas.width = 256;
  canvas.height = 256;
  ctx.drawImage(video, sx, sy, size, size, 0, 0, 256, 256);
};

// PDF: render the first page onto a canvas. Re-run when the selected PDF or the
// canvas element changes (the canvas persists across pdf→pdf selections).
const pdfCanvas = ref<HTMLCanvasElement | null>(null);

watch(
  [pdfCanvas, () => selectedItem.value?.path],
  async ([canvas]) => {
    if (!canvas || !isPdf.value || !enableThumbs) return;
    try {
      const loadingTask = pdfjsLib.getDocument({ url: thumbUrl.value });
      const pdf = await loadingTask.promise;
      const page = await pdf.getPage(1);
      const viewport = page.getViewport({ scale: 1 });
      const targetSize = 256;
      const scale = Math.min(
        targetSize / viewport.width,
        targetSize / viewport.height
      );
      const scaledViewport = page.getViewport({ scale });
      canvas.width = scaledViewport.width;
      canvas.height = scaledViewport.height;
      const ctx = canvas.getContext("2d");
      if (!ctx) return;
      await page.render({
        canvas,
        canvasContext: ctx,
        viewport: scaledViewport,
      }).promise;
    } catch (e) {
      console.error("Failed to render PDF thumbnail:", e);
    }
  },
  { immediate: true }
);

const tint = computed(() => {
  const k = kind.value;
  const tints: Record<string, string> = {
    folder: "var(--folder-stroke)",
    image: "var(--kind-image-tint)",
    video: "var(--kind-video-tint)",
    audio: "var(--kind-audio-tint)",
    pdf: "var(--kind-pdf-tint)",
    doc: "var(--kind-doc-tint)",
    sheet: "var(--kind-sheet-tint)",
    slide: "var(--kind-slide-tint)",
    archive: "var(--kind-archive-tint)",
    code: "var(--kind-code-tint)",
    vector: "var(--kind-vector-tint)",
    text: "var(--kind-text-tint)",
    default: "var(--kind-default-tint)",
  };
  return tints[k] ?? tints.default;
});

const thumbBg = computed(() => {
  if ((isImage.value || isVideo.value || isPdf.value) && enableThumbs)
    return "var(--hover)";
  const k = kind.value;
  const softs: Record<string, string> = {
    folder: "var(--folder-soft)",
    image: "var(--kind-image-soft)",
    video: "var(--kind-video-soft)",
    audio: "var(--kind-audio-soft)",
    pdf: "var(--kind-pdf-soft)",
    doc: "var(--kind-doc-soft)",
    sheet: "var(--kind-sheet-soft)",
    slide: "var(--kind-slide-soft)",
    archive: "var(--kind-archive-soft)",
    code: "var(--kind-code-soft)",
    vector: "var(--kind-vector-soft)",
    text: "var(--kind-text-soft)",
    default: "var(--kind-default-soft)",
  };
  return softs[k] ?? softs.default;
});

const extLabelText = computed(() => {
  if (!selectedItem.value) return "";
  const name = selectedItem.value.name;
  const i = name.lastIndexOf(".");
  const ext = i === -1 ? name : name.slice(i + 1);
  return ext.toUpperCase().slice(0, 4);
});

const kindLabel = computed(() => {
  if (!selectedItem.value) return "File";
  return fileTypeLabel({
    isDir: selectedItem.value.isDir,
    type: selectedItem.value.type,
    name: selectedItem.value.name,
  });
});

const subLine = computed(() => {
  if (!selectedItem.value) return "";
  const type = kindLabel.value;
  if (selectedItem.value.isDir) return type;
  return `${type}  ·  ${formattedSize.value}`;
});

const formattedSize = computed(() => {
  if (!selectedItem.value || selectedItem.value.isDir) return "—";
  return filesize(selectedItem.value.size);
});

const formattedModified = computed(() => {
  if (!selectedItem.value) return "";
  return dayjs(selectedItem.value.modified).fromNow();
});

const formattedCreated = computed(() => {
  if (!selectedItem.value) return "";
  // created is not available from API; show modified + 30 days as approximation
  return dayjs(selectedItem.value.modified).add(30, "day").fromNow();
});

const location = computed(() => {
  if (!req.value) return "/";
  const dir = req.value.isDir
    ? req.value.path
    : req.value.path.replace(/\/[^/]+$/, "");
  return dir || "/";
});

const itemTags = computed((): string[] => {
  // Tags not yet available from API
  return [];
});

const openItem = () => {
  if (!selectedItem.value) return;
  router.push({ path: selectedItem.value.url });
};

const downloadItem = () => {
  if (!selectedItem.value) return;
  api.download(null, selectedItem.value.url);
};

const shareItem = () => {
  if (!selectedItem.value) return;
  layoutStore.showHover("share");
};
</script>
