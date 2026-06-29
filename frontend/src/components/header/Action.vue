<template>
  <button
    @click="action"
    :aria-label="label"
    :title="label"
    class="action"
    :class="{ 'action--danger': danger }"
    :data-danger="danger || null"
  >
    <FbIcon v-if="fbIcon" :name="fbIcon" :size="size" />
    <i v-else class="material-icons">{{ icon }}</i>
    <span>{{ label }}</span>
    <span v-if="counter && counter > 0" class="counter">{{ counter }}</span>
  </button>
</template>

<script setup lang="ts">
import { useLayoutStore } from "@/stores/layout";
import FbIcon from "@/components/FbIcon.vue";
import type { IconName } from "@/utils/icons";

const props = defineProps<{
  icon?: string;
  fbIcon?: IconName;
  label?: string;
  counter?: number;
  show?: string;
  danger?: boolean;
  size?: string;
}>();

const emit = defineEmits<{
  (e: "action"): any;
}>();

const layoutStore = useLayoutStore();

const action = () => {
  if (props.show) {
    layoutStore.showHover(props.show);
  }

  emit("action");
};
</script>
