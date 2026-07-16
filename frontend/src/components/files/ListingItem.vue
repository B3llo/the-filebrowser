<template>
  <div
    class="item"
    role="button"
    tabindex="0"
    :draggable="isDraggable"
    @dragstart="dragStart"
    @dragover="dragOver"
    @drop="drop"
    @click="itemClick"
    @mousedown="handleMouseDown"
    @mouseup="handleMouseUp"
    @mouseleave="handleMouseLeave"
    @touchstart="handleTouchStart"
    @touchend="handleTouchEnd"
    @touchcancel="handleTouchCancel"
    @touchmove="handleTouchMove"
    :class="folderColorClass"
    :data-dir="isDir"
    :data-type="type"
    :data-kind="kind"
    :data-thumb="
      hasImageThumb || hasVideoThumb || hasPdfThumb ? 'true' : 'false'
    "
    :data-dotfile="isDotfile ? 'true' : 'false'"
    :data-path="path"
    :aria-label="name"
    :aria-selected="isSelected"
    :data-ext="getExtension(name).toLowerCase()"
    @contextmenu="contextMenu"
  >
    <div>
      <img v-if="hasImageThumb" ref="thumbImgRef" />
      <canvas
        v-else-if="hasVideoThumb"
        ref="videoCanvas"
        class="fb-video-thumb"
      ></canvas>
      <canvas
        v-else-if="hasPdfThumb"
        ref="pdfCanvas"
        class="fb-pdf-thumb"
      ></canvas>
      <svg
        v-else-if="isDir"
        class="fb-folder-icon"
        viewBox="0 0 24 24"
        fill="var(--folder-fill)"
        stroke="var(--folder-stroke)"
        stroke-width="1.6"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path
          d="M20 19a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2h-7.6a1 1 0 0 1-.8-.4L10.3 4.9a1 1 0 0 0-.8-.4H4a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2Z"
        />
      </svg>
      <i v-else class="material-icons"></i>

      <!-- Hidden video element for frame capture -->
      <video
        v-if="hasVideoThumb"
        ref="videoEl"
        :src="thumbnailUrl + '#t=0.1'"
        preload="metadata"
        muted
        class="fb-video-hidden"
        @loadeddata="captureFrame"
      ></video>

      <!-- Play icon overlay for video thumbnails -->
      <div v-if="isVideoItem" class="fb-video-play-overlay" aria-hidden="true">
        <svg viewBox="0 0 24 24" fill="rgba(255,255,255,0.9)">
          <path d="M8 5v14l11-7z" />
        </svg>
      </div>

      <!-- Mosaic card: rendered markdown preview -->
      <div
        v-if="hasMarkdownPreview"
        class="fb-card-markdown markdown-body"
        aria-hidden="true"
        v-html="renderedMarkdown"
      ></div>

      <!-- Mosaic card: inline preview for code/text files -->
      <pre
        v-else-if="hasPreview"
        class="fb-card-preview"
        aria-hidden="true"
      ><code>{{ previewText }}</code></pre>

      <!-- Mosaic card: document outline + extension label for non-thumb files -->
      <div v-else class="fb-card-doc" aria-hidden="true">
        <svg viewBox="0 0 56 70">
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
        <span class="fb-card-doc-ext">{{ extLabelText }}</span>
      </div>

      <!-- Mosaic card: Office document thumbnail (lazy fetched) -->
      <template v-if="officeContent">
        <div
          class="fb-card-office"
          :class="`fb-card-office--${officeContent.kind}`"
          aria-hidden="true"
        >
          <pre
            v-if="officeContent.kind === 'docx'"
            class="fb-card-office-text"
            >{{ officeContent.text }}</pre
          >
          <table v-else class="fb-card-office-table">
            <tr v-for="(row, ri) in officeContent.grid" :key="ri">
              <td v-for="(cell, ci) in row" :key="ci">{{ cell }}</td>
            </tr>
          </table>
        </div>
      </template>

      <!-- List view: extension pill badge (colored by kind) -->
      <span class="fb-list-pill" aria-hidden="true">{{ extLabelText }}</span>
    </div>

    <div>
      <p class="name">{{ name }}</p>

      <p v-if="!hideSize && isDir" class="size" data-order="-1">&mdash;</p>
      <p v-else-if="!hideSize" class="size" :data-order="humanSize()">
        {{ humanSize() }}
      </p>

      <p class="modified">
        <time :datetime="modified">{{ humanTime() }}</time>
      </p>
    </div>

    <span v-if="isItemStarred" class="fb-star-badge" aria-hidden="true">
      <svg
        viewBox="0 0 24 24"
        width="11"
        height="11"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        aria-hidden="true"
      >
        <path
          d="M11.48 3.5a.6.6 0 0 1 1.04 0l2.36 4.78 5.27.77a.6.6 0 0 1 .33 1.02l-3.81 3.72.9 5.25a.6.6 0 0 1-.87.63L12 17.5l-4.71 2.47a.6.6 0 0 1-.87-.63l.9-5.25-3.81-3.72a.6.6 0 0 1 .33-1.02l5.27-.77z"
        />
      </svg>
    </span>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { isStarred, starVersion } from "@/utils/starred";

import { enableThumbs, enableVideoThumbs } from "@/utils/constants";
import { filesize } from "@/utils";
import { fileKind, extLabel } from "@/utils/fileKind";
import dayjs from "dayjs";
import { files as api } from "@/api";
import { createURL } from "@/api/utils";
import * as upload from "@/utils/upload";
import { computed, inject, onMounted, onUnmounted, ref, watch } from "vue";
import { loadThumbnail } from "@/utils/thumbnailCache";
import {
  officeThumbKind,
  fetchDocxText,
  fetchSheetGrid,
} from "@/utils/officeThumb";
import { useRouter } from "vue-router";
import { marked } from "marked";
import DOMPurify from "dompurify";
import * as pdfjsLib from "pdfjs-dist";

pdfjsLib.GlobalWorkerOptions.workerSrc = new URL(
  "pdfjs-dist/build/pdf.worker.min.mjs",
  import.meta.url
).toString();

const touches = ref<number>(0);

const longPressTimer = ref<number | null>(null);
const longPressTriggered = ref<boolean>(false);
const longPressDelay = ref<number>(500);
const startPosition = ref<{ x: number; y: number } | null>(null);
const moveThreshold = ref<number>(10);

const $showError = inject<IToastError>("$showError")!;
const router = useRouter();

const props = defineProps<{
  name: string;
  isDir: boolean;
  url: string;
  type: string;
  size: number;
  modified: string;
  index: number;
  readOnly?: boolean;
  noOpen?: boolean;
  onOpen?: (url: string, isDir: boolean) => void;
  path?: string;
  preview?: string;
  hideSize?: boolean;
}>();

const authStore = useAuthStore();
const fileStore = useFileStore();
const layoutStore = useLayoutStore();

const singleClick = computed(
  () => !props.readOnly && authStore.user?.singleClick
);
const isSelected = computed(
  () => fileStore.selected.indexOf(props.index) !== -1
);
const isDraggable = computed(
  () => !props.readOnly && authStore.user?.perm.rename
);

const folderColorClass = computed(() => {
  if (!props.isDir || !props.path) return "";
  const color = authStore.user?.folderColors?.[props.path];
  return color ? `folder-color-${color}` : "";
});

const isItemStarred = computed(() => {
  starVersion.value; // reactive dependency
  return isStarred(props.url);
});

const canDrop = computed(() => {
  if (!props.isDir || props.readOnly) return false;

  for (const i of fileStore.selected) {
    if (fileStore.req?.items[i].url === props.url) {
      return false;
    }
  }

  return true;
});

const sourceId = computed(() => {
  const parts = props.url.split("/");
  if (parts[1] === "files" && parts.length >= 3) {
    return parts[2];
  }
  return undefined;
});

const thumbnailUrl = computed(() => {
  const file = {
    path: props.path,
    modified: props.modified,
  };

  return api.getPreviewURL(file as Resource, "thumb", sourceId.value);
});

// Thumbnail lazy-load with concurrency limiting
const thumbImgRef = ref<HTMLImageElement | null>(null);
let thumbObserver: IntersectionObserver | null = null;

onMounted(() => {
  if (!thumbImgRef.value) return;
  thumbObserver = new IntersectionObserver(
    (entries) => {
      if (entries[0].isIntersecting) {
        loadThumbnail(thumbImgRef.value!, thumbnailUrl.value);
        thumbObserver?.disconnect();
        thumbObserver = null;
      }
    },
    { rootMargin: "200px" }
  );
  thumbObserver.observe(thumbImgRef.value);
});

// Office thumbnail lazy load. Fetches in the background for every visible
// item (same pattern as PDF page thumbnails) with a size cap to keep it light.
onMounted(async () => {
  const kind = hasOfficeThumb.value;
  if (!kind || !officeUrl.value) return;
  if (kind === "docx") {
    const text = await fetchDocxText(officeUrl.value);
    if (text) officeContent.value = { kind: "docx", text };
  } else if (kind === "sheet") {
    const grid = await fetchSheetGrid(officeUrl.value);
    if (grid) officeContent.value = { kind: "sheet", grid };
  }
});

onUnmounted(() => {
  thumbObserver?.disconnect();
  thumbObserver = null;
});

// Reload if the file is modified (url key changes)
watch(thumbnailUrl, (url) => {
  if (thumbImgRef.value && url) loadThumbnail(thumbImgRef.value, url);
});

const isThumbsEnabled = computed(() => {
  return enableThumbs;
});

const kind = computed(() =>
  fileKind({ isDir: props.isDir, type: props.type, name: props.name })
);

// Server-generated static JPEG thumbnail for video (reuses the same <img>
// lazy-load path as images). Falls back to client-side capture below when
// the server didn't generate one (flag off or ffmpeg unavailable).
const hasServerVideoThumb = computed(
  () =>
    !props.readOnly &&
    props.type === "video" &&
    isThumbsEnabled.value &&
    enableVideoThumbs
);

const hasImageThumb = computed(
  () =>
    !props.readOnly &&
    isThumbsEnabled.value &&
    (props.type === "image" || hasServerVideoThumb.value)
);

const isVideoItem = computed(
  () => !props.readOnly && props.type === "video" && isThumbsEnabled.value
);

// Client-side <video>+canvas capture fallback, only used when the server
// isn't providing a static thumbnail.
const hasVideoThumb = computed(() => isVideoItem.value && !enableVideoThumbs);

const hasPdfThumb = computed(
  () => !props.readOnly && props.type === "pdf" && isThumbsEnabled.value
);

// Client-side Office thumbnail for the grid card.
const officeContent = ref<
  { kind: "docx"; text: string } | { kind: "sheet"; grid: string[][] } | null
>(null);

const hasOfficeThumb = computed(() => {
  if (
    props.readOnly ||
    !isThumbsEnabled.value ||
    hasImageThumb.value ||
    hasVideoThumb.value ||
    hasPdfThumb.value
  )
    return false;
  return officeThumbKind(props.name, props.size);
});

const officeUrl = computed(() => {
  if (!props.path) return "";
  return createURL("api/raw" + props.path, { inline: "true" });
});

const isDotfile = computed(() => {
  const n = props.name;
  return n.length > 0 && n[0] === ".";
});

const extLabelText = computed(() => extLabel(props.name));

const hasPreview = computed(() => {
  if (props.isDir) return false;
  if (props.type !== "text" && props.type !== "textImmutable") return false;
  return !!props.preview && props.preview.trim().length > 0;
});

const previewText = computed(() => {
  if (!props.preview) return "";
  const lines = props.preview.split("\n").slice(0, 12).join("\n");
  return lines;
});

const isMarkdownFile = computed(() => {
  const lower = props.name.toLowerCase();
  return lower.endsWith(".md") || lower.endsWith(".markdown");
});

const hasMarkdownPreview = computed(() => {
  if (props.isDir) return false;
  if (!isMarkdownFile.value) return false;
  return !!props.preview && props.preview.trim().length > 0;
});

const renderedMarkdown = ref<string>("");

watch(
  () => props.preview,
  async (preview) => {
    if (!hasMarkdownPreview.value || !preview) {
      renderedMarkdown.value = "";
      return;
    }
    try {
      const stripped = preview.replace(/^---\n[\s\S]*?\n---\n?/, "");
      renderedMarkdown.value = DOMPurify.sanitize(await marked(stripped));
    } catch {
      renderedMarkdown.value = "";
    }
  },
  { immediate: true }
);

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

const pdfCanvas = ref<HTMLCanvasElement | null>(null);

watch(
  pdfCanvas,
  async (canvas) => {
    if (!canvas || !hasPdfThumb.value) return;
    try {
      const loadingTask = pdfjsLib.getDocument({ url: thumbnailUrl.value });
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

const humanSize = () => {
  return props.type == "invalid_link" ? "invalid link" : filesize(props.size);
};

const humanTime = () => {
  if (!props.readOnly && authStore.user?.dateFormat) {
    return dayjs(props.modified).format("L LT");
  }
  return dayjs(props.modified).fromNow();
};

const dragStart = () => {
  if (fileStore.selectedCount === 0) {
    fileStore.selected.push(props.index);
    return;
  }

  if (!isSelected.value) {
    fileStore.selected = [];
    fileStore.selected.push(props.index);
  }
};

const dragOver = (event: Event) => {
  if (!canDrop.value) return;

  event.preventDefault();
  let el = event.target as HTMLElement | null;
  if (el !== null) {
    for (let i = 0; i < 5; i++) {
      if (!el?.classList.contains("item")) {
        el = el?.parentElement ?? null;
      }
    }

    if (el !== null) el.style.opacity = "1";
  }
};

const drop = async (event: Event) => {
  if (!canDrop.value) return;
  event.preventDefault();

  if (fileStore.selectedCount === 0) return;

  let el = event.target as HTMLElement | null;
  for (let i = 0; i < 5; i++) {
    if (el !== null && !el.classList.contains("item")) {
      el = el.parentElement;
    }
  }

  const items: any[] = [];

  for (const i of fileStore.selected) {
    if (fileStore.req) {
      items.push({
        from: fileStore.req?.items[i].url,
        to: props.url + encodeURIComponent(fileStore.req?.items[i].name),
        name: fileStore.req?.items[i].name,
        size: fileStore.req?.items[i].size,
        isDir: fileStore.req?.items[i].isDir,
        modified: fileStore.req?.items[i].modified,
        overwrite: false,
        rename: false,
      });
    }
  }

  // Get url from ListingItem instance
  if (el === null) {
    return;
  }
  const path = el.__vue__.url;

  const action = (overwrite?: boolean, rename?: boolean) => {
    const action =
      (event as KeyboardEvent).ctrlKey || (event as KeyboardEvent).metaKey
        ? api.copy
        : api.move;
    action(items, overwrite, rename)
      .then(() => {
        fileStore.reload = true;
      })
      .catch($showError);
  };

  const conflict = await upload.checkConflict(items, path, true);

  if (conflict.length > 0) {
    layoutStore.showHover({
      prompt: "resolve-conflict",
      props: {
        conflict: conflict,
      },
      confirm: (event: Event, result: Array<ConflictingResource>) => {
        event.preventDefault();
        layoutStore.closeHovers();
        for (let i = result.length - 1; i >= 0; i--) {
          const item = result[i];
          if (item.checked.length == 2) {
            items[item.index].rename = true;
          } else if (item.checked.length == 1 && item.checked[0] == "origin") {
            items[item.index].overwrite = true;
          } else {
            items.splice(item.index, 1);
          }
        }
        if (items.length > 0) {
          action();
        }
      },
    });

    return;
  }

  action(false, false);
};

const itemClick = (event: Event | KeyboardEvent) => {
  // If long press was triggered, prevent normal click behavior
  if (longPressTriggered.value) {
    longPressTriggered.value = false;
    return;
  }

  if (
    singleClick.value &&
    !(event as KeyboardEvent).ctrlKey &&
    !(event as KeyboardEvent).metaKey &&
    !(event as KeyboardEvent).shiftKey &&
    !fileStore.multiple
  )
    open();
  else click(event);
};

const contextMenu = (event: MouseEvent) => {
  event.preventDefault();
  if (
    fileStore.selected.length === 0 ||
    event.ctrlKey ||
    fileStore.selected.indexOf(props.index) === -1
  ) {
    click(event);
  }
};

const click = (event: Event | KeyboardEvent) => {
  if (!singleClick.value && fileStore.selectedCount !== 0)
    event.preventDefault();

  setTimeout(() => {
    touches.value = 0;
  }, 300);

  touches.value++;
  if (touches.value > 1) {
    open();
  }

  if (fileStore.selected.indexOf(props.index) !== -1) {
    if (
      (event as KeyboardEvent).ctrlKey ||
      (event as KeyboardEvent).metaKey ||
      fileStore.multiple
    ) {
      fileStore.removeSelected(props.index);
    } else {
      fileStore.selected = [props.index];
    }
    return;
  }

  if ((event as KeyboardEvent).shiftKey && fileStore.selected.length > 0) {
    let fi = 0;
    let la = 0;

    if (props.index > fileStore.selected[0]) {
      fi = fileStore.selected[0] + 1;
      la = props.index;
    } else {
      fi = props.index;
      la = fileStore.selected[0] - 1;
    }

    for (; fi <= la; fi++) {
      if (fileStore.selected.indexOf(fi) == -1) {
        fileStore.selected.push(fi);
      }
    }

    return;
  }

  if (
    !(event as KeyboardEvent).ctrlKey &&
    !(event as KeyboardEvent).metaKey &&
    !fileStore.multiple
  ) {
    fileStore.selected = [];
  }
  fileStore.selected.push(props.index);
};

const open = () => {
  if (props.noOpen) return;
  if (props.onOpen) {
    props.onOpen(props.url, props.isDir);
  } else {
    router.push({ path: props.url });
  }
};

const getExtension = (fileName: string): string => {
  const lastDotIndex = fileName.lastIndexOf(".");
  if (lastDotIndex === -1) {
    return fileName;
  }
  return fileName.substring(lastDotIndex);
};

// Long-press helper functions
const startLongPress = (clientX: number, clientY: number) => {
  startPosition.value = { x: clientX, y: clientY };
  longPressTimer.value = window.setTimeout(() => {
    handleLongPress();
  }, longPressDelay.value);
};

const cancelLongPress = () => {
  if (longPressTimer.value !== null) {
    window.clearTimeout(longPressTimer.value);
    longPressTimer.value = null;
  }
  startPosition.value = null;
};

const handleLongPress = () => {
  if (singleClick.value) {
    longPressTriggered.value = true;
    click(new Event("longpress"));
  }
  cancelLongPress();
};

const checkMovement = (clientX: number, clientY: number): boolean => {
  if (!startPosition.value) return false;

  const deltaX = Math.abs(clientX - startPosition.value.x);
  const deltaY = Math.abs(clientY - startPosition.value.y);

  return deltaX > moveThreshold.value || deltaY > moveThreshold.value;
};

// Event handlers
const handleMouseDown = (event: MouseEvent) => {
  if (event.button === 0) {
    startLongPress(event.clientX, event.clientY);
  }
};

const handleMouseUp = () => {
  cancelLongPress();
};

const handleMouseLeave = () => {
  cancelLongPress();
};

const handleTouchStart = (event: TouchEvent) => {
  if (event.touches.length === 1) {
    const touch = event.touches[0];
    startLongPress(touch.clientX, touch.clientY);
  }
};

const handleTouchEnd = () => {
  cancelLongPress();
};

const handleTouchCancel = () => {
  cancelLongPress();
};

const handleTouchMove = (event: TouchEvent) => {
  if (event.touches.length === 1 && startPosition.value) {
    const touch = event.touches[0];
    if (checkMovement(touch.clientX, touch.clientY)) {
      cancelLongPress();
    }
  }
};
</script>
