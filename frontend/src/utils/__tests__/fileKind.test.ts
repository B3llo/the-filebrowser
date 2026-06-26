import { describe, it, expect } from "vitest";
import { fileKind, extLabel } from "@/utils/fileKind";

describe("fileKind", () => {
  it("classifies folders regardless of extension", () => {
    expect(fileKind({ isDir: true, name: "Photos" })).toBe("folder");
    expect(fileKind({ isDir: true, name: "archive.zip" })).toBe("folder");
  });

  it("classifies by backend mimetype type when no special extension", () => {
    expect(fileKind({ isDir: false, type: "image", name: "pic.jpg" })).toBe(
      "image"
    );
    expect(fileKind({ isDir: false, type: "video", name: "clip.mp4" })).toBe(
      "video"
    );
    expect(fileKind({ isDir: false, type: "audio", name: "song.mp3" })).toBe(
      "audio"
    );
    expect(fileKind({ isDir: false, type: "text", name: "notes.txt" })).toBe(
      "text"
    );
    expect(fileKind({ isDir: false, type: "textImmutable", name: "a.md" })).toBe(
      "text"
    );
  });

  it("extension wins over backend 'blob' type for office/archive/code kinds", () => {
    expect(fileKind({ isDir: false, type: "blob", name: "report.docx" })).toBe(
      "doc"
    );
    expect(fileKind({ isDir: false, type: "blob", name: "budget.xlsx" })).toBe(
      "sheet"
    );
    expect(fileKind({ isDir: false, type: "blob", name: "deck.pptx" })).toBe(
      "slide"
    );
    expect(fileKind({ isDir: false, type: "blob", name: "backup.zip" })).toBe(
      "archive"
    );
    expect(fileKind({ isDir: false, type: "blob", name: "app.go" })).toBe(
      "code"
    );
    expect(fileKind({ isDir: false, type: "blob", name: "logo.svg" })).toBe(
      "vector"
    );
  });

  it("classifies pdf directly", () => {
    expect(fileKind({ isDir: false, type: "pdf", name: "doc.pdf" })).toBe("pdf");
    expect(fileKind({ isDir: false, type: "blob", name: "doc.pdf" })).toBe(
      "pdf"
    );
  });

  it("falls back to default for unknown extensions and types", () => {
    expect(fileKind({ isDir: false, type: "blob", name: "data.xyz" })).toBe(
      "default"
    );
    expect(fileKind({ isDir: false, name: "noext" })).toBe("default");
  });

  it("is case-insensitive on extensions", () => {
    // image/png are classified via the backend mimetype `type`, not extension
    expect(fileKind({ isDir: false, type: "image", name: "IMG.PNG" })).toBe(
      "image"
    );
    // office/code kinds are extension-driven and must be case-insensitive
    expect(fileKind({ isDir: false, type: "blob", name: "CODE.VUE" })).toBe(
      "code"
    );
    expect(fileKind({ isDir: false, type: "blob", name: "Logo.SVG" })).toBe(
      "vector"
    );
  });
});

describe("extLabel", () => {
  it("uppercases and truncates to 4 chars", () => {
    expect(extLabel("photo.jpg")).toBe("JPG");
    expect(extLabel("archive.tar.gz")).toBe("GZ");
    expect(extLabel("a.verylongextension")).toBe("VERY");
  });

  it("falls back to the whole name when no extension", () => {
    expect(extLabel("README")).toBe("READ");
  });
});
