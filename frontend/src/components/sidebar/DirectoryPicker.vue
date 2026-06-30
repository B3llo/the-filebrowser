<template>
  <div class="fb-dir-picker">
    <!-- Breadcrumb of current path -->
    <div class="fb-dir-picker-breadcrumb">
      <button
        v-for="(segment, i) in breadcrumbs"
        :key="i"
        class="fb-dir-picker-crumb"
        :class="{ 'is-current': i === breadcrumbs.length - 1 }"
        @click="navigateTo(segment.path)"
      >
        {{ segment.label }}
      </button>
    </div>

    <!-- Filter input -->
    <input
      v-model="filter"
      class="fb-input fb-dir-picker-filter"
      :placeholder="$t('browse.filterFolders', 'Filter folders…')"
      type="text"
      autocomplete="off"
    />

    <!-- Directory list -->
    <div class="fb-dir-picker-list">
      <div v-if="loading" class="fb-dir-picker-loading">
        <fb-icon name="clock" size="16px" />
      </div>
      <div v-else-if="filteredDirs.length === 0" class="fb-dir-picker-empty">
        {{ $t("browse.noFolders", "No folders here") }}
      </div>
      <button
        v-for="dir in filteredDirs"
        :key="dir.path"
        class="fb-dir-picker-item"
        @click="navigateTo(dir.path)"
      >
        <fb-icon name="folder" size="16px" />
        <span>{{ dir.name }}</span>
        <fb-icon v-if="dir.hasChildren" name="chevron-right" size="14px" class="fb-dir-picker-chevron" />
      </button>
    </div>

    <!-- Selected path display + confirm -->
    <div class="fb-dir-picker-footer">
      <span class="fb-dir-picker-selected">{{ currentPath }}</span>
      <button class="fb-btn fb-btn--sm fb-btn--primary" @click="select">
        {{ $t("browse.selectFolder", "Select") }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { browse, type BrowseDirEntry } from "@/api/browse";
import FbIcon from "@/components/FbIcon.vue";

const props = defineProps<{
  modelValue: string;
}>();
const emit = defineEmits<{
  "update:modelValue": [value: string];
}>();

const currentPath = ref(props.modelValue || "/");
const dirs = ref<BrowseDirEntry[]>([]);
const loading = ref(false);
const filter = ref("");

const breadcrumbs = computed(() => {
  const parts = currentPath.value.split("/").filter(Boolean);
  const crumbs = [{ label: "/", path: "/" }];
  let acc = "/";
  for (const p of parts) {
    acc += p + "/";
    crumbs.push({ label: p, path: acc });
  }
  return crumbs;
});

const filteredDirs = computed(() => {
  const q = filter.value.toLowerCase();
  if (!q) return dirs.value;
  return dirs.value.filter((d) => d.name.toLowerCase().includes(q));
});

async function loadDirs(path: string) {
  loading.value = true;
  filter.value = "";
  try {
    dirs.value = await browse(path);
  } catch {
    dirs.value = [];
  } finally {
    loading.value = false;
  }
}

function navigateTo(path: string) {
  currentPath.value = path;
  emit("update:modelValue", path);
}

function select() {
  emit("update:modelValue", currentPath.value);
}

watch(currentPath, (path) => loadDirs(path));

onMounted(() => {
  loadDirs(currentPath.value);
});
</script>

<style scoped>
.fb-dir-picker {
  display: flex;
  flex-direction: column;
  gap: 8px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 10px;
  overflow: hidden;
  padding: 10px;
}

.fb-dir-picker-breadcrumb {
  display: flex;
  flex-wrap: wrap;
  gap: 2px;
  align-items: center;
  min-height: 28px;
}

.fb-dir-picker-crumb {
  font-size: 12px;
  color: var(--accent);
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 2px 6px;
  border-radius: 4px;
  transition: background 0.1s;
}

.fb-dir-picker-crumb:hover {
  background: var(--hover);
}

.fb-dir-picker-crumb.is-current {
  color: var(--text);
  font-weight: 600;
  cursor: default;
}

.fb-dir-picker-filter {
  font-size: 13px;
  padding: 6px 10px;
}

.fb-dir-picker-list {
  max-height: 200px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.fb-dir-picker-loading,
.fb-dir-picker-empty {
  padding: 12px;
  text-align: center;
  color: var(--dim);
  font-size: 13px;
}

.fb-dir-picker-item {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 6px 8px;
  border-radius: 6px;
  background: transparent;
  border: none;
  cursor: pointer;
  font-size: 13px;
  color: var(--text);
  text-align: left;
  transition: background 0.1s;
}

.fb-dir-picker-item:hover {
  background: var(--hover);
}

.fb-dir-picker-chevron {
  margin-left: auto;
  color: var(--faint);
}

.fb-dir-picker-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding-top: 6px;
  border-top: 1px solid var(--border);
  margin-top: 2px;
}

.fb-dir-picker-selected {
  font-size: 12px;
  color: var(--dim);
  font-family: var(--font-mono, monospace);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  min-width: 0;
}

.fb-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
  transition:
    background 0.1s,
    color 0.1s;
}

.fb-btn--sm {
  font-size: 12px;
  padding: 5px 12px;
}

.fb-btn--primary {
  background: var(--accent);
  color: var(--on-accent);
}

.fb-btn--primary:hover {
  filter: brightness(1.05);
}
</style>
