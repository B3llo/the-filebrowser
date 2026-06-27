<template>
  <div class="fb-settings-form">
    <div v-if="!isDefault && props.user !== null" class="fb-settings-field">
      <label class="fb-settings-field-label" for="username">{{ t("settings.username") }}</label>
      <input
        class="input input--block"
        type="text"
        v-model="user.username"
        id="username"
      />
    </div>

    <div v-if="!isDefault" class="fb-settings-field">
      <label class="fb-settings-field-label" for="password">{{ t("settings.password") }}</label>
      <input
        class="input input--block"
        type="password"
        :placeholder="passwordPlaceholder"
        v-model="user.password"
        id="password"
      />
    </div>

    <div class="fb-settings-field">
      <label class="fb-settings-field-label" for="scope">{{ t("settings.scope") }}</label>
      <input
        :disabled="createUserDirData ?? false"
        :placeholder="scopePlaceholder"
        class="input input--block"
        type="text"
        v-model="user.scope"
        id="scope"
      />
    </div>

    <div v-if="displayHomeDirectoryCheckbox" class="fb-settings-checkbox-item" style="background: transparent; padding: 0">
      <input type="checkbox" v-model="createUserDirData" id="createUserDir" />
      <label for="createUserDir">{{ t("settings.createUserHomeDirectory") }}</label>
    </div>

    <div class="fb-settings-field">
      <label class="fb-settings-field-label" for="locale">{{ t("settings.language") }}</label>
      <languages
        class="input input--block input--select"
        id="locale"
        v-model:locale="user.locale"
      ></languages>
    </div>

    <div v-if="!isDefault && user.perm" class="fb-settings-checkbox-item" style="background: transparent; padding: 0">
      <input
        type="checkbox"
        :disabled="user.perm.admin"
        v-model="user.lockPassword"
        id="lockPassword"
      />
      <label for="lockPassword">{{ t("settings.lockPassword") }}</label>
    </div>

    <permissions v-model:perm="user.perm" />
    <commands v-if="enableExec" v-model:commands="user.commands" />

    <div v-if="!isDefault">
      <h3 style="margin: 20px 0 8px; font-size: 14px; font-weight: 600; color: var(--text)">{{ t("settings.rules") }}</h3>
      <p class="small" style="margin-bottom: 12px">{{ t("settings.rulesHelp") }}</p>
      <rules v-model:rules="user.rules" />
    </div>
  </div>
</template>

<script setup lang="ts">
import Languages from "./Languages.vue";
import Rules from "./Rules.vue";
import Permissions from "./Permissions.vue";
import Commands from "./Commands.vue";
import { enableExec } from "@/utils/constants";
import { computed, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();

const createUserDirData = ref<boolean | null>(null);
const originalUserScope = ref<string | null>(null);

const props = defineProps<{
  user: IUserForm;
  isNew: boolean;
  isDefault: boolean;
  createUserDir?: boolean;
}>();

onMounted(() => {
  if (props.user.scope) {
    originalUserScope.value = props.user.scope;
    createUserDirData.value = props.createUserDir;
  }
});

const passwordPlaceholder = computed(() =>
  props.isNew ? "" : t("settings.avoidChanges")
);
const scopePlaceholder = computed(() =>
  createUserDirData.value ? t("settings.userScopeGenerationPlaceholder") : ""
);
const displayHomeDirectoryCheckbox = computed(
  () => props.isNew && createUserDirData.value
);

watch(
  () => props.user,
  () => {
    if (!props.user?.perm?.admin) return;
    props.user.lockPassword = false;
  }
);

watch(createUserDirData, () => {
  if (props.user?.scope) {
    props.user.scope = createUserDirData.value
      ? ""
      : (originalUserScope.value ?? "");
  }
});
</script>
