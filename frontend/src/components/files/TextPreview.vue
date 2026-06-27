<template>
  <div class="text-preview">
    <div class="text-header">
      <span class="file-type">{{ fileType }}</span>
      <span class="line-count">{{ lineCount }} lines</span>
    </div>
    <pre class="text-content">{{ content }}</pre>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface Props {
  content: string;
  filename?: string;
}

const props = withDefaults(defineProps<Props>(), {
  filename: "",
});

const lineCount = computed(() => {
  return props.content.split("\n").length;
});

const fileType = computed(() => {
  if (!props.filename) return "Text";

  const ext = props.filename.split(".").pop()?.toLowerCase() || "";
  const typeMap: Record<string, string> = {
    txt: "Text",
    log: "Log",
    ini: "INI",
    cfg: "Config",
    conf: "Config",
    md: "Markdown",
    markdown: "Markdown",
    csv: "CSV",
    tsv: "TSV",
  };

  return typeMap[ext] || "Text";
});
</script>

<style scoped>
.text-preview {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--bg, #ffffff);
  border-radius: 8px;
  overflow: hidden;
}

.text-header {
  padding: 12px 16px;
  background: var(--header-bg, #f5f5f5);
  border-bottom: 1px solid var(--border-color, #e0e0e0);
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
}

.file-type {
  background: var(--primary-color, #4a9eff);
  color: white;
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.line-count {
  color: var(--text-secondary, #666);
  font-size: 12px;
}

.text-content {
  flex: 1;
  margin: 0;
  padding: 16px;
  overflow: auto;
  font-family: "Monaco", "Menlo", "Ubuntu Mono", monospace;
  font-size: 14px;
  line-height: 1.6;
  color: var(--text-primary, #333);
  background: transparent;
  white-space: pre-wrap;
  word-wrap: break-word;
}

/* Custom scrollbar */
.text-content::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.text-content::-webkit-scrollbar-track {
  background: var(--header-bg, #f5f5f5);
}

.text-content::-webkit-scrollbar-thumb {
  background: var(--border-color, #ccc);
  border-radius: 4px;
}

.text-content::-webkit-scrollbar-thumb:hover {
  background: var(--text-secondary, #999);
}
</style>
