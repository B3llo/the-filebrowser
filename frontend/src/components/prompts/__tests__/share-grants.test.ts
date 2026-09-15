// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";
import SharePrompt from "@/components/prompts/Share.vue";
import { grants as grantApi, users as usersApi } from "@/api";

vi.mock("@/api", () => ({
  share: { get: vi.fn(), create: vi.fn(), remove: vi.fn() },
  grants: { list: vi.fn(), create: vi.fn(), remove: vi.fn() },
  users: { searchUsers: vi.fn() },
  files: {},
}));

vi.mock("@/api/utils", () => ({
  removePrefix: (value: string) => value.replace(/^\/files\/[^/]+/, "") || "/",
}));

vi.mock("@/utils/clipboard", () => ({
  copy: vi.fn(),
}));

vi.mock("@/utils/nativeShare", () => ({
  canShareFiles: () => false,
  shareViaOS: vi.fn(),
}));

function makeContext(overrides = {}) {
  return {
    // Bind sibling methods like Vue does, so methods calling
    // this.normalizedGrantPath()/this.loadGrants() work.
    ...(SharePrompt as any).methods,
    url: "/files/0/docs",
    grants: [],
    searchQuery: "",
    searchResults: [],
    selectedUser: null,
    grantRole: "viewer",
    $t: (key: string) => key,
    $showError: vi.fn(),
    $showSuccess: vi.fn(),
    ...overrides,
  };
}

describe("share prompt grants tab", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("normalizedGrantPath strips the files prefix", () => {
    const ctx = makeContext();
    expect((SharePrompt as any).methods.normalizedGrantPath.call(ctx)).toBe(
      "/docs"
    );
  });

  it("loadGrants keeps only grants for the current path", async () => {
    vi.mocked(grantApi.list).mockResolvedValue([
      { id: 1, path: "/docs", granteeID: 2, role: "viewer" },
      { id: 2, path: "/other", granteeID: 2, role: "viewer" },
    ] as any);
    const ctx = makeContext();
    await (SharePrompt as any).methods.loadGrants.call(ctx);
    expect(ctx.grants).toEqual([
      { id: 1, path: "/docs", granteeID: 2, role: "viewer" },
    ]);
  });

  it("runUserSearch clears results for short queries", async () => {
    const ctx = makeContext({ searchQuery: "a", searchResults: [{ id: 9 }] });
    await (SharePrompt as any).methods.runUserSearch.call(ctx);
    expect(usersApi.searchUsers).not.toHaveBeenCalled();
    expect(ctx.searchResults).toEqual([]);
  });

  it("runUserSearch queries from 2 chars", async () => {
    vi.mocked(usersApi.searchUsers).mockResolvedValue([
      { id: 2, username: "bob" },
    ] as any);
    const ctx = makeContext({ searchQuery: "bo" });
    await (SharePrompt as any).methods.runUserSearch.call(ctx);
    expect(usersApi.searchUsers).toHaveBeenCalledWith("bo");
    expect(ctx.searchResults).toEqual([{ id: 2, username: "bob" }]);
  });

  it("submitGrant creates, appends and resets", async () => {
    vi.mocked(grantApi.create).mockResolvedValue({
      id: 5,
      path: "/docs",
    } as any);
    const ctx = makeContext({
      selectedUser: { id: 2, username: "bob" },
      grantRole: "editor",
    });
    await (SharePrompt as any).methods.submitGrant.call(ctx);
    expect(grantApi.create).toHaveBeenCalledWith({
      path: "/docs",
      grantee: "2",
      role: "editor",
    });
    expect(ctx.grants).toEqual([{ id: 5, path: "/docs" }]);
    expect(ctx.selectedUser).toBeNull();
    expect(ctx.$showSuccess).toHaveBeenCalledWith("grants.grantCreated");
  });

  it("submitGrant maps 409 to alreadyShared", async () => {
    vi.mocked(grantApi.create).mockRejectedValue({ status: 409 });
    const ctx = makeContext({ selectedUser: { id: 2, username: "bob" } });
    await (SharePrompt as any).methods.submitGrant.call(ctx);
    expect(ctx.$showError).toHaveBeenCalledWith("grants.alreadyShared");
    expect(ctx.grants).toEqual([]);
  });

  it("submitGrant without a user does nothing", async () => {
    const ctx = makeContext();
    await (SharePrompt as any).methods.submitGrant.call(ctx);
    expect(grantApi.create).not.toHaveBeenCalled();
  });

  it("deleteGrant revokes and filters", async () => {
    vi.mocked(grantApi.remove).mockResolvedValue(undefined as any);
    const ctx = makeContext({ grants: [{ id: 1 }, { id: 2 }] });
    await (SharePrompt as any).methods.deleteGrant.call(
      ctx,
      { preventDefault: vi.fn() },
      { id: 1 }
    );
    expect(grantApi.remove).toHaveBeenCalledWith(1);
    expect(ctx.grants).toEqual([{ id: 2 }]);
    expect(ctx.$showSuccess).toHaveBeenCalledWith("grants.grantRevoked");
  });
});
