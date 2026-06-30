<template>
  <errors v-if="error" :errorCode="error.status" />
  <div class="fb-settings-section" v-else-if="!layoutStore.loading">
    <div class="fb-settings-stack">
      <div class="card">
        <div class="card-title">
          <h2>{{ t("settings.sources") }}</h2>
          <router-link to="/settings/sources/new">
            <button class="button">
              {{ t("buttons.new") }}
            </button>
          </router-link>
        </div>

        <div class="card-content full">
          <table v-if="sources.length > 0">
            <thead>
              <tr>
                <th>{{ t("settings.sourceName") }}</th>
                <th>{{ t("settings.sourcePath") }}</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="source in sources" :key="source.id">
                <td>{{ source.name }}</td>
                <td>{{ source.path }}</td>
                <td>
                  <router-link
                    :to="'/settings/sources/' + source.id"
                    class="fb-settings-table-link"
                  >
                    <svg
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      style="width: 16px; height: 16px"
                    >
                      <path d="M12 20h9" />
                      <path d="M16.5 3.5a2.12 2.12 0 0 1 3 3L7 19l-4 1 1-4z" />
                    </svg>
                  </router-link>
                </td>
              </tr>
            </tbody>
          </table>
          <div v-else class="fb-settings-empty">
            <div class="fb-settings-empty-icon">
              <svg
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="1.5"
                stroke-linecap="round"
                stroke-linejoin="round"
                style="width: 40px; height: 40px"
              >
                <path
                  d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"
                />
              </svg>
            </div>
            <h3 class="fb-settings-empty-title">
              {{ t("settings.noSources") }}
            </h3>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useLayoutStore } from "@/stores/layout";
import { sources as api } from "@/api";
import Errors from "@/views/Errors.vue";
import { onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { StatusError } from "@/api/utils";

const error = ref<StatusError | null>(null);
const sources = ref<ISource[]>([]);

const layoutStore = useLayoutStore();
const { t } = useI18n();

onMounted(async () => {
  layoutStore.loading = true;

  try {
    const all = await api.list();
    // The implicit legacy source (id 0) is virtual and cannot be edited; only
    // show real, admin-defined sources here.
    sources.value = all.filter((s) => s.id !== 0);
  } catch (err) {
    if (err instanceof Error) {
      error.value = err;
    }
  } finally {
    layoutStore.loading = false;
  }
});
</script>
