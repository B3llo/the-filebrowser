/**
 * Trash helpers — mirrored-structure trash, per source.
 *
 * Layout inside each source root (frontend convention, no backend changes):
 *
 *   .Trash/files/<original-relative-path>   mirrored file/folder tree, original names verbatim
 *   .Trash/info/<original-relative-path>.trashinfo   one JSON sidecar per trashed top-level entry
 *
 * Legacy layout (flat, pre-mirror) is still read for restore/empty:
 *
 *   .Trash/<epochMs>_<name>
 *
 * All helpers here are pure (no API imports) so they stay unit-testable.
 */

export const TRASH_DIR_NAME = ".Trash";
export const TRASH_FILES_DIR = "files";
export const TRASH_INFO_DIR = "info";
export const TRASH_INFO_EXT = ".trashinfo";

/** Source-relative prefix of the mirrored tree, e.g. "/.Trash/files". */
export const TRASH_FILES_PREFIX = `/${TRASH_DIR_NAME}/${TRASH_FILES_DIR}`;
/** Source-relative prefix of the legacy flat trash, e.g. "/.Trash". */
export const TRASH_ROOT_PREFIX = `/${TRASH_DIR_NAME}`;

const LEGACY_PREFIX_RE = /^\d+_/;

export interface TrashInfo {
  version: 1;
  /** Source-relative original path, e.g. "/Docs/Viagem/f1.jpg". */
  originalPath: string;
  name: string;
  isDir: boolean;
  /** ISO-8601 deletion timestamp. */
  deletedAt: string;
  sourceId: string;
  /** Shared Date.now() of the delete batch, for "deleted together" grouping. */
  batchId: number;
}

/** Strip the legacy `<epochMs>_` prefix; new mirrored names pass through. */
export function cleanTrashName(name: string): string {
  return name.replace(LEGACY_PREFIX_RE, "");
}

export function isLegacyTrashName(name: string): boolean {
  return LEGACY_PREFIX_RE.test(name);
}

/** `/files/<id>` + `/.Trash/files` (no trailing slash). */
export function trashFilesRoot(filesBase: string): string {
  return `${filesBase}/${TRASH_DIR_NAME}/${TRASH_FILES_DIR}`;
}

/** `/files/<id>` + `/.Trash/info` (no trailing slash). */
export function trashInfoRoot(filesBase: string): string {
  return `${filesBase}/${TRASH_DIR_NAME}/${TRASH_INFO_DIR}`;
}

/** `/files/<id>` + `/.Trash` (no trailing slash) — legacy root. */
export function legacyTrashRoot(filesBase: string): string {
  return `${filesBase}/${TRASH_DIR_NAME}`;
}

/** True when a source-relative path lives anywhere under `/.Trash`. */
export function isTrashPath(sourcePath: string): boolean {
  return (
    sourcePath === TRASH_ROOT_PREFIX ||
    sourcePath.startsWith(`${TRASH_ROOT_PREFIX}/`)
  );
}

/** True when a source-relative path lives under `/.Trash/files`. */
export function isMirroredTrashPath(sourcePath: string): boolean {
  return (
    sourcePath === TRASH_FILES_PREFIX ||
    sourcePath.startsWith(`${TRASH_FILES_PREFIX}/`)
  );
}

/**
 * Path inside the mirror for a source-relative original path.
 * `"/Docs/a.jpg"` -> `"/.Trash/files/Docs/a.jpg"`.
 */
export function mirroredTrashSourcePath(sourcePath: string): string {
  const rel = sourcePath.startsWith("/") ? sourcePath.slice(1) : sourcePath;
  return `${TRASH_FILES_PREFIX}/${rel}`;
}

/**
 * Inverse of mirroredTrashSourcePath: trash source-path -> original path.
 * Returns null when not inside the mirror.
 */
export function originalPathFromTrash(sourcePath: string): string | null {
  const rel = trashMirrorRelative(sourcePath);
  if (rel === null) return null;
  return `/${rel}`;
}

/**
 * Relative path inside `/.Trash/files` ("" for the mirror root itself),
 * or null when the path is not inside the mirror.
 */
export function trashMirrorRelative(sourcePath: string): string | null {
  if (sourcePath === TRASH_FILES_PREFIX) return "";
  const prefix = `${TRASH_FILES_PREFIX}/`;
  if (sourcePath.startsWith(prefix)) return sourcePath.slice(prefix.length);
  return null;
}

/** Parent dir of a source-relative path: "/Docs/Viagem/f" -> "/Docs/Viagem", "/f" -> "/". */
export function parentDir(sourcePath: string): string {
  if (!sourcePath || sourcePath === "/") return "/";
  const trimmed =
    sourcePath.endsWith("/") && sourcePath !== "/"
      ? sourcePath.slice(0, -1)
      : sourcePath;
  const idx = trimmed.lastIndexOf("/");
  if (idx <= 0) return "/";
  return trimmed.slice(0, idx);
}

/** Basename of a source-relative path: "/Docs/a.jpg" -> "a.jpg". */
export function baseName(sourcePath: string): string {
  const trimmed =
    sourcePath.endsWith("/") && sourcePath !== "/"
      ? sourcePath.slice(0, -1)
      : sourcePath;
  const idx = trimmed.lastIndexOf("/");
  return idx === -1 ? trimmed : trimmed.slice(idx + 1);
}

/** Split "archive.tar.gz" -> { base: "archive.tar", ext: ".gz" }; dotfiles keep leading dot in base. */
export function splitName(name: string): { base: string; ext: string } {
  const idx = name.lastIndexOf(".");
  if (idx <= 0) return { base: name, ext: "" };
  return { base: name.slice(0, idx), ext: name.slice(idx) };
}

/** "f.jpg", 1 -> "f(1).jpg"; "Viagem", 2 -> "Viagem(2)". Counter 0 returns the name unchanged. */
export function withVersionSuffix(name: string, counter: number): string {
  if (counter <= 0) return name;
  const { base, ext } = splitName(name);
  return `${base}(${counter})${ext}`;
}

/**
 * Full `/files/<id>/...` destination URL inside the mirror for an original
 * source-relative path. `filesBase` is `/files/<id>` (no trailing slash).
 */
export function trashDestinationUrl(
  filesBase: string,
  sourcePath: string
): string {
  return `${filesBase}${mirroredTrashSourcePath(sourcePath)}`;
}

/**
 * Full `/files/<id>/...` sidecar URL for a trashed source-relative path
 * (under `/.Trash/files/...`). Folders get `<name>.trashinfo` too — the sidecar
 * always describes the top-level entry that was moved.
 */
export function trashInfoUrl(
  filesBase: string,
  trashSourcePath: string
): string | null {
  const rel = trashMirrorRelative(trashSourcePath);
  if (rel === null || rel === "") return null;
  return `${filesBase}/${TRASH_DIR_NAME}/${TRASH_INFO_DIR}/${rel}${TRASH_INFO_EXT}`;
}

/**
 * Trash source-relative path back from a full `/files/<id>/...` destination URL.
 * Returns null when the URL is not under this filesBase.
 */
export function trashSourcePathFromUrl(
  url: string,
  filesBase: string
): string | null {
  const base = url.split("?")[0];
  let rel: string;
  if (base.startsWith(`${filesBase}/`)) {
    rel = base.slice(filesBase.length);
  } else if (base === filesBase) {
    return "/";
  } else {
    return null;
  }
  try {
    rel = decodeURIComponent(rel);
  } catch {
    // Keep the raw segment on malformed encodings; comparisons still work.
  }
  if (!rel.startsWith("/")) rel = `/${rel}`;
  return rel;
}

export function buildTrashInfo(input: {
  originalPath: string;
  name: string;
  isDir: boolean;
  deletedAt: string;
  sourceId: string;
  batchId: number;
}): TrashInfo {
  return {
    version: 1,
    originalPath: input.originalPath,
    name: input.name,
    isDir: input.isDir,
    deletedAt: input.deletedAt,
    sourceId: input.sourceId,
    batchId: input.batchId,
  };
}

/** Parse sidecar JSON; null on any malformed payload. */
export function parseTrashInfo(raw: string): TrashInfo | null {
  try {
    const data = JSON.parse(raw) as Partial<TrashInfo>;
    if (
      typeof data.originalPath !== "string" ||
      !data.originalPath.startsWith("/") ||
      typeof data.name !== "string" ||
      data.name.length === 0 ||
      typeof data.deletedAt !== "string" ||
      typeof data.sourceId !== "string" ||
      typeof data.batchId !== "number"
    ) {
      return null;
    }
    return {
      version: 1,
      originalPath: data.originalPath,
      name: data.name,
      isDir: data.isDir === true,
      deletedAt: data.deletedAt,
      sourceId: data.sourceId,
      batchId: data.batchId,
    };
  } catch {
    return null;
  }
}

/** Display location of an original path: parent dir, "/" for top-level items. */
export function originalLocation(originalPath: string): string {
  return parentDir(originalPath);
}

/**
 * Collapse a chain of single-child folders for display: ["Docs", "Viagem"] ->
 * "Docs › Viagem". Used for flat-view "original location" cells and for the
 * collapsed breadcrumb chip so one trashed file doesn't force deep drilling.
 */
export function collapseChainLabel(parts: string[]): string {
  return parts.filter((p) => p !== "").join(" › ");
}

/** Sidecar/aux files that must never show up in the trash listing. */
export function shouldShowInTrash(name: string): boolean {
  return !name.endsWith(TRASH_INFO_EXT);
}

/**
 * Merge the new mirrored listing with legacy flat items into a single Resource.
 * The legacy listing must already exclude the `files`/`info` dirs. Items are
 * re-indexed so selection indices stay valid. Returns whichever side exists
 * when only one does.
 */
export function mergeTrashResources(
  mirrored: Resource | null,
  legacy: Resource | null
): Resource | null {
  if (!mirrored) return legacy;
  if (!legacy) return mirrored;
  const items = [...mirrored.items, ...legacy.items].map(
    (item, index) => ({ ...item, index }) as ResourceItem
  );
  return {
    ...mirrored,
    items,
    numDirs: items.filter((i) => i.isDir).length,
    numFiles: items.filter((i) => !i.isDir).length,
  };
}

/**
 * Build a synthetic flat Resource from a recursive listing of the mirror root,
 * so the flat "all items" view can reuse FileListing (search/sort/selection).
 * `filesBase` is `/files/<id>`; entry paths are source-relative.
 */
export function buildFlatTrashResource(
  entries: RecursiveEntry[],
  filesBase: string,
  sourceId: string
): Resource {
  void sourceId;
  const items: ResourceItem[] = entries
    .filter(
      (e) =>
        e.path !== TRASH_FILES_PREFIX && e.path !== `${TRASH_FILES_PREFIX}/`
    )
    .map((e, index) => {
      const url = `${filesBase}${e.path}`;
      return {
        path: e.path,
        name: e.name,
        size: e.size,
        extension: e.name.includes(".")
          ? e.name.slice(e.name.lastIndexOf("."))
          : "",
        modified:
          typeof e.modified === "string"
            ? e.modified
            : new Date(e.modified).toISOString(),
        mode: 0,
        isDir: e.isDir,
        isSymlink: false,
        type: (e.isDir ? "dir" : "blob") as ResourceType,
        url: e.isDir && !url.endsWith("/") ? `${url}/` : url,
        index,
      } as ResourceItem;
    });
  return {
    path: TRASH_FILES_PREFIX,
    name: TRASH_FILES_DIR,
    size: 0,
    extension: "",
    modified: new Date().toISOString(),
    mode: 0,
    isDir: true,
    isSymlink: false,
    type: "dir",
    url: `${filesBase}${TRASH_FILES_PREFIX}/`,
    items,
    numDirs: items.filter((i) => i.isDir).length,
    numFiles: items.filter((i) => !i.isDir).length,
    sorting: { by: "name", asc: true },
    index: 0,
  } as Resource;
}

/** Top-level entry names (first segment) below a root, for recursive empty-trash. */
export function topLevelNames(paths: string[], rootPrefix: string): string[] {
  const prefix = rootPrefix.endsWith("/") ? rootPrefix : `${rootPrefix}/`;
  const names = new Set<string>();
  for (const p of paths) {
    if (!p.startsWith(prefix)) continue;
    const rest = p.slice(prefix.length).replace(/\/$/, "");
    if (!rest) continue;
    names.add(rest.split("/")[0]);
  }
  return [...names];
}

/** Deepest paths first, so individual removes never orphan parents. */
export function sortByDepthDesc(paths: string[]): string[] {
  return [...paths].sort(
    (a, b) => b.split("/").length - a.split("/").length || (a < b ? 1 : -1)
  );
}
