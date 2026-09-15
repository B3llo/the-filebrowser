import { describe, it, expect } from "vitest";
import {
  buildSearchSections,
  escapeHtml,
  highlightMatch,
  highlightTermFromPrompt,
  iconForSearchResult,
  splitSearchPath,
} from "@/utils/searchResults";

describe("iconForSearchResult", () => {
  it("classifies dirs and extensions", () => {
    expect(iconForSearchResult("Docs", true)).toBe("folder");
    expect(iconForSearchResult("a.jpg", false)).toBe("image");
    expect(iconForSearchResult("a.mp4", false)).toBe("video");
    expect(iconForSearchResult("a.mp3", false)).toBe("audio");
    expect(iconForSearchResult("a.pdf", false)).toBe("pdf");
    expect(iconForSearchResult("a.ts", false)).toBe("code");
    expect(iconForSearchResult("a.zip", false)).toBe("archive");
    expect(iconForSearchResult("README", false)).toBe("file");
    expect(iconForSearchResult(".gitignore", false)).toBe("file");
  });
});

describe("splitSearchPath", () => {
  it("splits name and parent dir", () => {
    expect(splitSearchPath("Docs/Viagem/f.jpg")).toEqual({
      name: "f.jpg",
      dir: "Docs/Viagem",
    });
    expect(splitSearchPath("a.jpg")).toEqual({ name: "a.jpg", dir: "" });
    expect(splitSearchPath("Docs/")).toEqual({ name: "Docs", dir: "" });
  });
});

describe("escapeHtml / highlightMatch", () => {
  it("escapes html", () => {
    expect(escapeHtml('<a href="x">&')).toBe(
      "&lt;a href=&quot;x&quot;&gt;&amp;"
    );
  });

  it("wraps the first case-insensitive match in mark", () => {
    expect(highlightMatch("Report.pdf", "rep")).toBe("<mark>Rep</mark>ort.pdf");
    expect(highlightMatch("a.jpg", "zzz")).toBe("a.jpg");
    expect(highlightMatch("a.jpg")).toBe("a.jpg");
    expect(highlightMatch("<b>.jpg", ".jpg")).toBe(
      "&lt;b&gt;<mark>.jpg</mark>"
    );
  });
});

describe("highlightTermFromPrompt", () => {
  it("uses the last non-type token minus leading dot", () => {
    expect(highlightTermFromPrompt("report")).toBe("report");
    expect(highlightTermFromPrompt("type:pdf report")).toBe("report");
    expect(highlightTermFromPrompt("type:pdf")).toBe("");
    expect(highlightTermFromPrompt(".gitignore")).toBe("gitignore");
    expect(highlightTermFromPrompt("  ")).toBe("");
  });
});

describe("buildSearchSections", () => {
  it("groups folders first with icons and parent dirs", () => {
    const sections = buildSearchSections([
      { path: "Docs/a.jpg", url: "/files/0/Docs/a.jpg", isDir: false },
      { path: "Docs/Sub", url: "/files/0/Docs/Sub/", isDir: true },
      { path: "root.png", url: "/files/0/root.png", isDir: false },
    ]);
    expect(sections.map((s) => s.id)).toEqual(["folders", "files"]);
    expect(sections[0].entries).toHaveLength(1);
    expect(sections[0].entries[0]).toMatchObject({
      icon: "folder",
      name: "Sub",
      dir: "Docs",
    });
    expect(sections[1].entries.map((e) => e.icon)).toEqual(["image", "image"]);
    expect(sections[1].entries[1].dir).toBe("");
  });

  it("omits empty sections", () => {
    expect(
      buildSearchSections([
        { path: "a.jpg", url: "/files/0/a.jpg", isDir: false },
      ]).map((s) => s.id)
    ).toEqual(["files"]);
    expect(buildSearchSections([])).toEqual([]);
  });
});
