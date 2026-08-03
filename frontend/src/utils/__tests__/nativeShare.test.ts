import { describe, it, expect, beforeEach, vi } from "vitest";
import { fetchURL } from "@/api/utils";

vi.mock("@/api/utils", () => ({
  fetchURL: vi.fn(),
  removePrefix: (url: string) => {
    const parts = url.split("/");
    // Files URLs carry the active source as their first segment
    // (/files/<sourceId>/<path...>), so drop three segments there.
    const drop = parts[1] === "files" ? 3 : 2;
    let clean = parts.slice(drop).join("/");
    if (clean === "") clean = "/";
    if (clean[0] !== "/") clean = "/" + clean;
    return clean;
  },
}));

const mockFetchURL = vi.mocked(fetchURL);

class MemoryStorage implements Storage {
  private store = new Map<string, string>();
  get length() {
    return this.store.size;
  }
  clear() {
    this.store.clear();
  }
  getItem(key: string) {
    return this.store.get(key) ?? null;
  }
  key(index: number) {
    return Array.from(this.store.keys())[index] ?? null;
  }
  removeItem(key: string) {
    this.store.delete(key);
  }
  setItem(key: string, value: string) {
    this.store.set(key, value);
  }
}

function mockRawResponse(disposition: string | null) {
  mockFetchURL.mockResolvedValue({
    blob: async () => new Blob(["data"]),
    headers: {
      get: (name: string) =>
        name === "Content-Disposition" ? disposition : null,
    },
  } as unknown as Response);
}

function stubNavigator(
  share: () => Promise<void> | undefined,
  canShare: () => boolean = () => true
) {
  vi.stubGlobal("navigator", { language: "en-US", share, canShare });
}

beforeEach(() => {
  vi.stubGlobal("localStorage", new MemoryStorage());
  vi.stubGlobal(
    "window",
    Object.assign({
      FileBrowser: {
        Name: "File Browser",
        BaseURL: "",
        StaticURL: "/static",
        TusSettings: {},
      },
      location: { origin: "http://localhost" },
    })
  );
  mockFetchURL.mockReset();
});

describe("canShareFiles", () => {
  it("returns false when navigator.share is missing", async () => {
    stubNavigator(undefined);
    vi.stubGlobal("navigator", {
      language: "en-US",
      share: undefined,
      canShare: undefined,
    });
    const { canShareFiles } = await import("@/utils/nativeShare");
    expect(canShareFiles()).toBe(false);
  });

  it("returns false when canShare rejects files", async () => {
    stubNavigator(vi.fn(), () => false);
    const { canShareFiles } = await import("@/utils/nativeShare");
    expect(canShareFiles()).toBe(false);
  });

  it("returns true when files can be shared", async () => {
    stubNavigator(vi.fn(), () => true);
    const { canShareFiles } = await import("@/utils/nativeShare");
    expect(canShareFiles()).toBe(true);
  });
});

describe("shareViaOS", () => {
  it("returns unsupported when the browser cannot share files", async () => {
    stubNavigator(vi.fn(), () => false);
    const { shareViaOS } = await import("@/utils/nativeShare");
    const outcome = await shareViaOS([
      { url: "/files/0/docs/report.pdf", name: "report.pdf" },
    ]);
    expect(outcome).toBe("unsupported");
    expect(mockFetchURL).not.toHaveBeenCalled();
  });

  it("shares a single file with its mime type", async () => {
    const share = vi.fn().mockResolvedValue(undefined);
    stubNavigator(share, () => true);
    mockRawResponse(null);

    const { shareViaOS } = await import("@/utils/nativeShare");
    const outcome = await shareViaOS([
      { url: "/files/0/docs/report.pdf", name: "report.pdf" },
    ]);

    expect(outcome).toBe("shared");
    expect(mockFetchURL).toHaveBeenCalledWith("/api/raw/docs/report.pdf?");
    expect(share).toHaveBeenCalledTimes(1);
    const files = share.mock.calls[0][0].files;
    expect(files).toHaveLength(1);
    expect(files[0].name).toBe("report.pdf");
    expect(files[0].type).toBe("application/pdf");
  });

  it("zips a folder and uses the server-provided zip filename", async () => {
    const share = vi.fn().mockResolvedValue(undefined);
    stubNavigator(share, () => true);
    mockRawResponse("attachment; filename*=utf-8''my%20folder.zip");

    const { shareViaOS } = await import("@/utils/nativeShare");
    const outcome = await shareViaOS([
      { url: "/files/0/docs/photos/", name: "photos/", isDir: true },
    ]);

    expect(outcome).toBe("shared");
    expect(mockFetchURL).toHaveBeenCalledWith(
      "/api/raw/docs/photos/?algo=zip&"
    );
    const files = share.mock.calls[0][0].files;
    expect(files[0].name).toBe("my folder.zip");
    expect(files[0].type).toBe("application/zip");
  });

  it("zips multiple selected items", async () => {
    const share = vi.fn().mockResolvedValue(undefined);
    stubNavigator(share, () => true);
    mockRawResponse(null);

    const { shareViaOS } = await import("@/utils/nativeShare");
    const outcome = await shareViaOS([
      { url: "/files/0/a.txt", name: "a.txt" },
      { url: "/files/0/b.txt", name: "b.txt" },
    ]);

    expect(outcome).toBe("shared");
    expect(mockFetchURL).toHaveBeenCalledWith(
      "/api/raw/?files=%2Fa.txt%2C%2Fb.txt&algo=zip&"
    );
    const files = share.mock.calls[0][0].files;
    expect(files[0].type).toBe("application/zip");
    expect(files[0].name).toBe("files.zip");
  });

  it("falls back to octet-stream for unknown extensions", async () => {
    const share = vi.fn().mockResolvedValue(undefined);
    stubNavigator(share, () => true);
    mockRawResponse(null);

    const { shareViaOS } = await import("@/utils/nativeShare");
    await shareViaOS([{ url: "/files/0/data.xyz", name: "data.xyz" }]);

    expect(share.mock.calls[0][0].files[0].type).toBe(
      "application/octet-stream"
    );
  });

  it("returns canceled when the user dismisses the share sheet", async () => {
    stubNavigator(
      vi.fn().mockRejectedValue(new DOMException("aborted", "AbortError")),
      () => true
    );
    mockRawResponse(null);

    const { shareViaOS } = await import("@/utils/nativeShare");
    const outcome = await shareViaOS([
      { url: "/files/0/a.txt", name: "a.txt" },
    ]);

    expect(outcome).toBe("canceled");
  });

  it("returns unsupported when the user activation expired", async () => {
    stubNavigator(
      vi.fn().mockRejectedValue(new DOMException("expired", "NotAllowedError")),
      () => true
    );
    mockRawResponse(null);

    const { shareViaOS } = await import("@/utils/nativeShare");
    const outcome = await shareViaOS([
      { url: "/files/0/a.txt", name: "a.txt" },
    ]);

    expect(outcome).toBe("unsupported");
  });

  it("rethrows unexpected share errors", async () => {
    const error = new Error("boom");
    stubNavigator(vi.fn().mockRejectedValue(error), () => true);
    mockRawResponse(null);

    const { shareViaOS } = await import("@/utils/nativeShare");
    await expect(
      shareViaOS([{ url: "/files/0/a.txt", name: "a.txt" }])
    ).rejects.toThrow("boom");
  });

  it("lets fetch errors propagate", async () => {
    stubNavigator(vi.fn(), () => true);
    mockFetchURL.mockRejectedValue(new Error("network down"));

    const { shareViaOS } = await import("@/utils/nativeShare");
    await expect(
      shareViaOS([{ url: "/files/0/a.txt", name: "a.txt" }])
    ).rejects.toThrow("network down");
  });
});
