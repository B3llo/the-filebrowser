import { describe, it, expect } from "vitest";
import {
  baseName,
  buildFlatTrashResource,
  buildTrashInfo,
  cleanTrashName,
  collapseChainLabel,
  isLegacyTrashName,
  isMirroredTrashPath,
  isTrashPath,
  mergeTrashResources,
  mirroredTrashSourcePath,
  originalLocation,
  originalPathFromTrash,
  parentDir,
  parseTrashInfo,
  shouldShowInTrash,
  sortByDepthDesc,
  splitName,
  topLevelNames,
  trashDestinationUrl,
  trashInfoUrl,
  trashMirrorRelative,
  trashSourcePathFromUrl,
  withVersionSuffix,
} from "@/utils/trash";

describe("cleanTrashName / isLegacyTrashName", () => {
  it("strips legacy epoch prefixes", () => {
    expect(cleanTrashName("1757284800000_foto.jpg")).toBe("foto.jpg");
    expect(isLegacyTrashName("1757284800000_foto.jpg")).toBe(true);
  });

  it("keeps mirrored names verbatim, including digit-underscore names", () => {
    expect(cleanTrashName("123_foo.txt")).toBe("foo.txt"); // legacy ambiguity preserved for old items
    expect(cleanTrashName("foto.jpg")).toBe("foto.jpg");
    expect(isLegacyTrashName("foto.jpg")).toBe(false);
  });
});

describe("mirror paths", () => {
  it("mirrors source paths under /.Trash/files", () => {
    expect(mirroredTrashSourcePath("/Docs/a.jpg")).toBe(
      "/.Trash/files/Docs/a.jpg"
    );
    expect(mirroredTrashSourcePath("/a.jpg")).toBe("/.Trash/files/a.jpg");
  });

  it("recovers the original path from a trash path", () => {
    expect(originalPathFromTrash("/.Trash/files/Docs/a.jpg")).toBe(
      "/Docs/a.jpg"
    );
    expect(originalPathFromTrash("/.Trash/123_a.jpg")).toBeNull();
    expect(originalPathFromTrash("/Docs/a.jpg")).toBeNull();
  });

  it("classifies trash paths", () => {
    expect(isTrashPath("/.Trash")).toBe(true);
    expect(isTrashPath("/.Trash/files/Docs")).toBe(true);
    expect(isTrashPath("/Docs")).toBe(false);
    expect(isMirroredTrashPath("/.Trash/files/Docs")).toBe(true);
    expect(isMirroredTrashPath("/.Trash/123_a.jpg")).toBe(false);
  });

  it("computes mirror relative paths", () => {
    expect(trashMirrorRelative("/.Trash/files")).toBe("");
    expect(trashMirrorRelative("/.Trash/files/Docs/a.jpg")).toBe("Docs/a.jpg");
    expect(trashMirrorRelative("/.Trash/info/x")).toBeNull();
  });

  it("builds destination and sidecar urls", () => {
    expect(trashDestinationUrl("/files/2", "/Docs/a.jpg")).toBe(
      "/files/2/.Trash/files/Docs/a.jpg"
    );
    expect(trashInfoUrl("/files/2", "/.Trash/files/Docs/a.jpg")).toBe(
      "/files/2/.Trash/info/Docs/a.jpg.trashinfo"
    );
    expect(trashInfoUrl("/files/2", "/.Trash/files")).toBeNull();
    expect(trashInfoUrl("/files/2", "/Docs/a.jpg")).toBeNull();
  });

  it("recovers source paths from urls", () => {
    expect(
      trashSourcePathFromUrl("/files/2/.Trash/files/Docs/a.jpg", "/files/2")
    ).toBe("/.Trash/files/Docs/a.jpg");
    expect(
      trashSourcePathFromUrl("/files/9/.Trash/files/a.jpg", "/files/2")
    ).toBeNull();
  });
});

describe("parentDir / baseName / splitName / withVersionSuffix", () => {
  it("parentDir", () => {
    expect(parentDir("/Docs/Viagem/f.jpg")).toBe("/Docs/Viagem");
    expect(parentDir("/f.jpg")).toBe("/");
    expect(parentDir("/")).toBe("/");
    expect(parentDir("/Docs/Viagem/")).toBe("/Docs");
  });

  it("baseName", () => {
    expect(baseName("/Docs/a.jpg")).toBe("a.jpg");
    expect(baseName("/Docs/Viagem/")).toBe("Viagem");
  });

  it("splitName keeps dotfiles intact", () => {
    expect(splitName("archive.tar.gz")).toEqual({
      base: "archive.tar",
      ext: ".gz",
    });
    expect(splitName(".gitignore")).toEqual({ base: ".gitignore", ext: "" });
    expect(splitName("README")).toEqual({ base: "README", ext: "" });
  });

  it("withVersionSuffix", () => {
    expect(withVersionSuffix("f.jpg", 0)).toBe("f.jpg");
    expect(withVersionSuffix("f.jpg", 1)).toBe("f(1).jpg");
    expect(withVersionSuffix("Viagem", 2)).toBe("Viagem(2)");
  });
});

describe("trashinfo sidecars", () => {
  const info = buildTrashInfo({
    originalPath: "/Docs/a.jpg",
    name: "a.jpg",
    isDir: false,
    deletedAt: "2026-09-08T00:00:00.000Z",
    sourceId: "2",
    batchId: 42,
  });

  it("round-trips through JSON", () => {
    expect(parseTrashInfo(JSON.stringify(info))).toEqual(info);
  });

  it("rejects malformed payloads", () => {
    expect(parseTrashInfo("not json")).toBeNull();
    expect(parseTrashInfo(JSON.stringify({}))).toBeNull();
    expect(
      parseTrashInfo(JSON.stringify({ ...info, originalPath: "relative" }))
    ).toBeNull();
    expect(
      parseTrashInfo(JSON.stringify({ ...info, batchId: "42" }))
    ).toBeNull();
  });
});

describe("display helpers", () => {
  it("originalLocation is the parent dir", () => {
    expect(originalLocation("/Docs/Viagem/f.jpg")).toBe("/Docs/Viagem");
    expect(originalLocation("/f.jpg")).toBe("/");
  });

  it("collapseChainLabel joins with ›", () => {
    expect(collapseChainLabel(["Docs", "Viagem"])).toBe("Docs › Viagem");
    expect(collapseChainLabel(["Docs"])).toBe("Docs");
    expect(collapseChainLabel([])).toBe("");
  });

  it("hides sidecars from listings", () => {
    expect(shouldShowInTrash("a.jpg.trashinfo")).toBe(false);
    expect(shouldShowInTrash("a.jpg")).toBe(true);
  });
});

describe("mergeTrashResources", () => {
  const item = (name: string, index: number) =>
    ({ name, index, isDir: false }) as unknown as ResourceItem;
  const res = (names: string[]) =>
    ({
      path: "/.Trash/files",
      name: "files",
      items: names.map(item),
      numDirs: 0,
      numFiles: names.length,
    }) as unknown as Resource;

  it("returns whichever side exists", () => {
    expect(mergeTrashResources(null, res(["a"]))).toEqual(res(["a"]));
    expect(mergeTrashResources(res(["a"]), null)).toEqual(res(["a"]));
    expect(mergeTrashResources(null, null)).toBeNull();
  });

  it("concatenates and re-indexes", () => {
    const merged = mergeTrashResources(res(["a"]), res(["b", "c"]));
    expect(merged!.items.map((i) => i.name)).toEqual(["a", "b", "c"]);
    expect(merged!.items.map((i) => i.index)).toEqual([0, 1, 2]);
    expect(merged!.numFiles).toBe(3);
  });
});

describe("buildFlatTrashResource", () => {
  it("maps recursive entries to selectable items", () => {
    const res = buildFlatTrashResource(
      [
        {
          path: "/.Trash/files/Docs",
          name: "Docs",
          size: 0,
          modified: "2026-09-08T00:00:00Z",
          isDir: true,
        },
        {
          path: "/.Trash/files/Docs/a.jpg",
          name: "a.jpg",
          size: 10,
          modified: "2026-09-08T00:00:00Z",
          isDir: false,
        },
      ],
      "/files/2",
      "2"
    );
    expect(res.items).toHaveLength(2);
    expect(res.items[0].url).toBe("/files/2/.Trash/files/Docs/");
    expect(res.items[1].url).toBe("/files/2/.Trash/files/Docs/a.jpg");
    expect(res.items[1].index).toBe(1);
    expect(res.numDirs).toBe(1);
    expect(res.numFiles).toBe(1);
  });
});

describe("topLevelNames / sortByDepthDesc", () => {
  it("extracts first segments", () => {
    expect(
      topLevelNames(
        [
          "/.Trash/files/Docs/a.jpg",
          "/.Trash/files/Fotos/2024/b.jpg",
          "/.Trash/files/loose.txt",
        ],
        "/.Trash/files"
      ).sort()
    ).toEqual(["Docs", "Fotos", "loose.txt"]);
  });

  it("sorts deepest first", () => {
    expect(sortByDepthDesc(["/a", "/a/b/c", "/a/b"])).toEqual([
      "/a/b/c",
      "/a/b",
      "/a",
    ]);
  });
});
