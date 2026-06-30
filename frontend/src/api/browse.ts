import { fetchJSON, fetchURL } from "./utils";

export interface BrowseDirEntry {
  name: string;
  path: string;
  hasChildren: boolean;
}

export interface SourceDir {
  name: string;
  path: string; // absolute path within source scope (e.g. "/docs/")
  url: string; // router URL (e.g. "/files/1/docs/")
}

// browse lists immediate subdirectories on the host filesystem (admin-only).
// Used by DirectoryPicker when adding a new source.
export async function browse(path: string): Promise<BrowseDirEntry[]> {
  return fetchJSON<BrowseDirEntry[]>(
    `/api/browse?path=${encodeURIComponent(path)}`
  );
}

// listDirs lists immediate subdirectories within a source scope.
// Uses the existing resources API with the X-Source header to scope to the
// given source without changing the globally active source.
export async function listDirs(
  sourceId: string | number,
  path: string = "/"
): Promise<SourceDir[]> {
  const apiPath = path === "/" ? "/" : path;
  const res = await fetchURL(`/api/resources${apiPath}`, {
    headers: { "X-Source": String(sourceId) },
  });

  const data = await res.json();
  if (!data.isDir || !Array.isArray(data.items)) return [];

  return data.items
    .filter((item: any) => item.isDir === true)
    .map((item: any) => {
      const childPath = apiPath === "/" ? `/${item.name}/` : `${apiPath}${item.name}/`;
      return {
        name: item.name,
        path: childPath,
        url: `/files/${sourceId}${childPath}`,
      };
    });
}
