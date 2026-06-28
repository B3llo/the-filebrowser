<template>
  <errors v-if="error" :errorCode="error.status" />
  <div
    class="fb-settings-section"
    v-else-if="!layoutStore.loading && settings !== null"
  >
    <div class="dashboard row">
      <div class="column">
        <form class="card" @submit.prevent="save">
          <div class="card-title">
            <h2>{{ t("settings.globalSettings") }}</h2>
          </div>

          <div class="card-content">
            <div class="fb-settings-checkbox-list">
              <div class="fb-settings-checkbox-item">
                <input type="checkbox" v-model="settings.signup" id="signup" />
                <label for="signup">{{ t("settings.allowSignup") }}</label>
              </div>
              <div class="fb-settings-checkbox-item">
                <input
                  type="checkbox"
                  v-model="settings.createUserDir"
                  id="createUserDir"
                />
                <label for="createUserDir">{{
                  t("settings.createUserDir")
                }}</label>
              </div>
              <div class="fb-settings-checkbox-item">
                <input
                  type="checkbox"
                  v-model="settings.hideLoginButton"
                  id="hideLoginButton"
                />
                <label for="hideLoginButton">{{
                  t("settings.hideLoginButton")
                }}</label>
              </div>
            </div>

            <div class="fb-settings-divider"></div>

            <div class="fb-settings-field">
              <label class="fb-settings-field-label" for="userHomeBasePath">{{
                t("settings.userHomeBasePath")
              }}</label>
              <input
                class="input input--block"
                type="text"
                v-model="settings.userHomeBasePath"
                id="userHomeBasePath"
              />
            </div>

            <div class="fb-settings-field">
              <label
                class="fb-settings-field-label"
                for="minimumPasswordLength"
                >{{ t("settings.minimumPasswordLength") }}</label
              >
              <vue-number-input
                controls
                v-model.number="settings.minimumPasswordLength"
                id="minimumPasswordLength"
                :min="1"
              />
            </div>

            <div class="fb-settings-divider"></div>

            <h3>{{ t("settings.rules") }}</h3>
            <p class="small">{{ t("settings.globalRules") }}</p>
            <rules v-model:rules="settings.rules" />

            <div
              v-if="enableExec"
              class="fb-settings-field"
              style="margin-top: 20px"
            >
              <h3>{{ t("settings.executeOnShell") }}</h3>
              <p class="small">{{ t("settings.executeOnShellDescription") }}</p>
              <input
                class="input input--block"
                type="text"
                placeholder="bash -c, cmd /c, ..."
                v-model="shellValue"
              />
            </div>

            <div class="fb-settings-divider"></div>

            <h3>{{ t("settings.branding") }}</h3>

            <i18n-t
              keypath="settings.brandingHelp"
              tag="p"
              class="small"
              scope="global"
            >
              <a
                class="link"
                target="_blank"
                href="https://filebrowser.org/customization.html#custom-branding"
                >{{ t("settings.documentation") }}</a
              >
            </i18n-t>

            <div class="fb-settings-checkbox-list" style="margin-top: 12px">
              <div class="fb-settings-checkbox-item">
                <input
                  type="checkbox"
                  v-model="settings.branding.disableExternal"
                  id="branding-links"
                />
                <label for="branding-links">{{
                  t("settings.disableExternalLinks")
                }}</label>
              </div>
              <div class="fb-settings-checkbox-item">
                <input
                  type="checkbox"
                  v-model="settings.branding.disableUsedPercentage"
                  id="branding-used-disk"
                />
                <label for="branding-used-disk">{{
                  t("settings.disableUsedDiskPercentage")
                }}</label>
              </div>
            </div>

            <div class="fb-settings-field" style="margin-top: 16px">
              <label class="fb-settings-field-label" for="theme">{{
                t("settings.themes.title")
              }}</label>
              <themes
                class="input input--block input--select"
                v-model:theme="settings.branding.theme"
                id="theme"
              ></themes>
            </div>

            <div class="fb-settings-field">
              <label class="fb-settings-field-label" for="branding-name">{{
                t("settings.instanceName")
              }}</label>
              <input
                class="input input--block"
                type="text"
                v-model="settings.branding.name"
                id="branding-name"
              />
            </div>

            <div class="fb-settings-field">
              <label class="fb-settings-field-label" for="branding-files">{{
                t("settings.brandingDirectoryPath")
              }}</label>
              <input
                class="input input--block"
                type="text"
                v-model="settings.branding.files"
                id="branding-files"
              />
            </div>

            <div class="fb-settings-divider"></div>

            <h3>{{ t("settings.tusUploads") }}</h3>

            <p class="small">{{ t("settings.tusUploadsHelp") }}</p>

            <div class="tusConditionalSettings" style="margin-top: 12px">
              <div class="fb-settings-field">
                <label class="fb-settings-field-label" for="tus-chunkSize">{{
                  t("settings.tusUploadsChunkSize")
                }}</label>
                <input
                  class="input input--block"
                  type="text"
                  v-model="formattedChunkSize"
                  id="tus-chunkSize"
                />
              </div>

              <div class="fb-settings-field">
                <label class="fb-settings-field-label" for="tus-retryCount">{{
                  t("settings.tusUploadsRetryCount")
                }}</label>
                <vue-number-input
                  controls
                  v-model.number="settings.tus.retryCount"
                  id="tus-retryCount"
                  :min="0"
                />
              </div>
            </div>
          </div>

          <div class="card-action">
            <input class="button" type="submit" :value="t('buttons.update')" />
          </div>
        </form>
      </div>

      <div class="column">
        <form class="card" @submit.prevent="save">
          <div class="card-title">
            <h2>{{ t("settings.userDefaults") }}</h2>
          </div>

          <div class="card-content">
            <p class="small">{{ t("settings.defaultUserDescription") }}</p>

            <user-form
              :isNew="false"
              :isDefault="true"
              v-model:user="settings.defaults"
            />
          </div>

          <div class="card-action">
            <input class="button" type="submit" :value="t('buttons.update')" />
          </div>
        </form>
      </div>

      <div class="column" v-if="enableExec">
        <form class="card" @submit.prevent="save">
          <div class="card-title">
            <h2>{{ t("settings.commandRunner") }}</h2>
          </div>

          <div class="card-content">
            <i18n-t
              keypath="settings.commandRunnerHelp"
              tag="p"
              class="small"
              scope="global"
            >
              <code>FILE</code>
              <code>SCOPE</code>
              <a
                class="link"
                target="_blank"
                href="https://filebrowser.org/command-execution.html#hook-runner"
                >{{ t("settings.documentation") }}</a
              >
            </i18n-t>

            <div
              v-for="(command, key) in settings.commands"
              :key="key"
              class="collapsible"
            >
              <input :id="key" type="checkbox" />
              <label :for="key">
                <p
                  style="
                    margin: 0;
                    font-size: 14px;
                    font-weight: 550;
                    color: var(--text);
                  "
                >
                  {{ capitalize(key) }}
                </p>
                <svg
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  style="width: 16px; height: 16px; color: var(--dim)"
                >
                  <polyline points="6 9 12 15 18 9" />
                </svg>
              </label>
              <div class="collapse">
                <textarea
                  class="input input--block input--textarea"
                  v-model.trim="commandObject[key]"
                ></textarea>
              </div>
            </div>
          </div>

          <div class="card-action">
            <input class="button" type="submit" :value="t('buttons.update')" />
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { settings as api } from "@/api";
import { StatusError } from "@/api/utils";
import Rules from "@/components/settings/Rules.vue";
import Themes from "@/components/settings/Themes.vue";
import UserForm from "@/components/settings/UserForm.vue";
import { useLayoutStore } from "@/stores/layout";
import { enableExec } from "@/utils/constants";
import { getTheme, setTheme } from "@/utils/theme";
import Errors from "@/views/Errors.vue";
import { computed, inject, onBeforeUnmount, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";

const error = ref<StatusError | null>(null);
const originalSettings = ref<ISettings | null>(null);
const settings = ref<ISettings | null>(null);
const debounceTimeout = ref<number | null>(null);

const commandObject = ref<{
  [key: string]: string[] | string;
}>({});
const shellValue = ref<string>("");

const $showError = inject<IToastError>("$showError")!;
const $showSuccess = inject<IToastSuccess>("$showSuccess")!;

const { t } = useI18n();

const layoutStore = useLayoutStore();

const formattedChunkSize = computed({
  get() {
    return settings?.value?.tus?.chunkSize
      ? formatBytes(settings?.value?.tus?.chunkSize)
      : "";
  },
  set(value: string) {
    if (debounceTimeout.value) {
      clearTimeout(debounceTimeout.value);
    }

    debounceTimeout.value = window.setTimeout(() => {
      if (settings.value) settings.value.tus.chunkSize = parseBytes(value);
    }, 1500);
  },
});

const capitalize = (name: string, where: string | RegExp = "_") => {
  if (where === "caps") where = /(?=[A-Z])/;
  const split = name.split(where);
  name = "";

  for (let i = 0; i < split.length; i++) {
    name += split[i].charAt(0).toUpperCase() + split[i].slice(1) + " ";
  }

  return name.slice(0, -1);
};

const save = async () => {
  if (settings.value === null) return false;
  const newSettings: ISettings = {
    ...settings.value,
    shell:
      settings.value?.shell
        .join(" ")
        .trim()
        .split(" ")
        .filter((s: string) => s !== "") ?? [],
    commands: {},
  };

  const keys = Object.keys(settings.value.commands) as Array<
    keyof SettingsCommand
  >;
  for (const key of keys) {
    const newValue = commandObject.value[key];
    if (!newValue) continue;

    if (Array.isArray(newValue)) {
      newSettings.commands[key] = newValue;
    } else if (key in commandObject.value) {
      newSettings.commands[key] = newValue
        .split("\n")
        .filter((cmd: string) => cmd !== "");
    }
  }
  newSettings.shell = shellValue.value
    .trim()
    .split(" ")
    .filter((s) => s !== "");

  if (newSettings.branding.theme !== getTheme()) {
    setTheme(newSettings.branding.theme);
  }

  try {
    await api.update(newSettings);
    $showSuccess(t("settings.settingsUpdated"));
  } catch (e: any) {
    $showError(e);
  }

  return true;
};

const parseBytes = (input: string) => {
  const regex = /^(\d+)(\.\d+)?(B|K|KB|M|MB|G|GB|T|TB)?$/i;
  const matches = input.match(regex);
  if (matches) {
    const size = parseFloat(matches[1].concat(matches[2] || ""));
    let unit: keyof SettingsUnit =
      matches[3].toUpperCase() as keyof SettingsUnit;
    if (!unit.endsWith("B")) {
      unit += "B";
    }
    const units: SettingsUnit = {
      KB: 1024,
      MB: 1024 ** 2,
      GB: 1024 ** 3,
      TB: 1024 ** 4,
    };
    return size * (units[unit as keyof SettingsUnit] || 1);
  } else {
    return 1024 ** 2;
  }
};

const formatBytes = (bytes: number) => {
  const units = ["B", "KB", "MB", "GB", "TB"];
  let size = bytes;
  let unitIndex = 0;
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex++;
  }
  return `${size}${units[unitIndex]}`;
};

onMounted(async () => {
  try {
    layoutStore.loading = true;
    const original: ISettings = await api.get();
    const newSettings: ISettings = { ...original, commands: {} };

    const keys = Object.keys(original.commands) as Array<keyof SettingsCommand>;
    for (const key of keys) {
      newSettings.commands[key] = original.commands[key];
      commandObject.value[key] = original.commands[key]!.join("\n");
    }

    originalSettings.value = original;
    settings.value = newSettings;
    shellValue.value = newSettings.shell.join(" ");
  } catch (err) {
    if (err instanceof Error) {
      error.value = err;
    }
  } finally {
    layoutStore.loading = false;
  }
});

onBeforeUnmount(() => {
  if (debounceTimeout.value) {
    clearTimeout(debounceTimeout.value);
  }
});
</script>
