import type { RouteLocation } from "vue-router";
import { createRouter, createWebHistory } from "vue-router";
import Login from "@/views/Login.vue";
import Layout from "@/views/Layout.vue";
import Files from "@/views/Files.vue";
import Share from "@/views/Share.vue";
import Users from "@/views/settings/Users.vue";
import User from "@/views/settings/User.vue";
import Sources from "@/views/settings/Sources.vue";
import Source from "@/views/settings/Source.vue";
import Settings from "@/views/Settings.vue";
import GlobalSettings from "@/views/settings/Global.vue";
import ProfileSettings from "@/views/settings/Profile.vue";
import Shares from "@/views/settings/Shares.vue";
import Recents from "@/views/Recents.vue";
import Starred from "@/views/Starred.vue";
import Trash from "@/views/Trash.vue";
import Errors from "@/views/Errors.vue";
import { useAuthStore } from "@/stores/auth";
import { useSourceStore } from "@/stores/source";
import { baseURL, name } from "@/utils/constants";
import i18n from "@/i18n";
import { recaptcha, loginPage } from "@/utils/constants";
import { login, validateLogin } from "@/utils/auth";

const titles = {
  Login: "sidebar.login",
  Share: "buttons.share",
  Files: "files.files",
  Settings: "sidebar.settings",
  ProfileSettings: "settings.profileSettings",
  Shares: "settings.shareManagement",
  GlobalSettings: "settings.globalSettings",
  Users: "settings.users",
  User: "settings.user",
  Sources: "settings.sources",
  Source: "settings.source",
  Recents: "sidebar.recent",
  Starred: "sidebar.starred",
  Trash: "sidebar.trash",
  Forbidden: "errors.forbidden",
  NotFound: "errors.notFound",
  InternalServerError: "errors.internal",
};

const routes = [
  {
    path: "/login",
    name: "Login",
    component: Login,
  },
  {
    path: "/share",
    component: Layout,
    children: [
      {
        path: ":path*",
        name: "Share",
        component: Share,
      },
    ],
  },
  {
    path: "/files",
    component: Layout,
    meta: {
      requiresAuth: true,
    },
    children: [
      {
        path: ":sourceId/:path*",
        name: "Files",
        component: Files,
      },
    ],
  },
  {
    path: "/recent",
    component: Layout,
    meta: { requiresAuth: true },
    children: [{ path: "", name: "Recents", component: Recents }],
  },
  {
    path: "/starred",
    component: Layout,
    meta: { requiresAuth: true },
    children: [{ path: "", name: "Starred", component: Starred }],
  },
  {
    path: "/trash",
    component: Layout,
    meta: { requiresAuth: true },
    children: [{ path: "", name: "Trash", component: Trash }],
  },
  {
    path: "/settings",
    component: Layout,
    meta: {
      requiresAuth: true,
    },
    children: [
      {
        path: "",
        name: "Settings",
        component: Settings,
        redirect: {
          path: "/settings/profile",
        },
        children: [
          {
            path: "profile",
            name: "ProfileSettings",
            component: ProfileSettings,
          },
          {
            path: "shares",
            name: "Shares",
            component: Shares,
          },
          {
            path: "global",
            name: "GlobalSettings",
            component: GlobalSettings,
            meta: {
              requiresAdmin: true,
            },
          },
          {
            path: "users",
            name: "Users",
            component: Users,
            meta: {
              requiresAdmin: true,
            },
          },
          {
            path: "users/:id",
            name: "User",
            component: User,
            meta: {
              requiresAdmin: true,
            },
          },
          {
            path: "sources",
            name: "Sources",
            component: Sources,
            meta: {
              requiresAdmin: true,
            },
          },
          {
            path: "sources/:id",
            name: "Source",
            component: Source,
            meta: {
              requiresAdmin: true,
            },
          },
        ],
      },
    ],
  },
  {
    path: "/403",
    name: "Forbidden",
    component: Errors,
    props: {
      errorCode: 403,
      showHeader: true,
    },
  },
  {
    path: "/404",
    name: "NotFound",
    component: Errors,
    props: {
      errorCode: 404,
      showHeader: true,
    },
  },
  {
    path: "/500",
    name: "InternalServerError",
    component: Errors,
    props: {
      errorCode: 500,
      showHeader: true,
    },
  },
  {
    path: "/:catchAll(.*)*",
    redirect: (to: RouteLocation) => {
      const catchAll = to.params.catchAll;
      if (!catchAll) return "/files/";
      return `/files/${Array.isArray(catchAll) ? catchAll.join("/") : catchAll}`;
    },
  },
];

async function initAuth() {
  if (loginPage) {
    await validateLogin();
  } else {
    await login("", "", "");
  }

  if (recaptcha) {
    await new Promise<void>((resolve) => {
      const check = () => {
        if (typeof window.grecaptcha === "undefined") {
          setTimeout(check, 100);
        } else {
          resolve();
        }
      };

      check();
    });
  }
}

const router = createRouter({
  history: createWebHistory(baseURL),
  routes,
});

router.beforeResolve(async (to, from, next) => {
  const titleKey = to.name ? titles[to.name as keyof typeof titles] : undefined;
  document.title = (titleKey ? i18n.global.t(titleKey) : name) + " - " + name;

  const authStore = useAuthStore();
  const sourceStore = useSourceStore();

  // this will only be null on first route
  if (from.name == null) {
    try {
      await initAuth();
    } catch (error) {
      console.error(error);
    }
  }

  // Load the user's accessible sources once authenticated.
  if (authStore.isLoggedIn && !sourceStore.loaded) {
    try {
      await sourceStore.load();
    } catch (error) {
      console.error(error);
    }
  }

  // Normalize /files URLs so every deep link carries a valid source segment:
  //   /files           -> /files/<defaultSource>/
  //   /files/<seg>/... -> if <seg> is a known source, activate it; otherwise it
  //                       is a legacy path and is rebased under the default source.
  if (authStore.isLoggedIn && sourceStore.loaded) {
    if (to.path === "/files" || to.path === "/files/") {
      next({ path: `${sourceStore.filesBase}/` });
      return;
    }
    if (to.path.startsWith("/files/")) {
      const rest = to.path.slice("/files/".length);
      const slash = rest.indexOf("/");
      const seg = slash === -1 ? rest : rest.slice(0, slash);
      if (!sourceStore.isKnown(seg)) {
        next({ path: `/files/${sourceStore.defaultId}/${rest}` });
        return;
      }
      sourceStore.setActive(seg);
    }
  }

  if (to.path.endsWith("/login") && authStore.isLoggedIn) {
    next({ path: `${sourceStore.filesBase}/` });
    return;
  }

  if (to.matched.some((record) => record.meta.requiresAuth)) {
    if (!authStore.isLoggedIn) {
      next({
        path: "/login",
        query: { redirect: to.fullPath },
      });

      return;
    }

    if (to.matched.some((record) => record.meta.requiresAdmin)) {
      if (authStore.user === null || !authStore.user.perm.admin) {
        next({ path: "/403" });
        return;
      }
    }
  }

  next();
});

export { router, router as default };
