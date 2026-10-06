// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const { store, push } = vi.hoisted(() => ({
  store: {
    jwt: "",
    user: null as unknown,
    logoutTimer: null as number | null,
    setUser: vi.fn(),
    setLogoutTimer: vi.fn(),
    clearUser: vi.fn(),
  },
  push: vi.fn(),
}));
vi.mock("@/stores/auth", () => ({ useAuthStore: () => store }));
vi.mock("@/router", () => ({ default: { push } }));
vi.mock("@/utils/constants", () => ({
  baseURL: "",
  authMethod: "json",
  noAuth: false,
  logoutPage: "/login",
}));

function token(seconds = 120) {
  return `${btoa("{}")}.${btoa(
    JSON.stringify({
      exp: Math.floor(Date.now() / 1000) + seconds,
      user: { id: 1 },
    })
  )}.signature`;
}

describe("session lifetime", () => {
  const fetchMock = vi.fn();
  beforeEach(() => {
    vi.resetModules();
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-01-01T00:00:00Z"));
    vi.clearAllMocks();
    vi.stubGlobal("fetch", fetchMock);
    fetchMock.mockReset();
    store.jwt = "";
    store.user = null;
    store.logoutTimer = null;
    store.setUser.mockImplementation((user) => {
      store.user = user;
    });
    store.setLogoutTimer.mockImplementation((timer) => {
      store.logoutTimer = timer;
    });
    store.clearUser.mockImplementation(() => {
      store.jwt = "";
      store.user = null;
    });
  });
  afterEach(() => {
    vi.clearAllTimers();
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  it("renews before expiration instead of logging out", async () => {
    const { parseToken } = await import("@/utils/auth");
    parseToken(token(2));
    fetchMock.mockImplementation(async () => new Response(token()));
    await vi.advanceTimersByTimeAsync(2100);
    expect(store.user).not.toBeNull();
    expect(fetchMock).toHaveBeenCalledWith("/api/renew", expect.anything());
    expect(fetchMock.mock.calls.every(([url]) => url !== "/api/logout")).toBe(
      true
    );
    expect(push).not.toHaveBeenCalled();
  });

  it("retries a transient renewal failure without clearing the session", async () => {
    const { parseToken } = await import("@/utils/auth");
    parseToken(token(2));
    fetchMock.mockRejectedValueOnce(new TypeError("offline"));
    fetchMock.mockImplementation(async () => new Response(token()));
    await vi.advanceTimersByTimeAsync(2100);
    expect(store.user).not.toBeNull();
    await vi.advanceTimersByTimeAsync(30_000);
    expect(
      fetchMock.mock.calls.filter(([url]) => url === "/api/renew")
    ).toHaveLength(2);
    expect(push).not.toHaveBeenCalled();
  });

  it("coalesces simultaneous renewal requests and uses the freshest cookie", async () => {
    const { renew } = await import("@/utils/auth");
    store.jwt = "stale-tab-token";
    fetchMock.mockImplementation(async () => new Response(token()));
    await Promise.all([renew(), renew(), renew()]);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/renew",
      expect.objectContaining({
        credentials: "include",
        headers: {},
      })
    );
  });

  it("does not restore a session if an in-flight renewal finishes after logout", async () => {
    const { renew, logout } = await import("@/utils/auth");
    let complete!: (res: Response) => void;
    fetchMock.mockImplementationOnce(
      () =>
        new Promise<Response>((resolve) => {
          complete = resolve;
        })
    );
    fetchMock.mockResolvedValue(new Response(""));
    const renewing = renew().catch(() => undefined);
    await logout();
    complete(new Response(token()));
    await renewing;
    expect(store.user).toBeNull();
    expect(store.jwt).toBe("");
  });

  it("clears only local state when automatic renewal is rejected", async () => {
    const { parseToken } = await import("@/utils/auth");
    parseToken(token(2));
    fetchMock.mockResolvedValue(new Response("Unauthorized", { status: 401 }));
    await vi.advanceTimersByTimeAsync(2100);
    expect(store.user).toBeNull();
    expect(fetchMock.mock.calls.every(([url]) => url !== "/api/logout")).toBe(
      true
    );
    expect(push).toHaveBeenCalled();
  });
});
