<template>
  <div class="fb-avatar-picker">
    <div class="fb-avatar-picker-preview">
      <AvatarBadge :user="previewUser" :size="88" />
      <div class="fb-avatar-picker-preview-actions">
        <button
          type="button"
          class="fb-avatar-picker-upload"
          @click="triggerUpload"
        >
          {{ t("settings.uploadPhoto") }}
        </button>
        <button
          v-if="currentAvatar"
          type="button"
          class="fb-avatar-picker-remove"
          @click="emit('remove')"
        >
          {{ t("settings.removeAvatar") }}
        </button>
      </div>
      <input
        ref="fileInput"
        type="file"
        accept="image/*"
        style="display: none"
        @change="onFileChange"
      />
    </div>

    <div class="fb-avatar-picker-grid">
      <button
        v-for="preset in presets"
        :key="preset.id"
        type="button"
        class="fb-avatar-picker-btn"
        :class="{ active: currentAvatar === `preset:${preset.id}` }"
        :title="t(`avatarPresets.${preset.id}`)"
        :aria-label="t(`avatarPresets.${preset.id}`)"
        @click="emit('select', preset.id)"
      >
        <!-- preset.svg is static, developer-authored markup (see avatarPresets.ts) -->
        <div class="fb-avatar-picker-icon" v-html="preset.svg"></div>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useI18n } from "vue-i18n";
import { avatarPresets } from "@/utils/avatarPresets";
import AvatarBadge from "@/components/AvatarBadge.vue";

const { t } = useI18n();

const props = defineProps<{
  currentAvatar?: string;
  user?: IUser | null;
}>();

const emit = defineEmits<{
  (e: "select", id: string): void;
  (e: "upload", file: File): void;
  (e: "remove"): void;
}>();

const presets = avatarPresets;
const fileInput = ref<HTMLInputElement | null>(null);

const previewUser = computed(() => {
  if (!props.user) return props.user;
  return { ...props.user, avatar: props.currentAvatar ?? props.user.avatar };
});

const triggerUpload = () => {
  fileInput.value?.click();
};

const onFileChange = (event: Event) => {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  if (file) {
    emit("upload", file);
  }
  input.value = "";
};
</script>

<style scoped>
.fb-avatar-picker {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.fb-avatar-picker-preview {
  display: flex;
  align-items: center;
  gap: 16px;
}

.fb-avatar-picker-preview-actions {
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
  align-items: flex-start;
  gap: 8px;
}

.fb-avatar-picker-upload,
.fb-avatar-picker-remove {
  flex: none;
  padding: 7px 14px;
  border-radius: 8px;
  border: 1px solid var(--border);
  font-size: 13px;
  cursor: pointer;
  text-align: left;
  transition:
    background 0.15s ease,
    color 0.15s ease;
}

.fb-avatar-picker-upload {
  background: var(--accent);
  color: var(--on-accent);
  border-color: transparent;
}

.fb-avatar-picker-upload:hover {
  background: var(--accent-press);
}

.fb-avatar-picker-remove {
  background: var(--hover);
  color: var(--dim);
}

.fb-avatar-picker-remove:hover {
  background: var(--border);
  color: var(--text);
}

.fb-avatar-picker-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.fb-avatar-picker-btn {
  width: 52px;
  height: 52px;
  flex: none;
  border-radius: 50%;
  border: 2px solid transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  overflow: hidden;
  transition:
    transform 0.15s ease,
    box-shadow 0.15s ease;
}

.fb-avatar-picker-btn:hover {
  transform: scale(1.08);
}

.fb-avatar-picker-btn.active {
  box-shadow:
    0 0 0 2px var(--surface),
    0 0 0 4px var(--accent);
}

.fb-avatar-picker-icon {
  width: 100%;
  height: 100%;
  display: flex;
}

.fb-avatar-picker-icon :deep(svg) {
  width: 100%;
  height: 100%;
}
</style>
