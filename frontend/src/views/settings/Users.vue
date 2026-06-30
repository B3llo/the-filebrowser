<template>
  <errors v-if="error" :errorCode="error.status" />
  <div class="fb-settings-section" v-else-if="!layoutStore.loading">
    <div class="fb-settings-stack">
      <div class="card">
        <div class="card-title">
          <h2>{{ t("settings.users") }}</h2>
          <router-link to="/settings/users/new">
            <button class="button">
              {{ t("buttons.new") }}
            </button>
          </router-link>
        </div>

        <div class="card-content full">
          <table v-if="users.length > 0">
            <thead>
              <tr>
                <th>{{ t("settings.username") }}</th>
                <th>{{ t("settings.admin") }}</th>
                <th>{{ t("settings.scope") }}</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="user in users" :key="user.id">
                <td>{{ user.username }}</td>
                <td>
                  <svg
                    v-if="user.perm.admin"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="var(--accent)"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    style="width: 18px; height: 18px"
                  >
                    <polyline points="20 6 9 17 4 12" />
                  </svg>
                  <svg
                    v-else
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="var(--faint)"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    style="width: 18px; height: 18px"
                  >
                    <line x1="18" y1="6" x2="6" y2="18" />
                    <line x1="6" y1="6" x2="18" y2="18" />
                  </svg>
                </td>
                <td>{{ user.scope }}</td>
                <td>
                  <router-link
                    :to="'/settings/users/' + user.id"
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
                <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
                <circle cx="9" cy="7" r="4" />
                <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
                <path d="M16 3.13a4 4 0 0 1 0 7.75" />
              </svg>
            </div>
            <h3 class="fb-settings-empty-title">{{ t("files.lonely") }}</h3>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useLayoutStore } from "@/stores/layout";
import { users as api } from "@/api";
import Errors from "@/views/Errors.vue";
import { onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { StatusError } from "@/api/utils";

const error = ref<StatusError | null>(null);
const users = ref<IUser[]>([]);

const layoutStore = useLayoutStore();
const { t } = useI18n();

onMounted(async () => {
  layoutStore.loading = true;

  try {
    users.value = await api.getAll();
  } catch (err) {
    if (err instanceof Error) {
      error.value = err;
    }
  } finally {
    layoutStore.loading = false;
  }
});
</script>
