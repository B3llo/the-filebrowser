// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("@/stores/auth", () => ({
  useAuthStore: () => ({ jwt: "test-jwt" }),
}));

vi.mock("@/utils/auth", () => ({
  renew: vi.fn(),
  logout: vi.fn(),
}));

function jsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

describe("api/grants", () => {
  const fetchMock = vi.fn();
  let api: typeof import("@/api/grants");

  beforeEach(async () => {
    fetchMock.mockReset();
    vi.stubGlobal("fetch", fetchMock);
    vi.stubGlobal("window", {
      FileBrowser: { Name: "File Browser", BaseURL: "" },
      location: { origin: "http://localhost" },
    });
    vi.resetModules();
    api = await import("@/api/grants");
  });

  it("list hits /api/grants", async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse([{ id: 1 }]));
    const out = await api.list();
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining("/api/grants"),
      expect.objectContaining({
        headers: expect.objectContaining({ "X-Auth": "test-jwt" }),
      })
    );
    expect(out).toEqual([{ id: 1 }]);
  });

  it("sharedWithMe hits /api/grants/shared-with-me", async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse([]));
    await api.sharedWithMe();
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining("/api/grants/shared-with-me"),
      expect.anything()
    );
  });

  it("create posts path/grantee/role and strips /files prefix", async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse({ id: 7 }));
    await api.create({ path: "/files/0/docs", grantee: "bob", role: "viewer" });
    const [, opts] = fetchMock.mock.calls[0];
    expect(opts.method).toBe("POST");
    const body = JSON.parse(opts.body);
    expect(body).toMatchObject({
      path: "/docs",
      grantee: "bob",
      role: "viewer",
    });
    expect(body.owner).toBeUndefined();
  });

  it("create includes owner/expires when given", async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse({ id: 8 }));
    await api.create({
      path: "/docs",
      grantee: "bob",
      owner: "alice",
      role: "editor",
      expires: "2",
      unit: "days",
    });
    const body = JSON.parse(fetchMock.mock.calls[0][1].body);
    expect(body).toMatchObject({ owner: "alice", expires: "2", unit: "days" });
  });

  it("remove sends DELETE", async () => {
    fetchMock.mockResolvedValueOnce(new Response("", { status: 200 }));
    await api.remove(3);
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining("/api/grants/3"),
      expect.objectContaining({ method: "DELETE" })
    );
  });

  it("patch sends PATCH with role", async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse({ id: 3, role: "editor" }));
    const out = await api.patch(3, { role: "editor" });
    const [, opts] = fetchMock.mock.calls[0];
    expect(opts.method).toBe("PATCH");
    expect(JSON.parse(opts.body)).toEqual({ role: "editor" });
    expect(out.role).toBe("editor");
  });

  it("surfaces 409 as StatusError with status", async () => {
    fetchMock.mockResolvedValueOnce(
      new Response("409 Conflict", { status: 409, statusText: "Conflict" })
    );
    await expect(
      api.create({ path: "/docs", grantee: "bob", role: "viewer" })
    ).rejects.toMatchObject({ status: 409 });
  });
});
