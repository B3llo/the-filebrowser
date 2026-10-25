<template>
  <Transition name="fb-upload-toast">
    <div v-if="uploads.length > 0 && !dismissed" class="fb-upload-toast">
      <!-- Header -->
      <div class="fb-upload-toast-header">
        <!-- Spinner while uploading -->
        <svg
          v-if="uploadsBusy"
          class="fb-upload-toast-spinner"
          viewBox="0 0 24 24"
          fill="none"
          stroke="var(--accent)"
          stroke-width="2.2"
          stroke-linecap="round"
        >
          <path d="M21 12a9 9 0 1 1-6.2-8.5" />
        </svg>
        <!-- Checkmark when complete -->
        <svg
          v-else
          class="fb-upload-toast-check"
          viewBox="0 0 24 24"
          fill="none"
          stroke="var(--accent)"
          stroke-width="2.2"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <path d="M20 6L9 17l-5-5" />
        </svg>
        <span class="fb-upload-toast-title">{{ uploadTitle }}</span>
        <button
          class="fb-upload-toast-close"
          @click="closeToast"
          aria-label="Close"
        >
          <svg
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M18 6L6 18" />
            <path d="M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- File list -->
      <div class="fb-upload-toast-files">
        <div
          v-for="upload in uploads"
          :key="upload.path"
          class="fb-upload-toast-file"
        >
          <div class="fb-upload-toast-file-info">
            <div class="fb-upload-toast-file-name">{{ upload.name }}</div>
            <div class="fb-upload-toast-file-progress">
              <div
                class="fb-upload-toast-file-progress-fill"
                :style="{ width: upload.pct + '%' }"
              />
            </div>
          </div>
          <!-- Checkmark when done -->
          <svg
            v-if="upload.done"
            class="fb-upload-toast-file-check"
            viewBox="0 0 24 24"
            fill="none"
            stroke="var(--accent)"
            stroke-width="2.4"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M20 6L9 17l-5-5" />
          </svg>
          <!-- Percentage while uploading -->
          <span v-else class="fb-upload-toast-file-pct">{{ upload.pct }}%</span>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { useFileStore } from "@/stores/file";
import { useUploadStore } from "@/stores/upload";
import { computed, onBeforeUnmount, ref, watch } from "vue";
import buttons from "@/utils/buttons";
import { useI18n } from "vue-i18n";

const { t } = useI18n({});

const fileStore = useFileStore();
const uploadStore = useUploadStore();

interface UploadDisplay {
  path: string;
  name: string;
  pct: number;
  done: boolean;
}

// Uploads that already finished stay in this list until the whole batch is
// done. Rebuilding it from `activeUploads` alone made each completed file
// vanish from the toast as soon as the next one started.
const uploads = ref<UploadDisplay[]>([]);
const dismissed = ref(false);
let autoCloseTimer: number | null = null;

// The toast is a short summary, not a full transfer log. Keep the list
// bounded on very large batches (active uploads are always the newest ones).
const MAX_DISPLAYED_UPLOADS = 100;

const uploadsBusy = computed(() => uploadStore.activeUploads.size > 0);

const uploadsDone = computed(
  () => uploads.value.length > 0 && !uploadsBusy.value
);

const uploadTitle = computed(() => {
  if (uploadsDone.value) return "Upload complete";
  // `pendingUploadCount` includes queued files that have not started yet,
  // unlike `activeUploads.size` (capped by the concurrency limit).
  const count = Math.max(0, uploadStore.pendingUploadCount);
  return `Uploading ${count} item${count !== 1 ? "s" : ""}`;
});

const cancelAutoClose = () => {
  if (autoCloseTimer !== null) {
    window.clearTimeout(autoCloseTimer);
    autoCloseTimer = null;
  }
};

const scheduleAutoClose = () => {
  cancelAutoClose();
  autoCloseTimer = window.setTimeout(() => {
    autoCloseTimer = null;
    uploads.value = [];
  }, 2000);
};

// Sync the current upload state into the display array, keeping entries that
// have already completed.
const syncUploads = () => {
  const active = Array.from(uploadStore.activeUploads);
  const activePaths = new Set(active.map((upload) => upload.path));
  const known = new Map(uploads.value.map((upload) => [upload.path, upload]));

  for (const upload of active) {
    const pct =
      upload.totalBytes > 0
        ? Math.round((upload.sentBytes / upload.totalBytes) * 100)
        : 0;
    const display = known.get(upload.path);

    if (display) {
      display.pct = pct;
      display.done = pct >= 100;
    } else {
      uploads.value.push({
        path: upload.path,
        name: upload.name,
        pct,
        done: pct >= 100,
      });
    }
  }

  // Anything that left the active set has finished (or was aborted).
  for (const display of uploads.value) {
    if (!activePaths.has(display.path)) {
      display.done = true;
      display.pct = 100;
    }
  }

  if (uploads.value.length > MAX_DISPLAYED_UPLOADS) {
    uploads.value.splice(0, uploads.value.length - MAX_DISPLAYED_UPLOADS);
  }
};

// Watch for upload progress changes and for memberships changing (a file
// finishing while another queued file starts keeps the same active count).
watch(() => uploadStore.sentBytes, syncUploads);
watch(() => Array.from(uploadStore.activeUploads), syncUploads);

// Start a fresh list for a new batch and auto-close once the batch is done.
watch(
  () => uploadStore.activeUploads.size,
  (size, previous) => {
    if (size > 0) {
      cancelAutoClose();
      if (previous === 0) {
        uploads.value = [];
        dismissed.value = false;
      }
      syncUploads();
      return;
    }

    syncUploads();
    scheduleAutoClose();
  }
);

const closeToast = () => {
  cancelAutoClose();

  if (uploadsBusy.value) {
    if (!confirm(t("upload.abortUpload"))) return;

    buttons.done("upload");
    uploadStore.abort();
    fileStore.resetScroll = true;
    fileStore.reload = true;
    dismissed.value = true;
    return;
  }

  uploads.value = [];
};

onBeforeUnmount(cancelAutoClose);
</script>

<style scoped>
.fb-upload-toast {
  position: fixed;
  right: 24px;
  bottom: 24px;
  width: 346px;
  z-index: 55;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 13px;
  box-shadow: var(--shadow-menu);
  overflow: hidden;
  animation: fbup 0.22s;
}

.fb-upload-toast-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 13px 16px;
  border-bottom: 1px solid var(--border);
}

.fb-upload-toast-spinner {
  width: 16px;
  height: 16px;
  flex: 0 0 auto;
  animation: fbspin 0.9s linear infinite;
}

.fb-upload-toast-check {
  width: 16px;
  height: 16px;
  flex: 0 0 auto;
}

.fb-upload-toast-title {
  flex: 1;
  font-size: 13.5px;
  font-weight: 600;
  color: var(--text);
}

.fb-upload-toast-close {
  width: 26px;
  height: 26px;
  border: none;
  background: transparent;
  color: var(--dim);
  border-radius: 7px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
}

.fb-upload-toast-close:hover {
  background: var(--hover);
  color: var(--text);
}

.fb-upload-toast-close svg {
  width: 15px;
  height: 15px;
}

.fb-upload-toast-files {
  max-height: 240px;
  overflow-y: auto;
}

.fb-upload-toast-file {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 11px 16px;
}

.fb-upload-toast-file-info {
  flex: 1;
  min-width: 0;
}

.fb-upload-toast-file-name {
  font-size: 13px;
  color: var(--text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.fb-upload-toast-file-progress {
  height: 4px;
  border-radius: 3px;
  background: var(--hover);
  overflow: hidden;
  margin-top: 7px;
}

.fb-upload-toast-file-progress-fill {
  height: 100%;
  background: var(--accent);
  border-radius: 3px;
  transition: width 0.2s;
}

.fb-upload-toast-file-check {
  width: 16px;
  height: 16px;
  flex: 0 0 auto;
}

.fb-upload-toast-file-pct {
  font-size: 11.5px;
  color: var(--faint);
  font-variant-numeric: tabular-nums;
  flex: 0 0 auto;
  width: 34px;
  text-align: right;
}

/* Transition */
.fb-upload-toast-enter-active {
  animation: fbup 0.22s;
}

.fb-upload-toast-leave-active {
  animation: fbfade 0.15s reverse;
}

/* Mobile: span the width minus safe margins */
@media (max-width: 736px) {
  .fb-upload-toast {
    right: 12px;
    left: 12px;
    bottom: 12px;
    width: auto;
    margin-bottom: env(safe-area-inset-bottom);
  }
}
</style>
