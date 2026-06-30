<template>
  <div>
    <router-view></router-view>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from "vue";
import { useI18n } from "vue-i18n";
import { setHtmlLocale } from "./i18n";
import { getMediaPreference, setTheme } from "./utils/theme";
import { useAuthStore } from "./stores/auth";

const { locale } = useI18n();
const authStore = useAuthStore();

// Priority: localStorage (instant) > backend user.theme > OS preference
const cached = localStorage.getItem("fb-theme") as UserTheme;
const userTheme = ref<UserTheme>(
  cached || authStore.user?.theme || getMediaPreference()
);

onMounted(() => {
  setTheme(userTheme.value);
  setHtmlLocale(locale.value);
  // this might be null during HMR
  const loading = document.getElementById("loading");
  loading?.classList.add("done");

  setTimeout(function () {
    loading?.parentNode?.removeChild(loading);
  }, 200);
});

// handles ltr/rtl changes
watch(locale, (newValue) => {
  newValue && setHtmlLocale(newValue);
});
</script>
