// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("@/stores/auth", () => ({
  useAuthStore: () => ({ jwt: "test-jwt" }),
}));

vi.mock("@/utils/auth", () => ({
  renew: vi.fn(),
  logout: vi.fn(),
}));

function response(body: string, status = 200) {
  return new Response(body, {
    status,
    statusText: status === 401 ? "Unauthorized" : "OK",
  });
}

describe("api/utils fetchURL session recovery", () => {
  const fetchMock = vi.fn();
  let fetchURL: typeof import("@/api/utils").fetchURL;
  let StatusError: typeof import("@/api/utils").StatusError;
  let authUtils: typeof import("@/utils/auth");

  beforeEach(async () => {
    fetchMock.mockReset();
    vi.stubGlobal("fetch", fetchMock);
    vi.stubGlobal("window", {
      FileBrowser: { Name: "File Browser", BaseURL: "" },
      location: { origin: "http://localhost" },
    });
    vi.resetModules();

    const apiUtils = await import("@/api/utils");
    fetchURL = apiUtils.fetchURL;
    StatusError = apiUtils.StatusError;
    authUtils = await import("@/utils/auth");
    vi.mocked(authUtils.renew).mockReset();
    vi.mocked(authUtils.logout).mockReset();
  });

  it("renews and replays the request on 401 instead of logging out", async () => {
    fetchMock
      .mockResolvedValueOnce(response("401 Unauthorized", 401))
      .mockResolvedValueOnce(response("ok", 200));
    vi.mocked(authUtils.renew).mockResolvedValueOnce(undefined);

    const res = await fetchURL("/api/resources/", {});

    expect(res.status).toBe(200);
    expect(authUtils.renew).toHaveBeenCalledTimes(1);
    expect(authUtils.logout).not.toHaveBeenCalled();
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it("logs out only when the refresh itself is rejected", async () => {
    fetchMock.mockResolvedValueOnce(response("401 Unauthorized", 401));
    vi.mocked(authUtils.renew).mockRejectedValueOnce(
      new StatusError("401 Unauthorized", 401)
    );

    await expect(fetchURL("/api/resources/", {})).rejects.toMatchObject({
      status: 401,
    });
    expect(authUtils.logout).toHaveBeenCalledWith(undefined, false);
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("does not log out on a resource 401 after successful renewal", async () => {
    fetchMock.mockResolvedValue(response("Unauthorized", 401));
    vi.mocked(authUtils.renew).mockResolvedValue(undefined);
    await expect(fetchURL("/api/resources/", {})).rejects.toMatchObject({
      status: 401,
    });
    expect(authUtils.logout).not.toHaveBeenCalled();
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it("returns a successful response even if proactive renewal is offline", async () => {
    const res = response("ok");
    res.headers.set("X-Renew-Token", "true");
    fetchMock.mockResolvedValue(res);
    vi.mocked(authUtils.renew).mockRejectedValue(new TypeError("offline"));
    expect((await fetchURL("/api/resources/", {})).status).toBe(200);
    expect(authUtils.logout).not.toHaveBeenCalled();
  });

  it("preserves cancellation when the replayed request is aborted", async () => {
    fetchMock
      .mockResolvedValueOnce(response("Unauthorized", 401))
      .mockRejectedValueOnce(new DOMException("Aborted", "AbortError"));
    vi.mocked(authUtils.renew).mockResolvedValue(undefined);
    await expect(fetchURL("/api/resources/", {})).rejects.toMatchObject({
      status: 0,
      is_canceled: true,
    });
    expect(authUtils.logout).not.toHaveBeenCalled();
  });

  it("keeps the session on transient renew failures", async () => {
    fetchMock.mockResolvedValueOnce(response("401 Unauthorized", 401));
    vi.mocked(authUtils.renew).mockRejectedValueOnce(
      new StatusError("000 No connection", 0)
    );

    await expect(fetchURL("/api/resources/", {})).rejects.toMatchObject({
      status: 401,
    });
    expect(authUtils.logout).not.toHaveBeenCalled();
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });
});
