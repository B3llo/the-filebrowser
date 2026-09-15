/**
 * Shared search-result presentation used by the header search dropdown and
 * the Ctrl+K command palette: per-type icons, Folders/Files sections,
 * match highlighting and parent-dir subtitles.
 *
 * Pure (no API imports) so it stays unit-testable.
 */

import type { IconName } from "@/utils/icons";

export interface SearchResultEntry {
  key: string;
  icon: IconName;
  name: string;
  /** Parent directory shown as dimmed subtitle. */
  dir: string;
  url: string;
  isDir: boolean;
}

export interface SearchResultSection {
  id: string;
  entries: SearchResultEntry[];
}

const IMAGE_EXTS = new Set([
  ".jpg",
  ".jpeg",
  ".png",
  ".gif",
  ".webp",
  ".bmp",
  ".svg",
  ".ico",
  ".avif",
  ".heic",
  ".tiff",
]);
const VIDEO_EXTS = new Set([
  ".mp4",
  ".mov",
  ".webm",
  ".mkv",
  ".avi",
  ".m4v",
  ".wmv",
  ".flv",
]);
const AUDIO_EXTS = new Set([
  ".mp3",
  ".wav",
  ".flac",
  ".ogg",
  ".m4a",
  ".aac",
  ".opus",
]);
const CODE_EXTS = new Set([
  ".js",
  ".mjs",
  ".cjs",
  ".ts",
  ".tsx",
  ".jsx",
  ".vue",
  ".go",
  ".py",
  ".rb",
  ".rs",
  ".java",
  ".kt",
  ".c",
  ".cpp",
  ".h",
  ".cs",
  ".php",
  ".sh",
  ".html",
  ".css",
  ".scss",
  ".json",
  ".xml",
  ".yml",
  ".yaml",
  ".toml",
  ".sql",
]);
const ARCHIVE_EXTS = new Set([
  ".zip",
  ".rar",
  ".7z",
  ".tar",
  ".gz",
  ".bz2",
  ".xz",
  ".tgz",
]);

const extOf = (name: string): string => {
  const i = name.lastIndexOf(".");
  return i <= 0 ? "" : name.slice(i).toLowerCase();
};

/** Icon from filename extension (backend search results carry no type). */
export function iconForSearchResult(name: string, isDir: boolean): IconName {
  if (isDir) return "folder";
  const ext = extOf(name);
  if (IMAGE_EXTS.has(ext)) return "image";
  if (VIDEO_EXTS.has(ext)) return "video";
  if (AUDIO_EXTS.has(ext)) return "audio";
  if (ext === ".pdf") return "pdf";
  if (CODE_EXTS.has(ext)) return "code";
  if (ARCHIVE_EXTS.has(ext)) return "archive";
  return "file";
}

/**
 * Split a raw (already-decoded) FS path into name + parent directory.
 * The search backend returns raw paths, so no URI decoding here.
 */
export function splitSearchPath(path: string): { name: string; dir: string } {
  const clean = path.replace(/\/$/, "");
  const slash = clean.lastIndexOf("/");
  return slash === -1
    ? { name: clean, dir: "" }
    : { name: clean.slice(slash + 1), dir: clean.slice(0, slash) };
}

const HTML_ESCAPES: Record<string, string> = {
  "&": "&amp;",
  "<": "&lt;",
  ">": "&gt;",
  '"': "&quot;",
};

export function escapeHtml(s: string): string {
  return s.replace(/[&<>"]/g, (c) => HTML_ESCAPES[c]);
}

/** Name with the first case-insensitive match of `q` wrapped in <mark>. */
export function highlightMatch(text: string, q?: string): string {
  if (!q) return escapeHtml(text);
  const idx = text.toLowerCase().indexOf(q.toLowerCase());
  if (idx === -1) return escapeHtml(text);
  return (
    escapeHtml(text.slice(0, idx)) +
    "<mark>" +
    escapeHtml(text.slice(idx, idx + q.length)) +
    "</mark>" +
    escapeHtml(text.slice(idx + q.length))
  );
}

/**
 * Substring to highlight from a raw search prompt: the last non-`type:`
 * token (minus a leading dot used for the extension shortcut), so
 * `type:pdf report` highlights "report" instead of the whole query.
 */
export function highlightTermFromPrompt(prompt: string): string {
  const tokens = prompt.trim().split(/\s+/).filter(Boolean);
  for (let i = tokens.length - 1; i >= 0; i--) {
    const token = tokens[i];
    if (!token.includes(":")) {
      return token.startsWith(".") ? token.slice(1) : token;
    }
  }
  return "";
}

/**
 * Group raw backend results into Folders/Files sections with icons and
 * parent-dir subtitles. Order inside each section is preserved.
 */
export function buildSearchSections(
  items: { path: string; url: string; isDir: boolean }[]
): SearchResultSection[] {
  const toEntry = (item: {
    path: string;
    url: string;
    isDir: boolean;
  }): SearchResultEntry => {
    const { name, dir } = splitSearchPath(item.path);
    return {
      key: `res:${item.url}`,
      icon: iconForSearchResult(name, item.isDir),
      name,
      dir,
      url: item.url,
      isDir: item.isDir,
    };
  };
  const sections: SearchResultSection[] = [];
  const folders = items.filter((r) => r.isDir);
  const files = items.filter((r) => !r.isDir);
  if (folders.length)
    sections.push({ id: "folders", entries: folders.map(toEntry) });
  if (files.length) sections.push({ id: "files", entries: files.map(toEntry) });
  return sections;
}
