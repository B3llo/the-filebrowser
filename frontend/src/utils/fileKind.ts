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

const VECTOR_EXTS = new Set([".svg", ".eps", ".fig", ".ai"]);
const DOC_EXTS = new Set([".doc", ".docx", ".odt", ".rtf", ".pages", ".log"]);
const SHEET_EXTS = new Set([
  ".xls",
  ".xlsx",
  ".ods",
  ".csv",
  ".numbers",
  ".db",
  ".sqlite",
]);
const SLIDE_EXTS = new Set([".ppt", ".pptx", ".odp", ".key"]);
const ARCHIVE_EXTS = new Set([
  ".zip",
  ".rar",
  ".7z",
  ".tar",
  ".gz",
  ".bz2",
  ".xz",
  ".zst",
  ".dmg",
  ".iso",
  ".deb",
  ".rpm",
  ".pkg",
  ".msi",
  ".apk",
  ".cab",
  ".tgz",
]);
const CODE_EXTS = new Set([
  ".js",
  ".ts",
  ".jsx",
  ".tsx",
  ".mjs",
  ".cjs",
  ".vue",
  ".svelte",
  ".json",
  ".html",
  ".htm",
  ".css",
  ".scss",
  ".sass",
  ".less",
  ".go",
  ".py",
  ".rb",
  ".rs",
  ".java",
  ".kt",
  ".c",
  ".cpp",
  ".h",
  ".hpp",
  ".cs",
  ".php",
  ".xml",
  ".yml",
  ".yaml",
  ".toml",
  ".ini",
  ".sh",
  ".bat",
  ".ps1",
  ".swift",
  ".dart",
  ".lua",
  ".pl",
  ".r",
  ".scala",
  ".clj",
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

/**
 * File type label for the details panel "Kind" row.
 * Common types get readable names (e.g. "Markdown", "JSON", "YAML").
 * Unknown types show the raw extension uppercased (e.g. "DSTOR", "XYZ").
 * Special filenames without extensions get a readable label.
 */

// Readable labels for well-known extensions
const EXT_LABELS: Record<string, string> = {
  // Text / Markup
  ".md": "Markdown",
  ".markdown": "Markdown",
  ".txt": "Text",
  ".log": "Log",
  ".csv": "CSV",
  // Web
  ".html": "HTML",
  ".htm": "HTML",
  ".css": "CSS",
  ".scss": "SCSS",
  ".sass": "Sass",
  ".less": "Less",
  ".xml": "XML",
  // Data
  ".json": "JSON",
  ".yml": "YAML",
  ".yaml": "YAML",
  ".toml": "TOML",
  ".ini": "INI",
  ".env": "ENV",
  // Code
  ".js": "JavaScript",
  ".mjs": "ES Module",
  ".cjs": "CommonJS",
  ".ts": "TypeScript",
  ".jsx": "React JSX",
  ".tsx": "React TSX",
  ".vue": "Vue",
  ".svelte": "Svelte",
  ".py": "Python",
  ".rb": "Ruby",
  ".go": "Go",
  ".rs": "Rust",
  ".java": "Java",
  ".kt": "Kotlin",
  ".c": "C",
  ".cpp": "C++",
  ".h": "C Header",
  ".hpp": "C++ Header",
  ".cs": "C#",
  ".php": "PHP",
  ".swift": "Swift",
  ".dart": "Dart",
  ".lua": "Lua",
  ".pl": "Perl",
  ".r": "R",
  ".scala": "Scala",
  ".clj": "Clojure",
  ".sh": "Shell",
  ".bat": "Batch",
  ".ps1": "PowerShell",
  // Documents
  ".pdf": "PDF",
  ".doc": "Word",
  ".docx": "Word",
  ".odt": "ODT",
  ".rtf": "RTF",
  ".pages": "Pages",
  // Spreadsheets
  ".xls": "Excel",
  ".xlsx": "Excel",
  ".ods": "ODS",
  ".numbers": "Numbers",
  // Presentations
  ".ppt": "PowerPoint",
  ".pptx": "PowerPoint",
  ".odp": "ODP",
  ".key": "Keynote",
  // Images
  ".jpg": "JPEG",
  ".jpeg": "JPEG",
  ".png": "PNG",
  ".gif": "GIF",
  ".webp": "WebP",
  ".bmp": "BMP",
  ".tiff": "TIFF",
  ".tif": "TIFF",
  ".ico": "ICO",
  ".svg": "SVG",
  ".eps": "EPS",
  ".fig": "Figma",
  ".ai": "Illustrator",
  ".psd": "Photoshop",
  ".heic": "HEIC",
  ".heif": "HEIF",
  // Video
  ".mp4": "MP4",
  ".mov": "MOV",
  ".avi": "AVI",
  ".mkv": "MKV",
  ".webm": "WebM",
  ".wmv": "WMV",
  ".flv": "FLV",
  ".m4v": "M4V",
  // Audio
  ".mp3": "MP3",
  ".wav": "WAV",
  ".flac": "FLAC",
  ".ogg": "OGG",
  ".aac": "AAC",
  ".m4a": "M4A",
  ".wma": "WMA",
  // Archives
  ".zip": "ZIP",
  ".rar": "RAR",
  ".7z": "7-Zip",
  ".tar": "Tar",
  ".gz": "Gzip",
  ".bz2": "Bzip2",
  ".xz": "XZ",
  ".zst": "Zstandard",
  ".dmg": "Disk Image",
  ".iso": "Disc Image",
  ".deb": "Debian",
  ".rpm": "RPM",
  ".msi": "MSI",
  ".apk": "APK",
  ".cab": "Cabinet",
  ".pkg": "Package",
  ".tgz": "Tarball",
  // Fonts
  ".ttf": "TTF",
  ".otf": "OTF",
  ".woff": "WOFF",
  ".woff2": "WOFF2",
};

// Special filenames (no extension) that get a readable label
const NAME_LABELS: Record<string, string> = {
  dockerfile: "Dockerfile",
  "docker-compose.yml": "Docker Compose",
  "docker-compose.yaml": "Docker Compose",
  ".gitignore": "Git Ignore",
  ".gitattributes": "Git Attributes",
  ".editorconfig": "EditorConfig",
  ".prettierrc": "Prettier",
  ".eslintrc": "ESLint",
  makefile: "Makefile",
  cmake: "CMake",
  license: "License",
  readme: "Readme",
  changelog: "Changelog",
};

export function fileTypeLabel(item: {
  isDir?: boolean;
  type?: string;
  name?: string;
}): string {
  if (item.isDir) return "Folder";

  const name = item.name ?? "";
  const lowerName = name.toLowerCase();

  // Check exact name matches for special files (e.g. "Dockerfile")
  if (NAME_LABELS[lowerName]) return NAME_LABELS[lowerName];

  // Check extension for known types
  const ext = extOf(name);
  if (EXT_LABELS[ext]) return EXT_LABELS[ext];

  // Unknown extension — show raw extension uppercased
  if (ext) return ext.slice(1).toUpperCase();

  // No extension at all — fall back to backend type
  switch (item.type) {
    case "image":
      return "Image";
    case "video":
      return "Video";
    case "audio":
      return "Audio";
    case "text":
    case "textImmutable":
      return "Text";
    case "pdf":
      return "PDF";
    default:
      return "File";
  }
}
