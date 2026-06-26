export type FileKind =
  | "folder"
  | "image"
  | "video"
  | "audio"
  | "pdf"
  | "doc"
  | "sheet"
  | "slide"
  | "archive"
  | "code"
  | "vector"
  | "text"
  | "default";

const VECTOR_EXTS = new Set([
  ".svg", ".eps", ".fig", ".ai",
]);
const DOC_EXTS = new Set([
  ".doc", ".docx", ".odt", ".rtf", ".pages", ".log",
]);
const SHEET_EXTS = new Set([
  ".xls", ".xlsx", ".ods", ".csv", ".numbers", ".db", ".sqlite",
]);
const SLIDE_EXTS = new Set([
  ".ppt", ".pptx", ".odp", ".key",
]);
const ARCHIVE_EXTS = new Set([
  ".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".xz", ".zst",
  ".dmg", ".iso", ".deb", ".rpm", ".pkg", ".msi", ".apk",
  ".cab", ".tgz",
]);
const CODE_EXTS = new Set([
  ".js", ".ts", ".jsx", ".tsx", ".mjs", ".cjs", ".vue", ".svelte",
  ".json", ".html", ".htm", ".css", ".scss", ".sass", ".less",
  ".go", ".py", ".rb", ".rs", ".java", ".kt", ".c", ".cpp", ".h", ".hpp",
  ".cs", ".php", ".xml", ".yml", ".yaml", ".toml", ".ini", ".sh", ".bat",
  ".ps1", ".swift", ".dart", ".lua", ".pl", ".r", ".scala", ".clj",
]);

function extOf(name: string): string {
  const lower = name.toLowerCase();
  const i = lower.lastIndexOf(".");
  return i === -1 ? "" : lower.slice(i);
}

/**
 * Map a Resource-like item to a coarse file "kind" used for tinting and
 * iconography in the listing views. Extension wins over backend mimetype
 * `type` so that e.g. .xlsx is classified as "sheet" even though the backend
 * reports it as "blob".
 */
export function fileKind(item: {
  isDir?: boolean;
  type?: string;
  name?: string;
}): FileKind {
  if (item.isDir) return "folder";

  const ext = extOf(item.name ?? "");
  if (VECTOR_EXTS.has(ext)) return "vector";
  if (DOC_EXTS.has(ext)) return "doc";
  if (SHEET_EXTS.has(ext)) return "sheet";
  if (SLIDE_EXTS.has(ext)) return "slide";
  if (ARCHIVE_EXTS.has(ext)) return "archive";
  if (CODE_EXTS.has(ext)) return "code";
  if (ext === ".pdf") return "pdf";

  switch (item.type) {
    case "image":
      return "image";
    case "video":
      return "video";
    case "audio":
      return "audio";
    case "text":
    case "textImmutable":
      return "text";
    default:
      return "default";
  }
}

/**
 * Short (≤4 char), uppercased extension label rendered inside the document
 * shape on mosaic cards. Matches the design's `extLabel` formatting.
 */
export function extLabel(name: string): string {
  const i = name.lastIndexOf(".");
  const ext = i === -1 ? name : name.slice(i + 1);
  return ext.toUpperCase().slice(0, 4);
}
