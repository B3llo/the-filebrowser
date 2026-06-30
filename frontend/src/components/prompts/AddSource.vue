<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ t("settings.newSource") }}</h2>
    </div>

    <div class="card-content">
      <div class="fb-settings-form">
        <div class="fb-settings-field">
          <label class="fb-settings-field-label" for="add-source-name">{{
            t("settings.sourceName")
          }}</label>
          <input
            id="add-source-name"
            class="input input--block"
            type="text"
            v-model="name"
            autofocus
          />
        </div>

        <div class="fb-settings-field">
          <label class="fb-settings-field-label" for="add-source-path">{{
            t("settings.sourcePath")
          }}</label>
          <div class="fb-add-source-path-row">
            <input
              id="add-source-path"
              class="input"
              style="flex: 1; min-width: 0"
              type="text"
              placeholder="/absolute/path"
              v-model="path"
            />
            <button
              type="button"
              class="button button--flat button--grey"
              style="white-space: nowrap"
              @click="showPicker = !showPicker"
            >
              {{ t("sidebar.browseFolder") }}
            </button>
          </div>
          <p class="small" style="margin-top: 6px">
            {{ t("settings.sourcePathHelp") }}
          </p>
        </div>

        <directory-picker
          v-if="showPicker"
          v-model="path"
          class="fb-add-source-picker"
        />
      </div>
    </div>

    <div class="card-action">
      <button
        class="button button--flat button--grey"
        @click="layoutStore.closeHovers()"
        :aria-label="t('buttons.cancel')"
      >
        {{ t("buttons.cancel") }}
      </button>
      <button
        class="button button--flat"
        :disabled="saving || !name || !path"
        @click="save"
      >
        {{ saving ? t("sidebar.adding") : t("buttons.save") }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, inject } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { useLayoutStore } from "@/stores/layout";
import { useSourceStore } from "@/stores/source";
import { sources as api } from "@/api";
import DirectoryPicker from "@/components/sidebar/DirectoryPicker.vue";

const { t } = useI18n();
const layoutStore = useLayoutStore();
const sourceStore = useSourceStore();
const router = useRouter();
const $showError = inject<IToastError>("$showError")!;
const $showSuccess = inject<IToastSuccess>("$showSuccess")!;

const name = ref("");
const path = ref("/");
const saving = ref(false);
const showPicker = ref(false);

const save = async () => {
  if (!name.value || !path.value) return;
  saving.value = true;
  try {
    await api.create({ name: name.value, path: path.value });
    await sourceStore.load();
    $showSuccess(t("settings.sourceCreated"));
    layoutStore.closeHovers();
    router.push("/settings/sources");
  } catch (err: any) {
    $showError(err);
  } finally {
    saving.value = false;
  }
};
</script>

<style scoped>
.fb-add-source-path-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.fb-add-source-picker {
  margin-top: 4px;
}
</style>
