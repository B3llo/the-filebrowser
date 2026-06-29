<template>
  <errors v-if="error" :errorCode="error.status" />
  <div class="fb-settings-section" v-else-if="!layoutStore.loading">
    <div class="dashboard row">
      <div class="column">
        <form @submit.prevent="save" class="card">
          <div class="card-title">
            <h2 v-if="isNew">{{ t("settings.newSource") }}</h2>
            <h2 v-else>{{ t("settings.source") }} {{ original?.name }}</h2>
          </div>

          <div class="card-content">
            <div class="fb-settings-form">
              <div class="fb-settings-field">
                <label class="fb-settings-field-label" for="name">{{
                  t("settings.sourceName")
                }}</label>
                <input
                  class="input input--block"
                  type="text"
                  v-model="source.name"
                  id="name"
                />
              </div>

              <div class="fb-settings-field">
                <label class="fb-settings-field-label" for="path">{{
                  t("settings.sourcePath")
                }}</label>
                <input
                  class="input input--block"
                  type="text"
                  placeholder="/absolute/path"
                  v-model="source.path"
                  id="path"
                />
                <p class="small" style="margin-top: 6px">
                  {{ t("settings.sourcePathHelp") }}
                </p>
              </div>
            </div>
          </div>

          <div class="card-action">
            <button
              v-if="!isNew"
              @click.prevent="deletePrompt"
              type="button"
              class="button button--flat button--red"
              :aria-label="t('buttons.delete')"
              :title="t('buttons.delete')"
            >
              {{ t("buttons.delete") }}
            </button>
            <router-link to="/settings/sources">
              <button
                type="button"
                class="button button--flat button--grey"
                :aria-label="t('buttons.cancel')"
                :title="t('buttons.cancel')"
              >
                {{ t("buttons.cancel") }}
              </button>
            </router-link>
            <input class="button" type="submit" :value="t('buttons.save')" />
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useLayoutStore } from "@/stores/layout";
import { useSourceStore } from "@/stores/source";
import { sources as api } from "@/api";
import Errors from "@/views/Errors.vue";
import { computed, inject, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { StatusError } from "@/api/utils";

const error = ref<StatusError>();
const original = ref<ISource>();
const source = ref<{ name: string; path: string }>({ name: "", path: "" });

const $showError = inject<IToastError>("$showError")!;
const $showSuccess = inject<IToastSuccess>("$showSuccess")!;

const layoutStore = useLayoutStore();
const sourceStore = useSourceStore();
const route = useRoute();
const router = useRouter();
const { t } = useI18n();

const isNew = computed(() => route.path === "/settings/sources/new");

onMounted(() => fetchData());
watch(route, () => fetchData());

const fetchData = async () => {
  layoutStore.loading = true;
  try {
    if (isNew.value) {
      source.value = { name: "", path: "" };
    } else {
      const id = Array.isArray(route.params.id)
        ? route.params.id.join("")
        : route.params.id;
      const data = await api.get(parseInt(id));
      original.value = data;
      source.value = { name: data.name, path: data.path ?? "" };
    }
  } catch (err) {
    if (err instanceof Error) {
      error.value = err;
    }
  } finally {
    layoutStore.loading = false;
  }
};

const save = async () => {
  if (!source.value.name || !source.value.path) {
    $showError(t("settings.sourceFieldsRequired"));
    return;
  }
  try {
    if (isNew.value) {
      await api.create({ name: source.value.name, path: source.value.path });
      $showSuccess(t("settings.sourceCreated"));
    } else {
      await api.update(original.value!.id, ["Name", "Path"], {
        id: original.value!.id,
        name: source.value.name,
        path: source.value.path,
      });
      $showSuccess(t("settings.sourceUpdated"));
    }
    // Refresh the cached source list so the sidebar and switcher reflect the
    // change without a full page reload.
    await sourceStore.load();
    router.push({ path: "/settings/sources" });
  } catch (err: any) {
    $showError(err);
  }
};

const deletePrompt = () => {
  layoutStore.showHover({
    prompt: "deleteSource",
    confirm: () => {
      layoutStore.closeHovers();
      deleteSource();
    },
  });
};

const deleteSource = async () => {
  try {
    await api.remove(original.value!.id);
    await sourceStore.load();
    $showSuccess(t("settings.sourceDeleted"));
    router.push({ path: "/settings/sources" });
  } catch (err) {
    if (err instanceof StatusError && err.status === 403) {
      $showError(t("errors.forbidden"));
    } else if (err instanceof Error) {
      $showError(err);
    }
  }
};
</script>
