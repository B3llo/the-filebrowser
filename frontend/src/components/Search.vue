<template>
  <div
    ref="wrapperEl"
    class="fb-search-wrap"
    :class="{ 'fb-search-wrap--open': isMobileOpen }"
    @keydown.esc.stop="close"
    @focusout="onFocusout"
  >
    <FbIcon name="search" size="16px" class="fb-search-icon" />
    <input
      ref="inputEl"
      class="fb-search"
      type="text"
      v-model.trim="prompt"
      :placeholder="$t('search.search')"
      :aria-label="$t('search.search')"
      @focus="onFocus"
      @keyup.enter="submit"
    />
    <button
      v-if="ongoing"
      class="fb-search-end"
      @mousedown.prevent="stopSearch"
      :title="$t('buttons.stopSearch')"
      :aria-label="$t('buttons.stopSearch')"
    >
      <FbIcon name="stop-circle" size="15px" />
    </button>
    <button
      v-else-if="prompt"
      class="fb-search-end"
      @mousedown.prevent="clearSearch"
    >
      <FbIcon name="x" size="15px" />
    </button>

    <div
      v-show="isOpen"
      ref="dropdownEl"
      class="fb-search-dropdown"
      @mousedown.prevent
    >
      <div v-if="isEmpty">
        <p class="fb-search-hint">{{ text }}</p>
        <template v-if="prompt === ''">
          <div class="fb-search-types">
            <h3>{{ $t("search.types") }}</h3>
            <div class="fb-search-type-grid">
              <button
                v-for="(v, k) in boxes"
                :key="k"
                class="fb-search-type-btn"
                @click="init('type:' + k)"
              >
                <FbIcon :name="v.icon" size="16px" />
                <span>{{ $t("search." + v.label) }}</span>
              </button>
            </div>
          </div>
        </template>
      </div>
      <ul v-show="results.length > 0">
        <li v-for="(s, k) in filteredResults" :key="k">
          <router-link @click="close" :to="s.url">
            <FbIcon :name="s.dir ? 'folder' : 'file'" size="15px" />
            <span>./{{ s.path }}</span>
          </router-link>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useFileStore } from "@/stores/file";
import FbIcon from "@/components/FbIcon.vue";
import type { IconName } from "@/utils/icons";
import url from "@/utils/url";
import { search } from "@/api";
import {
  computed,
  inject,
  nextTick,
  onMounted,
  onUnmounted,
  ref,
  watch,
} from "vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";
import { StatusError } from "@/api/utils";

const boxes: Record<string, { label: string; icon: IconName }> = {
  image: { label: "images", icon: "image" },
  audio: { label: "music", icon: "audio" },
  video: { label: "video", icon: "video" },
  pdf: { label: "pdf", icon: "pdf" },
};

const fileStore = useFileStore();
let searchAbortController = new AbortController();

const prompt = ref<string>("");
const ongoing = ref<boolean>(false);
const results = ref<any[]>([]);
const resultsCount = ref<number>(50);
const isOpen = ref<boolean>(false);
const isMobileOpen = ref<boolean>(false);

const $showError = inject<IToastError>("$showError")!;

const inputEl = ref<HTMLInputElement | null>(null);
const dropdownEl = ref<HTMLElement | null>(null);
const wrapperEl = ref<HTMLElement | null>(null);

const { t } = useI18n();
const route = useRoute();

watch(prompt, () => {
  reset();
});

const isEmpty = computed(() => results.value.length === 0);

const text = computed(() => {
  if (ongoing.value) return "";
  return prompt.value === ""
    ? t("search.typeToSearch")
    : t("search.pressToSearch");
});

const filteredResults = computed(() =>
  results.value.slice(0, resultsCount.value)
);

onMounted(() => {
  if (!dropdownEl.value) return;
  dropdownEl.value.addEventListener("scroll", (event: Event) => {
    const el = event.target as HTMLElement;
    if (el.offsetHeight + el.scrollTop >= el.scrollHeight - 100) {
      resultsCount.value += 50;
    }
  });
});

onUnmounted(() => {
  abortLastSearch();
});

const onFocus = () => {
  isOpen.value = true;
};

const onFocusout = (e: FocusEvent) => {
  if (!wrapperEl.value?.contains(e.relatedTarget as Node)) {
    isOpen.value = false;
    isMobileOpen.value = false;
  }
};

const close = () => {
  isOpen.value = false;
  isMobileOpen.value = false;
  prompt.value = "";
  reset();
  inputEl.value?.blur();
};

const stopSearch = () => {
  abortLastSearch();
  ongoing.value = false;
};

const clearSearch = () => {
  prompt.value = "";
  reset();
  inputEl.value?.focus();
};

const init = (string: string) => {
  prompt.value = `${string} `;
  inputEl.value?.focus();
};

const reset = () => {
  abortLastSearch();
  ongoing.value = false;
  resultsCount.value = 50;
  results.value = [];
};

const abortLastSearch = () => {
  searchAbortController.abort();
};

const submit = async (event: Event) => {
  event.preventDefault();

  if (prompt.value === "") return;

  let path = route.path;
  if (!fileStore.isListing) {
    path = url.removeLastDir(path) + "/";
  }

  ongoing.value = true;

  try {
    abortLastSearch();
    searchAbortController = new AbortController();
    results.value = [];
    await search(path, prompt.value, searchAbortController.signal, (item) =>
      results.value.push(item)
    );
  } catch (error: any) {
    if (error instanceof StatusError && error.is_canceled) {
      return;
    }
    $showError(error);
  }

  ongoing.value = false;
};

const focus = () => {
  isMobileOpen.value = true;
  nextTick(() => inputEl.value?.focus());
};

defineExpose({ focus });
</script>
