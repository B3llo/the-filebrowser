<template>
  <div class="t-container">
    <span class="t-message">{{ message }}</span>
    <button v-if="isReport" class="t-action" @click.stop="clicked">
      {{ reportText }}
    </button>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  message: string;
  reportText?: string;
  isReport?: boolean;
}>();

const clicked = () => {
  window.open("https://github.com/filebrowser/filebrowser/issues/new/choose");
};
</script>

<style scoped>
.t-container {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.t-message {
  flex: 1;
  min-width: 0;
}

.t-action {
  flex-shrink: 0;
  height: 30px;
  padding: 0 12px;
  border-radius: 8px;
  color: var(--accent);
  cursor: pointer;
  border: 1px solid var(--border);
  background: var(--hover);
  font-size: 12.5px;
  font-weight: 500;
  font-family: inherit;
  transition:
    background 0.12s,
    border-color 0.12s;
}

.t-action:hover {
  background: var(--sel);
  border-color: var(--sel-border);
}

html[dir="rtl"] .t-action {
  margin-left: 0;
  margin-right: 12px;
}
</style>
