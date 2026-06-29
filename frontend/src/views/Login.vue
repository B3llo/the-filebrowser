<template>
  <div id="login" :class="{ recaptcha: recaptcha }">
    <form @submit="submit">
      <img :src="logoURL" alt="File Browser" />
      <h1>{{ name }}</h1>
      <p v-if="reason != null" class="logout-message">
        {{ t(`login.logout_reasons.${reason}`) }}
      </p>
      <div v-if="error !== ''" class="wrong">{{ error }}</div>

      <input
        autofocus
        class="input input--block"
        type="text"
        autocapitalize="off"
        v-model="username"
        :placeholder="t('login.username')"
      />
      <input
        class="input input--block"
        type="password"
        v-model="password"
        :placeholder="t('login.password')"
      />
      <input
        class="input input--block"
        v-if="createMode"
        type="password"
        v-model="passwordConfirm"
        :placeholder="t('login.passwordConfirm')"
      />

      <div v-if="recaptcha" id="recaptcha"></div>
      <input
        class="button button--block"
        type="submit"
        :disabled="loading"
        :value="
          loading
            ? t('login.submitting')
            : createMode
              ? t('login.signup')
              : t('login.submit')
        "
      />

      <p @click="toggleMode" v-if="signup">
        {{ createMode ? t("login.loginInstead") : t("login.createAnAccount") }}
      </p>
    </form>
  </div>
</template>

<script setup lang="ts">
import { StatusError } from "@/api/utils";
import * as auth from "@/utils/auth";
import { useAuthStore } from "@/stores/auth";
import {
  name,
  logoURL,
  recaptcha,
  recaptchaKey,
  signup,
} from "@/utils/constants";
import { inject, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute, useRouter } from "vue-router";

// Define refs
const createMode = ref<boolean>(false);
const error = ref<string>("");
const username = ref<string>("");
const password = ref<string>("");
const passwordConfirm = ref<string>("");
const loading = ref<boolean>(false);

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();
const { t } = useI18n({});
// Define functions
const toggleMode = () => (createMode.value = !createMode.value);

const $showError = inject<IToastError>("$showError")!;

const reason = route.query["logout-reason"] ?? null;

// Only allow same-origin, absolute-path redirects. A user-controlled
// `?redirect=` must never become an open redirect (e.g. //evil.com or
// https://evil.com), since we navigate with window.location below.
const safePath = (p: string) =>
  p.startsWith("/") && !p.startsWith("//") ? p : "/files/";

const submit = async (event: Event) => {
  event.preventDefault();
  event.stopPropagation();

  if (loading.value) return;

  const redirect = safePath((route.query.redirect as string) || "/files/");

  let captcha = "";
  if (recaptcha) {
    captcha = window.grecaptcha.getResponse();

    if (captcha === "") {
      error.value = t("login.wrongCredentials");
      return;
    }
  }

  if (createMode.value) {
    if (password.value !== passwordConfirm.value) {
      error.value = t("login.passwordsDontMatch");
      return;
    }
  }

  error.value = "";
  loading.value = true;

  try {
    if (createMode.value) {
      await auth.signup(username.value, password.value);
    }

    await auth.login(username.value, password.value, captcha);
  } catch (e: any) {
    if (e instanceof StatusError) {
      if (e.status === 409) {
        error.value = t("login.usernameTaken");
      } else if (e.status === 403) {
        error.value = t("login.wrongCredentials");
      } else if (e.status === 400) {
        const match = e.message.match(/minimum length is (\d+)/);
        if (match) {
          error.value = t("login.passwordTooShort", { min: match[1] });
        } else {
          error.value = e.message;
        }
      } else {
        $showError(e);
      }
    } else if (!authStore.isLoggedIn) {
      // Unknown failure and the session was not established — surface it
      // instead of silently leaving the user on a dead button.
      $showError(e);
    }
    // If the token was already saved despite a tail error, fall through to
    // the redirect below (authStore.isLoggedIn will be true).
  } finally {
    if (!authStore.isLoggedIn) {
      loading.value = false;
    }
  }

  // Navigate via a full load to the resolved URL. This mirrors the refresh
  // (F5) path that reliably establishes the session, sources and routing,
  // avoiding the in-SPA navigation race that left the button stuck (#29).
  if (authStore.isLoggedIn) {
    window.location.assign(router.resolve({ path: redirect }).href);
  }
};

// Run hooks
onMounted(() => {
  if (!recaptcha) return;

  window.grecaptcha.ready(function () {
    window.grecaptcha.render("recaptcha", {
      sitekey: recaptchaKey,
    });
  });
});
</script>
