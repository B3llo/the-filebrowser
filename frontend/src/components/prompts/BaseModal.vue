<template>
  <div id="modal-background" @click="backgroundClick">
    <div ref="modalContainer" role="dialog" aria-modal="true" tabindex="-1">
      <slot></slot>
    </div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref } from "vue";

const emit = defineEmits(["closed"]);

const modalContainer = ref<HTMLElement | null>(null);
let opener: HTMLElement | null = null;

const focusableSelector =
  'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';

const getFocusable = (): HTMLElement[] => {
  if (!modalContainer.value) return [];
  return Array.from(
    modalContainer.value.querySelectorAll<HTMLElement>(focusableSelector)
  ).filter((el) => el.offsetParent !== null || el === document.activeElement);
};

const focusInitial = () => {
  const element =
    modalContainer.value?.querySelector<HTMLElement>("#focus-prompt");
  if (element) {
    element.focus();
    return;
  }
  const first = getFocusable()[0];
  if (first) {
    first.focus();
    return;
  }
  modalContainer.value?.focus();
};

const trapTab = (event: KeyboardEvent) => {
  if (event.key !== "Tab") return;
  const focusable = getFocusable();
  if (focusable.length === 0) {
    event.preventDefault();
    modalContainer.value?.focus();
    return;
  }
  const first = focusable[0];
  const last = focusable[focusable.length - 1];
  const active = document.activeElement as HTMLElement | null;
  if (event.shiftKey) {
    if (active === first || !modalContainer.value?.contains(active)) {
      event.preventDefault();
      last.focus();
    }
  } else if (active === last) {
    event.preventDefault();
    first.focus();
  }
};

const onKeydown = (event: KeyboardEvent) => {
  if (event.key === "Escape") {
    event.stopImmediatePropagation();
    emit("closed");
    return;
  }
  trapTab(event);
};

onMounted(() => {
  opener = document.activeElement as HTMLElement | null;
  nextTick(() => focusInitial());

  window.addEventListener("keydown", onKeydown);
});

onBeforeUnmount(() => {
  window.removeEventListener("keydown", onKeydown);
  opener?.focus?.();
  opener = null;
});

const backgroundClick = (event: Event) => {
  const target = event.target as HTMLElement;
  if (target.id == "modal-background") {
    emit("closed");
  }
};
</script>

<style scoped>
#modal-background {
  position: fixed;
  inset: 0;
  background-color: #00000096;
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 10000;
  animation: ease-in 150ms opacity-enter;
}

@keyframes opacity-enter {
  from {
    opacity: 0;
  }

  to {
    opacity: 1;
  }
}
</style>
