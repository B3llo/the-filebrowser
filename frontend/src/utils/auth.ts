import { useAuthStore } from "@/stores/auth";
import router from "@/router";
import type { JwtPayload } from "jwt-decode";
import { jwtDecode } from "jwt-decode";
import { authMethod, baseURL, noAuth, logoutPage } from "./constants";
import { StatusError } from "@/api/utils";
import { setSafeTimeout } from "@/api/utils";

// Only same-origin relative paths are allowed as logout destinations.
// Rejects protocol-relative ("//evil"), schemes ("http:", "javascript:"),
// backslashes and anything that escapes the origin.
export function isSafeLogoutPage(page: unknown): page is string {
  if (typeof page !== "string" || page === "") return false;
  if (!page.startsWith("/") || page.startsWith("//")) return false;
  if (page.includes("\\") || page.includes(":") || /[\s<>"]/.test(page))
    return false;
  if (/^javascript:/i.test(page)) return false;
  try {
    const url = new URL(page, window.location.origin);
    if (url.origin !== window.location.origin) return false;
    if (url.protocol !== "http:" && url.protocol !== "https:")
      return false;
  } catch {
    return false;
  }
  return true;
}

function safeLogoutPage(): string {
  return isSafeLogoutPage(logoutPage) ? logoutPage : "/login";
}

export function parseToken(token: string) {
  // falsy or malformed jwt will throw InvalidTokenError
  const data = jwtDecode<JwtPayload & { user: IUser }>(token);

  // The JWT lives only in memory (pinia). The server also sets an HttpOnly
  // refresh cookie — never persist the token in localStorage or a readable
  // document.cookie (XSS exfiltration). Remove legacy copies if present.
  try {
    localStorage.removeItem("jwt");
  } catch {
    /* storage unavailable */
  }

  const authStore = useAuthStore();
  authStore.jwt = token;
  authStore.setUser(data.user);

  // proxy auth with custom logout subject to unknown external timeout
  if (logoutPage !== "/login" && authMethod === "proxy") {
    console.warn("idle timeout disabled with proxy auth and custom logout");
    return;
  }

  if (authStore.logoutTimer) {
    clearTimeout(authStore.logoutTimer);
  }

  const expiresAt = new Date(data.exp! * 1000);
  const timeout = expiresAt.getTime() - Date.now();
  authStore.setLogoutTimer(
    setSafeTimeout(() => {
      void logout("inactivity");
    }, timeout)
  );
}

export async function validateLogin() {
  // Session restore goes through the HttpOnly cookie — no localStorage.
  await renew();
}

export async function login(
  username: string,
  password: string,
  recaptcha: string
) {
  const data = { username, password, recaptcha };

  const res = await fetch(`${baseURL}/api/login`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
    credentials: "include",
  });

  const body = await res.text();

  if (res.status === 200) {
    parseToken(body);
  } else {
    throw new StatusError(
      body || `${res.status} ${res.statusText}`,
      res.status
    );
  }
}

export async function renew(jwt?: string) {
  const authStore = useAuthStore();
  const token = jwt ?? authStore.jwt;

  const res = await fetch(`${baseURL}/api/renew`, {
    method: "POST",
    headers: {
      ...(token ? { "X-Auth": token } : {}),
    },
    credentials: "include",
  });

  const body = await res.text();

  if (res.status === 200) {
    parseToken(body);
  } else {
    throw new StatusError(
      body || `${res.status} ${res.statusText}`,
      res.status
    );
  }
}

export async function signup(username: string, password: string) {
  const data = { username, password };

  const res = await fetch(`${baseURL}/api/signup`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  });

  if (res.status !== 200) {
    const body = await res.text();
    throw new StatusError(
      body || `${res.status} ${res.statusText}`,
      res.status
    );
  }
}

export async function logout(reason?: string) {
  const authStore = useAuthStore();
  const timer = authStore.logoutTimer;

  // Best effort: invalidate the server session (LastUpdate bump + cookie
  // clear) before dropping local state.
  try {
    await fetch(`${baseURL}/api/logout`, {
      method: "POST",
      headers: {
        ...(authStore.jwt ? { "X-Auth": authStore.jwt } : {}),
      },
      credentials: "include",
    });
  } catch {
    /* offline — still clear local state */
  }

  if (timer) {
    clearTimeout(timer);
  }

  authStore.clearUser();

  // Drop any legacy token copies from older versions.
  try {
    localStorage.removeItem("jwt");
  } catch {
    /* storage unavailable */
  }

  if (noAuth) {
    window.location.reload();
  } else {
    const page = safeLogoutPage();
    if (page !== "/login") {
      document.location.href = page;
    } else {
      if (typeof reason === "string" && reason.trim() !== "") {
        await router.push({
          path: "/login",
          query: { "logout-reason": reason },
        });
      } else {
        await router.push({
          path: "/login",
        });
      }
    }
  }
}
