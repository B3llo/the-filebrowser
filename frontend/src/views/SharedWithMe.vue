<template>
  <errors v-if="error" :errorCode="error.status" />
  <div class="fb-settings-section" v-else-if="!layoutStore.loading">
    <div class="fb-settings-stack">
      <div class="card">
        <div class="card-title">
          <h2>{{ t("grants.sharedWithMe") }}</h2>
        </div>

        <div class="card-content" v-if="grants.length > 0">
          <div class="fb-local-search">
            <FbIcon name="search" size="16px" />
            <input
              v-model.trim="searchQuery"
              :placeholder="t('buttons.search')"
              type="text"
            />
          </div>
        </div>

        <div class="card-content full" v-if="filtered.length > 0">
          <table>
            <thead>
              <tr>
                <th>{{ t("settings.path") }}</th>
                <th>{{ t("grants.owner") }}</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="grant in filtered"
                :key="grant.id"
                class="fb-grant-row"
                @click="openGrant(grant)"
              >
                <td>
                  <span class="fb-settings-table-link">{{
                    grantDisplayName(grant.path)
                  }}</span>
                  <span class="fb-grant-path">{{ grant.path }}</span>
                </td>
                <td>{{ grant.ownerUsername || `#${grant.ownerID}` }}</td>
                <td>
                  <span class="fb-grant-role">{{
                    roleLabel(grant.role, t)
                  }}</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-else-if="grants.length === 0" class="fb-settings-empty">
          <div class="fb-settings-empty-icon">
            <FbIcon name="share" size="40px" />
          </div>
          <h3 class="fb-settings-empty-title">
            {{ t("files.emptySharedTitle") }}
          </h3>
          <p class="fb-settings-empty-sub">{{ t("files.emptySharedSub") }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { useLayoutStore } from "@/stores/layout";
import { useSourceStore } from "@/stores/source";
import { grants as api } from "@/api";
import { StatusError } from "@/api/utils";
import Errors from "@/views/Errors.vue";
import FbIcon from "@/components/FbIcon.vue";
import {
  filterGrants,
  grantDisplayName,
  grantTarget,
  roleLabel,
} from "@/utils/grants";

const $showError = inject<IToastError>("$showError")!;
const { t } = useI18n();
const router = useRouter();
const layoutStore = useLayoutStore();
const sourceStore = useSourceStore();

const error = ref<StatusError | null>(null);
const grants = ref<Grant[]>([]);
const searchQuery = ref("");

const filtered = computed(() => filterGrants(grants.value, searchQuery.value));

onMounted(async () => {
  layoutStore.loading = true;
  try {
    grants.value = await api.sharedWithMe();
  } catch (err) {
    if (err instanceof Error) {
      if (err instanceof StatusError) error.value = err;
      else $showError(err);
    }
  } finally {
    layoutStore.loading = false;
  }
});

const openGrant = (grant: Grant) => {
  router.push({ path: grantTarget(sourceStore.filesBase, grant.path) });
};
</script>
