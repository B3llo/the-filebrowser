<template>
  <div
    id="previewer"
    @touchmove="onTouchMove"
    @wheel="onWheel"
    @mousemove="toggleNavigation"
    @touchstart="toggleNavigation"
  >
    <header-bar
      v-if="
        isPdf || isEpub || isCsv || isMarkdown || isCode || isText || showNav
      "
    >
      <span class="fb-preview-type">{{ typeLabel }}</span>
      <title class="fb-preview-name">{{ name }}</title>

      <template #actions>
        <action
          :disabled="layoutStore.loading"
          v-if="isResizeEnabled && fileStore.req?.type === 'image'"
          :icon="fullSize ? 'photo_size_select_large' : 'hd'"
          @action="toggleSize"
        />
        <action
          :disabled="layoutStore.loading"
          v-if="
            (isMarkdown || isCode || isText || isCsv) &&
            authStore.user?.perm.modify
          "
          fb-icon="rename"
          size="18px"
          :label="t('buttons.editAsText')"
          @action="editAsText"
        />
        <action
          :disabled="layoutStore.loading"
          v-if="authStore.user?.perm.download"
          fb-icon="download"
          size="18px"
          :label="$t('buttons.download')"
          @action="download"
        />
        <action
          :disabled="layoutStore.loading"
          v-if="
            ['image', 'audio', 'video'].includes(fileStore.req?.type || '') &&
            authStore.user?.perm.download
          "
          fb-icon="external-link"
          size="18px"
          :label="t('buttons.openDirect')"
          @action="openDirect"
        />
        <action
          :disabled="layoutStore.loading"
          v-if="authStore.user?.perm.delete"
          fb-icon="delete"
          size="18px"
          :label="$t('buttons.delete')"
          @action="deleteFile"
          id="delete-button"
          danger
        />
        <action
          fb-icon="x"
          size="18px"
          :label="$t('buttons.close')"
          @action="close()"
        />
      </template>
    </header-bar>

    <div class="loading delayed" v-if="layoutStore.loading">
      <div class="spinner">
        <div class="bounce1"></div>
        <div class="bounce2"></div>
        <div class="bounce3"></div>
      </div>
    </div>
    <template v-else>
      <div class="preview">
        <div v-if="isEpub" class="epub-reader">
          <vue-reader
            :location="location"
            :url="previewUrl"
            :get-rendition="getRendition"
            :epubInitOptions="{
              requestCredentials: true,
            }"
            :epubOptions="{
              allowPopups: true,
            }"
            @update:location="locationChange"
          />
          <div class="size">
            <button
              @click="changeSize(Math.max(100, size - 10))"
              class="reader-button"
            >
              <i class="material-icons">remove</i>
            </button>
            <button
              @click="changeSize(Math.min(150, size + 10))"
              class="reader-button"
            >
              <i class="material-icons">add</i>
            </button>
            <span>{{ size }}%</span>
          </div>
        </div>
        <CsvViewer v-else-if="isCsv" :content="csvContent" :error="csvError" />
        <div
          v-else-if="isMarkdown"
          class="md_preview markdown-body"
          v-html="markdownContent"
        ></div>
        <ExtendedImage
          v-else-if="fileStore.req?.type == 'image'"
          :src="previewUrl"
        />
        <audio
          v-else-if="fileStore.req?.type == 'audio'"
          ref="player"
          :src="previewUrl"
          controls
          :autoplay="autoPlay"
          @play="autoPlay = true"
        ></audio>
        <VideoPlayer
          v-else-if="fileStore.req?.type == 'video'"
          ref="player"
          :source="previewUrl"
          :subtitles="subtitles"
          :options="videoOptions"
        >
        </VideoPlayer>
        <PdfPreview
          v-else-if="isPdf"
          :src="previewUrl"
          :filename="name"
          :download-url="downloadUrl"
        />
        <CodePreview
          v-else-if="isCode"
          :content="codeContent"
          :language="codeLanguage"
          :filename="name"
        />
        <TextPreview
          v-else-if="isText"
          :content="textContent"
          :filename="name"
        />
        <div v-else-if="fileStore.req?.type == 'blob'" class="info">
          <div class="title">
            <i class="material-icons">feedback</i>
            {{ $t("files.noPreview") }}
          </div>
          <div>
            <a target="_blank" :href="downloadUrl" class="button button--flat">
              <div>
                <i class="material-icons">file_download</i
                >{{ $t("buttons.download") }}
              </div>
            </a>
            <a
              target="_blank"
              :href="previewUrl"
              class="button button--flat"
              v-if="!fileStore.req?.isDir"
            >
              <div>
                <i class="material-icons">open_in_new</i
                >{{ $t("buttons.openFile") }}
              </div>
            </a>
          </div>
        </div>
      </div>
    </template>

    <button
      @click="prev"
      @mouseover="hoverNav = true"
      @mouseleave="hoverNav = false"
      :class="{ hidden: !hasPrevious || !showNav }"
      :aria-label="$t('buttons.previous')"
      :title="$t('buttons.previous')"
    >
      <i class="material-icons">chevron_left</i>
    </button>
    <button
      @click="next"
      @mouseover="hoverNav = true"
      @mouseleave="hoverNav = false"
      :class="{ hidden: !hasNext || !showNav }"
      :aria-label="$t('buttons.next')"
      :title="$t('buttons.next')"
    >
      <i class="material-icons">chevron_right</i>
    </button>
    <link rel="prefetch" :href="previousRaw" />
    <link rel="prefetch" :href="nextRaw" />
  </div>
</template>

<script setup lang="ts">
import { useStorage } from "@vueuse/core";
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";

import { files as api } from "@/api";
import { createURL } from "@/api/utils";
import { resizePreview } from "@/utils/constants";
import url from "@/utils/url";
import { throttle } from "lodash-es";
import HeaderBar from "@/components/header/HeaderBar.vue";
import Action from "@/components/header/Action.vue";
import ExtendedImage from "@/components/files/ExtendedImage.vue";
import VideoPlayer from "@/components/files/VideoPlayer.vue";
import CsvViewer from "@/components/files/CsvViewer.vue";
import CodePreview from "@/components/files/CodePreview.vue";
import TextPreview from "@/components/files/TextPreview.vue";
import PdfPreview from "@/components/files/PdfPreview.vue";
import { VueReader } from "vue-reader";
import { computed, inject, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import type { Rendition } from "epubjs";
import { getTheme } from "@/utils/theme";
import { useI18n } from "vue-i18n";
import { marked } from "marked";
import DOMPurify from "dompurify";

// CSV file size limit for preview (5MB)
// Prevents browser memory issues with large files
const CSV_MAX_SIZE = 5 * 1024 * 1024;

const location = useStorage("book-progress", 0, undefined, {
  serializer: {
    read: (v) => JSON.parse(v),
    write: (v) => JSON.stringify(v),
  },
});
const size = useStorage("book-size", 120, undefined, {
  serializer: {
    read: (v) => JSON.parse(v),
    write: (v) => JSON.stringify(v),
  },
});

const locationChange = (epubcifi: number) => {
  location.value = epubcifi;
};
let rendition: Rendition | null = null;
const changeSize = (val: number) => {
  size.value = val;
  rendition?.themes.fontSize(`${val}%`);
};

const getRendition = (_rendition: Rendition) => {
  rendition = _rendition;
  switch (getTheme()) {
    case "dark": {
      rendition.themes.override("color", "rgba(255, 255, 255, 0.6)");
      break;
    }
    case "light": {
      rendition.themes.override("color", "rgb(111, 111, 111)");
      break;
    }
  }
  rendition.themes.registerRules("h2Transparent", {
    "h1,h2,h3,h4": {
      "background-color": "transparent !important",
    },
  });
  rendition?.themes.fontSize(`${size.value}%`);
  rendition.themes.select("h2Transparent");
  rendition.themes.override("background-color", "transparent", true);
};

const mediaTypes: ResourceType[] = [
  "image",
  "video",
  "audio",
  "blob",
  "text",
  "textImmutable",
];

const previousLink = ref<string>("");
const nextLink = ref<string>("");
const listing = ref<ResourceItem[] | null>(null);
const name = ref<string>("");
const fullSize = ref<boolean>(false);
const showNav = ref<boolean>(true);
const navTimeout = ref<null | number>(null);
const hoverNav = ref<boolean>(false);
const autoPlay = ref<boolean>(false);
const previousRaw = ref<string>("");
const nextRaw = ref<string>("");
const csvContent = ref<ArrayBuffer | string>("");
const csvError = ref<string>("");
const markdownContent = ref<string>("");

const player = ref<HTMLVideoElement | HTMLAudioElement | null>(null);

const $showError = inject<IToastError>("$showError")!;

const authStore = useAuthStore();
const fileStore = useFileStore();
const layoutStore = useLayoutStore();

const { t } = useI18n();

const route = useRoute();
const router = useRouter();

const hasPrevious = computed(() => previousLink.value !== "");

const hasNext = computed(() => nextLink.value !== "");

const downloadUrl = computed(() =>
  fileStore.req ? api.getDownloadURL(fileStore.req, false) : ""
);

const directUrl = computed(() =>
  fileStore.req ? api.getDownloadURL(fileStore.req, true) : ""
);

const previewUrl = computed(() => {
  if (!fileStore.req) {
    return "";
  }

  if (fileStore.req.type === "image" && !fullSize.value) {
    return api.getPreviewURL(fileStore.req, "big");
  }

  if (isEpub.value) {
    return createURL("api/raw" + fileStore.req.path, {});
  }

  return api.getDownloadURL(fileStore.req, true);
});

const isPdf = computed(() => fileStore.req?.extension.toLowerCase() == ".pdf");
const isEpub = computed(
  () => fileStore.req?.extension.toLowerCase() == ".epub"
);
const isCsv = computed(
  () =>
    fileStore.req?.extension.toLowerCase() == ".csv" &&
    fileStore.req.size <= CSV_MAX_SIZE
);
const isMarkdown = computed(
  () =>
    fileStore.req?.extension.toLowerCase() == ".md" ||
    fileStore.req?.extension.toLowerCase() == ".markdown"
);

// Code file extensions
const codeExtensions = new Set([
  ".js",
  ".mjs",
  ".cjs",
  ".ts",
  ".tsx",
  ".jsx",
  ".py",
  ".go",
  ".rs",
  ".java",
  ".kt",
  ".c",
  ".cpp",
  ".h",
  ".hpp",
  ".cs",
  ".php",
  ".rb",
  ".swift",
  ".dart",
  ".lua",
  ".pl",
  ".r",
  ".scala",
  ".clj",
  ".sh",
  ".bat",
  ".ps1",
  ".html",
  ".htm",
  ".css",
  ".scss",
  ".sass",
  ".json",
  ".xml",
  ".yaml",
  ".yml",
  ".toml",
  ".sql",
  ".graphql",
  ".dockerfile",
]);

const isCode = computed(() => {
  const ext = fileStore.req?.extension.toLowerCase() || "";
  if (codeExtensions.has(ext)) return true;
  // Handle extensionless files like "Dockerfile"
  const name = fileStore.req?.name?.toLowerCase() || "";
  if (ext === "" && name === "dockerfile") return true;
  return false;
});

// Plain-text fallback: any text-typed file that isn't markdown, code or CSV.
// This covers extensionless files (LICENSE, README, Makefile) and unknown text
// extensions that previously matched no viewer and rendered nothing (#3).
const isText = computed(() => {
  const type = fileStore.req?.type;
  if (type !== "text" && type !== "textImmutable") return false;
  const ext = fileStore.req?.extension.toLowerCase() || "";
  if (ext === ".csv") return false;
  return !isMarkdown.value && !isCode.value;
});

const isTextType = computed(
  () => isMarkdown.value || isCode.value || isText.value
);

// Short type label shown as a chip in the unified preview toolbar (#13).
const typeLabel = computed(() => {
  const ext = fileStore.req?.extension.toLowerCase() || "";
  if (isPdf.value) return "PDF";
  if (isEpub.value) return "EPUB";
  if (ext === ".csv") return "CSV";
  if (isMarkdown.value) return "Markdown";
  if (isCode.value) return codeLanguage.value.toUpperCase();
  const type = fileStore.req?.type;
  if (type === "image") return "Image";
  if (type === "video") return "Video";
  if (type === "audio") return "Audio";
  if (isText.value) return ext ? ext.slice(1).toUpperCase() : "Text";
  return "File";
});

const onWheel = (e: WheelEvent) => {
  if (!isTextType.value) {
    e.preventDefault();
    e.stopPropagation();
  }
};

const onTouchMove = (e: TouchEvent) => {
  if (!isTextType.value) {
    e.preventDefault();
    e.stopPropagation();
  }
};

const codeLanguage = computed(() => {
  const ext = fileStore.req?.extension.toLowerCase() || "";
  const langMap: Record<string, string> = {
    ".js": "javascript",
    ".mjs": "javascript",
    ".cjs": "javascript",
    ".ts": "typescript",
    ".tsx": "typescript",
    ".jsx": "javascript",
    ".py": "python",
    ".go": "go",
    ".rs": "rust",
    ".java": "java",
    ".kt": "kotlin",
    ".c": "c",
    ".cpp": "cpp",
    ".h": "c",
    ".hpp": "cpp",
    ".cs": "csharp",
    ".php": "php",
    ".rb": "ruby",
    ".swift": "swift",
    ".dart": "dart",
    ".lua": "lua",
    ".pl": "perl",
    ".r": "r",
    ".scala": "scala",
    ".clj": "clojure",
    ".sh": "bash",
    ".bat": "batch",
    ".ps1": "powershell",
    ".html": "html",
    ".htm": "html",
    ".css": "css",
    ".scss": "scss",
    ".sass": "sass",
    ".json": "json",
    ".xml": "xml",
    ".yaml": "yaml",
    ".yml": "yaml",
    ".toml": "toml",
    ".sql": "sql",
    ".graphql": "graphql",
    ".dockerfile": "dockerfile",
  };
  // Handle extensionless files like "Dockerfile"
  const name = fileStore.req?.name?.toLowerCase() || "";
  if (ext === "" && name === "dockerfile") return "dockerfile";
  return langMap[ext] || "text";
});

const codeContent = ref<string>("");
const textContent = ref<string>("");

const isResizeEnabled = computed(() => resizePreview);

const subtitles = computed(() => {
  if (fileStore.req?.subtitles) {
    return api.getSubtitlesURL(fileStore.req);
  }
  return [];
});

const videoOptions = computed(() => {
  return { autoplay: autoPlay.value };
});

watch(route, () => {
  updatePreview();
  toggleNavigation();
});

// Specify hooks
onMounted(async () => {
  window.addEventListener("keydown", key);
  listing.value = fileStore.oldReq?.items ?? null;
  updatePreview();
});

onBeforeUnmount(() => window.removeEventListener("keydown", key));

// Specify methods
const deleteFile = () => {
  layoutStore.showHover({
    prompt: "delete",
    confirm: () => {
      if (listing.value === null) {
        return;
      }

      const index = listing.value.findIndex((item) => item.name == name.value);
      listing.value.splice(index, 1);

      close();
    },
  });
};

const prev = () => {
  hoverNav.value = false;
  router.replace({ path: previousLink.value });
};

const next = () => {
  hoverNav.value = false;
  router.replace({ path: nextLink.value });
};

const key = (event: KeyboardEvent) => {
  if (layoutStore.currentPrompt !== null) {
    return;
  }
  // When previewing a video, let arrow keys fall through to video.js for
  // seeking instead of switching to the prev/next file. Enter still advances.
  const isVideo = fileStore.req?.type === "video";
  if (event.which === 13) {
    // enter
    if (hasNext.value) next();
  } else if (event.which === 39) {
    // right arrow
    if (isVideo) return;
    if (hasNext.value) next();
  } else if (event.which === 37) {
    // left arrow
    if (isVideo) return;
    if (hasPrevious.value) prev();
  } else if (event.which === 27) {
    // esc
    close();
  }
};
const updatePreview = async () => {
  if (player.value && player.value.paused && !player.value.ended) {
    autoPlay.value = false;
  }

  const dirs = route.path.split("/");
  name.value = decodeURIComponent(dirs[dirs.length - 1]);

  // Load CSV content if it's a CSV file
  if (isCsv.value && fileStore.req) {
    csvContent.value = "";
    csvError.value = "";

    if (fileStore.req.size > CSV_MAX_SIZE) {
      csvError.value = t("files.csvTooLarge");
    } else {
      if (fileStore.req.rawContent != null) {
        csvContent.value = fileStore.req.rawContent;
      } else {
        csvContent.value = fileStore.req.content ?? "";
      }
    }
  }

  // Load markdown content if it's a markdown file
  if (isMarkdown.value && fileStore.req) {
    markdownContent.value = "";
    const raw =
      fileStore.req.rawContent != null
        ? typeof fileStore.req.rawContent === "string"
          ? fileStore.req.rawContent
          : new TextDecoder().decode(fileStore.req.rawContent)
        : (fileStore.req.content ?? "");
    try {
      const stripped = raw.replace(/^---\n[\s\S]*?\n---\n?/, "");
      markdownContent.value = DOMPurify.sanitize(await marked(stripped));
    } catch (e) {
      console.error("Failed to render markdown:", e);
      markdownContent.value = "";
    }
  }

  // Load code content if it's a code file
  if (isCode.value && fileStore.req) {
    codeContent.value = "";
    const raw =
      fileStore.req.rawContent != null
        ? typeof fileStore.req.rawContent === "string"
          ? fileStore.req.rawContent
          : new TextDecoder().decode(fileStore.req.rawContent)
        : (fileStore.req.content ?? "");
    codeContent.value = raw;
  }

  // Load text content if it's a text file
  if (isText.value && fileStore.req) {
    textContent.value = "";
    const raw =
      fileStore.req.rawContent != null
        ? typeof fileStore.req.rawContent === "string"
          ? fileStore.req.rawContent
          : new TextDecoder().decode(fileStore.req.rawContent)
        : (fileStore.req.content ?? "");
    textContent.value = raw;
  }

  if (!listing.value) {
    try {
      const path = url.removeLastDir(route.path);
      const res = await api.fetch(path);
      listing.value = res.items;
    } catch (e: any) {
      $showError(e);
    }
  }

  previousLink.value = "";
  nextLink.value = "";
  if (listing.value) {
    for (let i = 0; i < listing.value.length; i++) {
      if (listing.value[i].name !== name.value) {
        continue;
      }

      for (let j = i - 1; j >= 0; j--) {
        if (mediaTypes.includes(listing.value[j].type)) {
          previousLink.value = listing.value[j].url;
          previousRaw.value = prefetchUrl(listing.value[j]);
          break;
        }
      }
      for (let j = i + 1; j < listing.value.length; j++) {
        if (mediaTypes.includes(listing.value[j].type)) {
          nextLink.value = listing.value[j].url;
          nextRaw.value = prefetchUrl(listing.value[j]);
          break;
        }
      }

      return;
    }
  }
};

const prefetchUrl = (item: ResourceItem) => {
  if (item.type !== "image") {
    return "";
  }

  return fullSize.value
    ? api.getDownloadURL(item, true)
    : api.getPreviewURL(item, "big");
};

const toggleSize = () => (fullSize.value = !fullSize.value);

const toggleNavigation = throttle(function () {
  showNav.value = true;

  if (navTimeout.value) {
    clearTimeout(navTimeout.value);
  }

  navTimeout.value = window.setTimeout(() => {
    showNav.value = false || hoverNav.value;
    navTimeout.value = null;
  }, 1500);
}, 500);

const close = () => {
  const uri = url.removeLastDir(route.path) + "/";
  router.push({ path: uri });
};

const download = () => window.open(downloadUrl.value);
const openDirect = () => window.open(directUrl.value);

const editAsText = () => {
  router.push({ path: route.path, query: { edit: "true" } });
};
</script>
