<template>
  <div class="pdf-preview">
    <div class="pdf-header">
      <span class="file-type">PDF</span>
      <span class="file-name">{{ filename }}</span>
    </div>
    <div class="pdf-content">
      <iframe
        v-if="src"
        :src="src"
        class="pdf-iframe"
        title="PDF Preview"
      ></iframe>
      <div v-else class="pdf-placeholder">
        <i class="material-icons">picture_as_pdf</i>
        <p>Unable to load PDF preview</p>
        <a :href="downloadUrl" class="download-button" target="_blank">
          Download PDF
        </a>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  src?: string;
  filename?: string;
  downloadUrl?: string;
}

withDefaults(defineProps<Props>(), {
  src: "",
  filename: "document.pdf",
  downloadUrl: "",
});
</script>

<style scoped>
.pdf-preview {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--bg, #ffffff);
  border-radius: 8px;
  overflow: hidden;
}

.pdf-header {
  padding: 12px 16px;
  background: var(--header-bg, #f5f5f5);
  border-bottom: 1px solid var(--border-color, #e0e0e0);
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
}

.file-type {
  background: #e74c3c;
  color: white;
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.file-name {
  color: var(--text-primary, #333);
  font-size: 14px;
  font-weight: 500;
}

.pdf-content {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--pdf-bg, #525659);
}

.pdf-iframe {
  width: 100%;
  height: 100%;
  border: none;
}

.pdf-placeholder {
  text-align: center;
  color: var(--text-secondary, #666);
}

.pdf-placeholder i {
  font-size: 64px;
  margin-bottom: 16px;
  color: #e74c3c;
}

.pdf-placeholder p {
  margin-bottom: 16px;
}

.download-button {
  display: inline-block;
  padding: 10px 20px;
  background: var(--primary-color, #4a9eff);
  color: white;
  text-decoration: none;
  border-radius: 4px;
  font-size: 14px;
  transition: background-color 0.2s;
}

.download-button:hover {
  background: var(--primary-hover, #3a8ee6);
}
</style>
