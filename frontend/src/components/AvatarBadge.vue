<template>
  <div class="fb-avatar-badge" :style="badgeStyle" aria-hidden="true">
    <img
      v-if="kind === 'custom'"
      class="fb-avatar-badge-img"
      :src="src"
      :alt="''"
    />
    <!-- preset.svg is static, developer-authored markup (see avatarPresets.ts) -->
    <div
      v-else-if="kind === 'preset' && preset"
      class="fb-avatar-badge-preset"
      v-html="preset.svg"
    ></div>
    <span v-else class="fb-avatar-badge-initials">{{ initials }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { avatarURL } from "@/api/users";
import { getAvatarPreset } from "@/utils/avatarPresets";
import { avatarBust } from "@/utils/avatarBust";

const props = withDefaults(
  defineProps<{
    user?: IUser | null;
    size?: number;
  }>(),
  { size: 30 }
);

const kind = computed<"custom" | "preset" | "initials">(() => {
  const avatar = props.user?.avatar;
  if (avatar === "custom") return "custom";
  if (avatar?.startsWith("preset:")) return "preset";
  return "initials";
});

const preset = computed(() => {
  const avatar = props.user?.avatar;
  if (!avatar?.startsWith("preset:")) return undefined;
  return getAvatarPreset(avatar.slice("preset:".length));
});

const src = computed(() =>
  props.user ? avatarURL(props.user.id, avatarBust.value) : ""
);

const initials = computed(() => {
  const name = props.user?.displayName || props.user?.username || "";
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length >= 2) {
    return (parts[0][0] + parts[1][0]).toUpperCase();
  }
  return name.slice(0, 2).toUpperCase() || "?";
});

const badgeStyle = computed(() => ({
  width: `${props.size}px`,
  height: `${props.size}px`,
  fontSize: `${Math.round(props.size * 0.43)}px`,
}));
</script>

<style scoped>
.fb-avatar-badge {
  border-radius: 50%;
  background: var(--accent);
  color: var(--on-accent);
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  user-select: none;
  overflow: hidden;
}

.fb-avatar-badge-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.fb-avatar-badge-preset {
  width: 100%;
  height: 100%;
  display: flex;
}

.fb-avatar-badge-preset :deep(svg) {
  width: 100%;
  height: 100%;
}
</style>
