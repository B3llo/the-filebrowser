<template>
  <Transition name="fb-extract-toast">
    <div v-if="extractStore.tasks.length > 0" class="fb-extract-toast">
      <div class="fb-extract-toast-header">
        <!-- Spinner while extracting -->
        <svg
          v-if="busy"
          class="fb-extract-toast-spinner"
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
          class="fb-extract-toast-check"
          viewBox="0 0 24 24"
          fill="none"
          stroke="var(--accent)"
          stroke-width="2.2"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <path d="M20 6L9 17l-5-5" />
        </svg>
        <span class="fb-extract-toast-title">{{ title }}</span>
      </div>

      <div class="fb-extract-toast-files">
        <div
          v-for="task in extractStore.tasks"
          :key="task.path"
          class="fb-extract-toast-file"
        >
          <div class="fb-extract-toast-file-name">{{ task.name }}</div>
          <!-- Checkmark when done -->
          <svg
            v-if="task.done"
            class="fb-extract-toast-file-check"
            viewBox="0 0 24 24"
            fill="none"
            stroke="var(--accent)"
            stroke-width="2.4"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M20 6L9 17l-5-5" />
          </svg>
          <!-- Spinner while this file is extracting -->
          <svg
            v-else
            class="fb-extract-toast-file-spinner"
            viewBox="0 0 24 24"
            fill="none"
            stroke="var(--dim)"
            stroke-width="2.2"
            stroke-linecap="round"
          >
            <path d="M21 12a9 9 0 1 1-6.2-8.5" />
          </svg>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useExtractStore } from "@/stores/extract";

const extractStore = useExtractStore();

const busy = computed(() => extractStore.tasks.some((t) => !t.done));
const title = computed(() =>
  busy.value ? "Extracting archive…" : "Extraction complete"
);
</script>

<style scoped>
.fb-extract-toast {
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

.fb-extract-toast-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 13px 16px;
  border-bottom: 1px solid var(--border);
}

.fb-extract-toast-spinner,
.fb-extract-toast-check {
  width: 16px;
  height: 16px;
  flex: 0 0 auto;
}

.fb-extract-toast-spinner {
  animation: fbspin 0.9s linear infinite;
}

.fb-extract-toast-title {
  flex: 1;
  font-size: 13.5px;
  font-weight: 600;
  color: var(--text);
}

.fb-extract-toast-files {
  max-height: 240px;
  overflow-y: auto;
}

.fb-extract-toast-file {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 11px 16px;
}

.fb-extract-toast-file-name {
  flex: 1;
  min-width: 0;
  font-size: 13px;
  color: var(--text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.fb-extract-toast-file-check {
  width: 16px;
  height: 16px;
  flex: 0 0 auto;
}

.fb-extract-toast-file-spinner {
  width: 14px;
  height: 14px;
  flex: 0 0 auto;
  animation: fbspin 0.9s linear infinite;
}

/* Transition */
.fb-extract-toast-enter-active {
  animation: fbup 0.22s;
}

.fb-extract-toast-leave-active {
  animation: fbfade 0.15s reverse;
}

/* Mobile: span the width minus safe margins */
@media (max-width: 736px) {
  .fb-extract-toast {
    right: 12px;
    left: 12px;
    bottom: 12px;
    width: auto;
    margin-bottom: env(safe-area-inset-bottom);
  }
}
</style>
