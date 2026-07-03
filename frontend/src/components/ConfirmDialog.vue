<template>
  <div class="confirm-dialog-backdrop" @click.self="$emit('cancel')">
    <div class="confirm-dialog">
      <div class="confirm-dialog-content">
        <p>{{ message }}</p>
      </div>
      <div class="confirm-dialog-actions">
        <button
          class="confirm-dialog-btn confirm-dialog-btn--cancel"
          @click="$emit('cancel')"
        >
          {{ cancelText }}
        </button>
        <button
          class="confirm-dialog-btn confirm-dialog-btn--confirm"
          :class="{ 'confirm-dialog-btn--danger': danger }"
          @click="$emit('confirm')"
        >
          {{ confirmText }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  message: string;
  confirmText: string;
  cancelText: string;
  danger?: boolean;
}>();

defineEmits<{
  confirm: [];
  cancel: [];
}>();
</script>

<style scoped>
.confirm-dialog-backdrop {
  position: fixed;
  inset: 0;
  background-color: oklch(0 0 0 / 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 10000;
  animation: fbfade 0.15s ease-out;
}

.confirm-dialog {
  background: var(--surface, oklch(1 0 0));
  border: 1px solid var(--border, oklch(0.87 0.01 100));
  border-radius: 12px;
  box-shadow: var(--shadow-menu, 0 4px 24px oklch(0 0 0 / 0.12));
  padding: 20px;
  min-width: 320px;
  max-width: 400px;
  animation: fbpop 0.12s ease-out;
}

.confirm-dialog-content {
  margin-bottom: 20px;
}

.confirm-dialog-content p {
  margin: 0;
  font-size: 14px;
  color: var(--dim, oklch(0.4 0.01 100));
  line-height: 1.5;
}

.confirm-dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.confirm-dialog-btn {
  padding: 8px 16px;
  border-radius: 8px;
  font-size: 13.5px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.08s, color 0.08s;
  font-family: inherit;
}

.confirm-dialog-btn--cancel {
  background: transparent;
  border: 1px solid var(--border, oklch(0.87 0.01 100));
  color: var(--dim, oklch(0.4 0.01 100));
}

.confirm-dialog-btn--cancel:hover {
  background: var(--hover, oklch(0.92 0.01 100));
}

.confirm-dialog-btn--confirm {
  background: var(--accent, oklch(0.55 0.13 255));
  border: 1px solid transparent;
  color: white;
}

.confirm-dialog-btn--confirm:hover {
  background: var(--accent-press, oklch(0.5 0.13 255));
}

.confirm-dialog-btn--danger {
  background: var(--red, oklch(0.55 0.2 25));
}

.confirm-dialog-btn--danger:hover {
  background: var(--dark-red, oklch(0.49 0.2 25));
}
</style>
