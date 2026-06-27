<template>
  <div
    class="context-menu-backdrop"
    v-show="show"
    @click.self="hideContextMenu"
    @contextmenu.prevent="hideContextMenu"
  />
  <div
    class="context-menu"
    ref="contextMenu"
    v-show="show"
    :style="menuStyle"
  >
    <slot />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, onUnmounted } from "vue";

const emit = defineEmits(["hide"]);
const props = defineProps<{ show: boolean; pos: { x: number; y: number } }>();
const contextMenu = ref<HTMLElement | null>(null);
// Kept hidden until measured & positioned so it never flashes at the corner.
const menuStyle = ref<Record<string, string>>({ visibility: "hidden" });

const hideContextMenu = () => {
  emit("hide");
};

const positionMenu = () => {
  const menu = contextMenu.value;
  if (!menu) return;

  // offsetWidth/Height are layout metrics, immune to the fbpop scale transform.
  const menuWidth = menu.offsetWidth;
  const menuHeight = menu.offsetHeight;
  const gap = 8;
  const vw = window.innerWidth;
  const vh = window.innerHeight;

  let left = props.pos.x;
  let top = props.pos.y;

  // Clamp horizontally inside the viewport.
  if (left + menuWidth > vw - gap) left = vw - menuWidth - gap;
  if (left < gap) left = gap;

  // Flip above the cursor if it would overflow the bottom, then clamp top.
  if (top + menuHeight > vh - gap) top = vh - menuHeight - gap;
  if (top < gap) top = gap;

  menuStyle.value = {
    visibility: "visible",
    top: `${top}px`,
    left: `${left}px`,
  };
};

watch(
  () => props.show,
  async (val) => {
    if (val) {
      menuStyle.value = { visibility: "hidden" };
      document.addEventListener("click", hideContextMenu);
      await nextTick();
      // rAF guarantees the browser has laid the menu out before we measure.
      requestAnimationFrame(positionMenu);
    } else {
      document.removeEventListener("click", hideContextMenu);
    }
  }
);

onUnmounted(() => {
  document.removeEventListener("click", hideContextMenu);
});
</script>
