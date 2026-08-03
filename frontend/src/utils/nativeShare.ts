import { fetchURL } from "@/api/utils";
import { buildRawURL } from "@/api/files";

export interface ShareableItem {
  url: string;
  name: string;
  isDir?: boolean;
}

export type ShareOutcome = "shared" | "canceled" | "unsupported";

const MIME_TYPES: Record<string, string> = {
  // images
  png: "image/png",
  jpg: "image/jpeg",
  jpeg: "image/jpeg",
  gif: "image/gif",
  webp: "image/webp",
  svg: "image/svg+xml",
  bmp: "image/bmp",
  avif: "image/avif",
  ico: "image/x-icon",
  // video
  mp4: "video/mp4",
  webm: "video/webm",
  mov: "video/quicktime",
  mkv: "video/x-matroska",
  avi: "video/x-msvideo",
  m4v: "video/x-m4v",
  ogv: "video/ogg",
  // audio
  mp3: "audio/mpeg",
  wav: "audio/wav",
  ogg: "audio/ogg",
  flac: "audio/flac",
  m4a: "audio/mp4",
  aac: "audio/aac",
  opus: "audio/opus",
  // documents
  pdf: "application/pdf",
  doc: "application/msword",
  docx: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
  xls: "application/vnd.ms-excel",
  xlsx: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
  ppt: "application/vnd.ms-powerpoint",
  pptx: "application/vnd.openxmlformats-officedocument.presentationml.presentation",
  zip: "application/zip",
  gz: "application/gzip",
  "7z": "application/x-7z-compressed",
  rar: "application/vnd.rar",
  // text / code
  txt: "text/plain",
  md: "text/markdown",
  json: "application/json",
  csv: "text/csv",
  xml: "application/xml",
  html: "text/html",
  htm: "text/html",
  css: "text/css",
  js: "text/javascript",
  mjs: "text/javascript",
  ts: "text/typescript",
  tsx: "text/typescript",
  jsx: "text/javascript",
  log: "text/plain",
  yaml: "text/yaml",
  yml: "text/yaml",
  toml: "text/plain",
  ini: "text/plain",
  conf: "text/plain",
};

function mimeFromName(name: string): string {
  const base = name.split(".").pop()?.toLowerCase() ?? "";
  return MIME_TYPES[base] ?? "application/octet-stream";
}

function parseContentDispositionFilename(header: string | null): string | null {
  if (!header) return null;

  const encoded = header.match(/filename\*=utf-8''([^;]+)/i);
  if (encoded) {
    try {
      return decodeURIComponent(encoded[1]);
    } catch {
      return encoded[1];
    }
  }

  const plain = header.match(/filename="?([^";]+)"?/i);
  return plain?.[1] ?? null;
}

function zipFallbackName(items: ShareableItem[]): string {
  if (items.length === 1) {
    return `${items[0].name.replace(/\/$/, "")}.zip`;
  }
  return "files.zip";
}

export function canShareFiles(): boolean {
  if (typeof navigator === "undefined") return false;
  if (
    typeof navigator.canShare !== "function" ||
    typeof navigator.share !== "function"
  ) {
    return false;
  }

  try {
    return navigator.canShare({
      files: [new File(["probe"], "probe.txt", { type: "text/plain" })],
    });
  } catch {
    return false;
  }
}

/**
 * Shares items through the OS native share sheet (WhatsApp, AirDrop, …).
 *
 * The bytes are streamed into a blob in memory and handed to the operating
 * system — nothing is written to the local disk. Folders and multi-selections
 * are bundled into a zip by the backend on the fly.
 *
 * Returns:
 *   - "shared"     the share sheet handed the item(s) to another app
 *   - "canceled"   the user dismissed the share sheet
 *   - "unsupported" the browser cannot share files (caller should fall back)
 */
export async function shareViaOS(
  items: ShareableItem[]
): Promise<ShareOutcome> {
  if (!canShareFiles()) return "unsupported";

  const isZip = items.length > 1 || items[0]?.isDir === true;
  const response = await fetchURL(
    buildRawURL(
      items.map((item) => item.url),
      isZip ? "zip" : null
    ),
    {}
  );
  const blob = await response.blob();

  const name =
    parseContentDispositionFilename(
      response.headers.get("Content-Disposition")
    ) ?? (isZip ? zipFallbackName(items) : (items[0]?.name ?? "file"));
  const type = isZip ? "application/zip" : mimeFromName(name);

  try {
    await navigator.share({
      files: [new File([blob], name, { type })],
      title: items[0]?.name ?? "file",
    });
    return "shared";
  } catch (e: any) {
    if (e?.name === "AbortError") return "canceled";
    if (e?.name === "NotAllowedError") return "unsupported";
    throw e;
  }
}
