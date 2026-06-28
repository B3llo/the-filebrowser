<template>
  <div>
    <h3
      style="
        margin: 20px 0 8px;
        font-size: 14px;
        font-weight: 600;
        color: var(--text);
      "
    >
      {{ $t("settings.permissions") }}
    </h3>
    <p class="small" style="margin-bottom: 12px">
      {{ $t("settings.permissionsHelp") }}
    </p>

    <div class="fb-settings-perm-grid">
      <div class="fb-settings-perm-item">
        <input type="checkbox" v-model="admin" id="perm-admin" />
        <label for="perm-admin">{{ $t("settings.administrator") }}</label>
      </div>

      <div class="fb-settings-perm-item">
        <input
          type="checkbox"
          :disabled="admin"
          v-model="perm.create"
          id="perm-create"
        />
        <label for="perm-create">{{ $t("settings.perm.create") }}</label>
      </div>

      <div class="fb-settings-perm-item">
        <input
          type="checkbox"
          :disabled="admin"
          v-model="perm.delete"
          id="perm-delete"
        />
        <label for="perm-delete">{{ $t("settings.perm.delete") }}</label>
      </div>

      <div class="fb-settings-perm-item">
        <input
          type="checkbox"
          :disabled="admin || perm.share"
          v-model="perm.download"
          id="perm-download"
        />
        <label for="perm-download">{{ $t("settings.perm.download") }}</label>
      </div>

      <div class="fb-settings-perm-item">
        <input
          type="checkbox"
          :disabled="admin"
          v-model="perm.modify"
          id="perm-modify"
        />
        <label for="perm-modify">{{ $t("settings.perm.modify") }}</label>
      </div>

      <div v-if="isExecEnabled" class="fb-settings-perm-item">
        <input
          type="checkbox"
          :disabled="admin"
          v-model="perm.execute"
          id="perm-execute"
        />
        <label for="perm-execute">{{ $t("settings.perm.execute") }}</label>
      </div>

      <div class="fb-settings-perm-item">
        <input
          type="checkbox"
          :disabled="admin"
          v-model="perm.rename"
          id="perm-rename"
        />
        <label for="perm-rename">{{ $t("settings.perm.rename") }}</label>
      </div>

      <div class="fb-settings-perm-item">
        <input
          type="checkbox"
          :disabled="admin"
          v-model="perm.share"
          id="perm-share"
        />
        <label for="perm-share">{{ $t("settings.perm.share") }}</label>
      </div>
    </div>
  </div>
</template>

<script>
import { enableExec } from "@/utils/constants";
export default {
  name: "permissions",
  props: ["perm"],
  computed: {
    admin: {
      get() {
        return this.perm.admin;
      },
      set(value) {
        if (value) {
          for (const key in this.perm) {
            this.perm[key] = true;
          }
        }

        this.perm.admin = value;
      },
    },
    isExecEnabled: () => enableExec,
  },
  watch: {
    perm: {
      deep: true,
      handler() {
        if (this.perm.share === true) {
          this.perm.download = true;
        }
      },
    },
  },
};
</script>
