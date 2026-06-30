<template>
  <div class="fb-folder-color-picker" @click.stop>
    <div class="fb-folder-color-grid">
      <button
        v-for="color in colors"
        :key="color.name"
        class="fb-folder-color-btn"
        :class="{ active: currentColor === color.name }"
        :style="{ background: color.fill, borderColor: color.stroke }"
        :title="t(`folderColors.${color.name}`)"
        @click="selectColor(color.name)"
        :aria-label="t(`folderColors.${color.name}`)"
      >
        <svg
          viewBox="0 0 24 24"
          :fill="color.fill"
          :stroke="color.stroke"
          stroke-width="1.6"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="fb-folder-color-icon"
        >
          <path
            d="M20 19a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2h-7.6a1 1 0 0 1-.8-.4L10.3 4.9a1 1 0 0 0-.8-.4H4a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2Z"
          />
        </svg>
      </button>
    </div>
    <button
      v-if="currentColor"
      class="fb-folder-color-remove"
      @click="removeColor"
    >
      {{ t("folderColors.remove") }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";

const { t } = useI18n();

defineProps<{
  currentColor?: string;
}>();

const emit = defineEmits<{
  (e: "select", color: string): void;
  (e: "remove"): void;
}>();

const colors = [
  {
    name: "blue",
    fill: "oklch(0.957 0.027 255)",
    stroke: "oklch(0.55 0.13 255)",
  },
  {
    name: "green",
    fill: "oklch(0.95 0.04 145)",
    stroke: "oklch(0.55 0.12 145)",
  },
  {
    name: "purple",
    fill: "oklch(0.94 0.04 300)",
    stroke: "oklch(0.50 0.13 300)",
  },
  {
    name: "orange",
    fill: "oklch(0.95 0.04 60)",
    stroke: "oklch(0.60 0.14 60)",
  },
  {
    name: "pink",
    fill: "oklch(0.94 0.04 350)",
    stroke: "oklch(0.55 0.12 350)",
  },
  {
    name: "red",
    fill: "oklch(0.94 0.04 25)",
    stroke: "oklch(0.55 0.15 25)",
  },
  {
    name: "yellow",
    fill: "oklch(0.96 0.04 85)",
    stroke: "oklch(0.70 0.12 85)",
  },
  {
    name: "cyan",
    fill: "oklch(0.95 0.04 190)",
    stroke: "oklch(0.55 0.10 190)",
  },
];

const selectColor = (color: string) => {
  emit("select", color);
};

const removeColor = () => {
  emit("remove");
};
</script>

<style scoped>
.fb-folder-color-picker {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 12px;
  box-shadow: var(--shadow-menu);
  min-width: 180px;
}

.fb-folder-color-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
}

.fb-folder-color-btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: 2px solid transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition:
    transform 0.15s ease,
    box-shadow 0.15s ease;
  padding: 0;
}

.fb-folder-color-btn:hover {
  transform: scale(1.1);
}

.fb-folder-color-btn.active {
  box-shadow:
    0 0 0 2px var(--surface),
    0 0 0 4px var(--accent);
}

.fb-folder-color-icon {
  width: 20px;
  height: 20px;
}

.fb-folder-color-remove {
  width: 100%;
  margin-top: 10px;
  padding: 6px 12px;
  background: var(--hover);
  border: 1px solid var(--border);
  border-radius: 8px;
  color: var(--dim);
  font-size: 12px;
  cursor: pointer;
  transition:
    background 0.15s ease,
    color 0.15s ease;
}

.fb-folder-color-remove:hover {
  background: var(--border);
  color: var(--text);
}
</style>
