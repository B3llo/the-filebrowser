import { useAuthStore } from "@/stores/auth";
import { renew, logout } from "@/utils/auth";
import { baseURL } from "@/utils/constants";
import { encodePath } from "@/utils/url";

// The active source id for the current view. It mirrors the "source" cookie
// consumed by the backend to rebind the user filesystem. The source store
// keeps this in sync on every navigation; api/files.ts reads it to rebuild
// resource URLs that include the source segment.
let activeSourceId = "0";

export function setActiveSourceId(id: string) {
  activeSourceId = id || "0";
}

export function getActiveSourceId(): string {
  return activeSourceId;
}

export class StatusError extends Error {
  constructor(
    message: any,
    public status?: number,
    public is_canceled?: boolean
  ) {
    super(message);
    this.name = "StatusError";
  }
}

export async function fetchURL(
  url: string,
  opts: ApiOpts,
  auth = true
): Promise<Response> {
  const authStore = useAuthStore();

  opts = opts || {};
  opts.headers = opts.headers || {};

  const { headers, ...rest } = opts;
  let res;
  try {
    res = await fetch(`${baseURL}${url}`, {
      headers: {
        "X-Auth": authStore.jwt,
        ...headers,
      },
      ...rest,
    });
  } catch (e) {
    // Check if the error is an intentional cancellation
    if (e instanceof Error && e.name === "AbortError") {
      throw new StatusError("000 No connection", 0, true);
    }
    throw new StatusError("000 No connection", 0);
  }

  if (auth && res.headers.get("X-Renew-Token") === "true") {
    await renew(authStore.jwt);
  }

  if (res.status < 200 || res.status > 299) {
    const body = await res.text();
    const error = new StatusError(
      body || `${res.status} ${res.statusText}`,
      res.status
    );

    if (auth && res.status == 401) {
      logout();
    }

    throw error;
  }

  return res;
}

export async function fetchJSON<T>(url: string, opts?: any): Promise<T> {
  const res = await fetchURL(url, opts);

  if (res.status === 200) {
    return res.json() as Promise<T>;
  }

  throw new StatusError(`${res.status} ${res.statusText}`, res.status);
}

export function removePrefix(url: string): string {
  const parts = url.split("/");
  // parts[0] === "" (leading slash); parts[1] === area ("files" | "share" ...).
  // Files URLs carry the active source as their first segment
  // (/files/<sourceId>/<path...>), so drop three segments there; everything else
  // drops two.
  const drop = parts[1] === "files" ? 3 : 2;
  url = parts.slice(drop).join("/");

  if (url === "") url = "/";
  if (url[0] !== "/") url = "/" + url;
  return url;
}

export function createURL(endpoint: string, searchParams = {}): string {
  let prefix = baseURL;
  if (!prefix.endsWith("/")) {
    prefix = prefix + "/";
  }
  const url = new URL(prefix + encodePath(endpoint), origin);
  url.search = new URLSearchParams(searchParams).toString();

  return url.toString();
}

export function setSafeTimeout(callback: () => void, delay: number): number {
  const MAX_DELAY = 86_400_000;
  let remaining = delay;

  function scheduleNext(): number {
    if (remaining <= MAX_DELAY) {
      return window.setTimeout(callback, remaining);
    } else {
      return window.setTimeout(() => {
        remaining -= MAX_DELAY;
        scheduleNext();
      }, MAX_DELAY);
    }
  }

  return scheduleNext();
}
