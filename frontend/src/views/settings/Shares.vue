<template>
  <errors v-if="error" :errorCode="error.status" />
  <div class="fb-settings-section" v-else-if="!layoutStore.loading">
    <div class="fb-settings-stack">
      <div class="card">
        <div class="card-title">
          <h2>{{ t("settings.shareManagement") }}</h2>
        </div>

        <div class="card-content full" v-if="links.length > 0">
          <table>
            <thead>
              <tr>
                <th>{{ t("settings.path") }}</th>
                <th>{{ t("settings.shareDuration") }}</th>
                <th v-if="authStore.user?.perm.admin">
                  {{ t("settings.username") }}
                </th>
                <th></th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="link in links" :key="link.hash">
                <td>
                  <a
                    :href="buildLink(link)"
                    target="_blank"
                    class="fb-settings-table-link"
                    >{{ link.path }}</a
                  >
                </td>
                <td>
                  <template v-if="link.expire !== 0">{{
                    humanTime(link.expire)
                  }}</template>
                  <template v-else>{{ t("permanent") }}</template>
                </td>
                <td v-if="authStore.user?.perm.admin">{{ link.username }}</td>
                <td>
                  <button
                    class="button button--icon button--ghost"
                    @click="deleteLink($event, link)"
                    :aria-label="t('buttons.delete')"
                    :title="t('buttons.delete')"
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
                      <path d="M3 6h18" />
                      <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6" />
                      <path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
                      <line x1="10" y1="11" x2="10" y2="17" />
                      <line x1="14" y1="11" x2="14" y2="17" />
                    </svg>
                  </button>
                </td>
                <td>
                  <button
                    class="button button--icon button--ghost copy-clipboard"
                    :aria-label="t('buttons.copyToClipboard')"
                    :title="t('buttons.copyToClipboard')"
                    @click="copyToClipboard(buildLink(link))"
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
                      <rect x="9" y="9" width="13" height="13" rx="2" />
                      <path
                        d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"
                      />
                    </svg>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
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
              <circle cx="18" cy="5" r="3" />
              <circle cx="6" cy="12" r="3" />
              <circle cx="18" cy="19" r="3" />
              <line x1="8.59" y1="13.51" x2="15.42" y2="17.49" />
              <line x1="15.41" y1="6.51" x2="8.59" y2="10.49" />
            </svg>
          </div>
          <h3 class="fb-settings-empty-title">{{ t("files.lonely") }}</h3>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import { share as api, users } from "@/api";
import dayjs from "dayjs";
import Errors from "@/views/Errors.vue";
import { inject, ref, onMounted } from "vue";
import { useI18n } from "vue-i18n";
import { StatusError } from "@/api/utils";
import { copy } from "@/utils/clipboard";

const $showError = inject<IToastError>("$showError")!;
const $showSuccess = inject<IToastSuccess>("$showSuccess")!;
const { t } = useI18n();

const layoutStore = useLayoutStore();
const authStore = useAuthStore();

const error = ref<StatusError | null>(null);
const links = ref<Share[]>([]);

onMounted(async () => {
  layoutStore.loading = true;

  try {
    const newLinks = await api.list();
    if (authStore.user?.perm.admin) {
      const userMap = new Map<number, string>();
      for (const user of await users.getAll())
        userMap.set(user.id, user.username);
      for (const link of newLinks) {
        if (link.userID && userMap.has(link.userID))
          link.username = userMap.get(link.userID);
      }
    }
    links.value = newLinks;
  } catch (err) {
    if (err instanceof Error) {
      error.value = err;
    }
  } finally {
    layoutStore.loading = false;
  }
});

const copyToClipboard = (text: string) => {
  copy({ text }).then(
    () => {
      $showSuccess(t("success.linkCopied"));
    },
    () => {
      copy({ text }, { permission: true }).then(
        () => {
          $showSuccess(t("success.linkCopied"));
        },
        (e) => {
          $showError(e);
        }
      );
    }
  );
};

const deleteLink = async (event: Event, link: any) => {
  event.preventDefault();

  layoutStore.showHover({
    prompt: "share-delete",
    confirm: () => {
      layoutStore.closeHovers();

      try {
        api.remove(link.hash);
        links.value = links.value.filter((item) => item.hash !== link.hash);
        $showSuccess(t("settings.shareDeleted"));
      } catch (err) {
        if (err instanceof Error) {
          $showError(err);
        }
      }
    },
  });
};

const humanTime = (time: number) => {
  return dayjs(time * 1000).fromNow();
};

const buildLink = (share: Share) => {
  return api.getShareURL(share);
};
</script>
