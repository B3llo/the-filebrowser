import { fetchURL, fetchJSON, removePrefix } from "./utils";

export type GrantRole = "viewer" | "editor";

export interface GrantCreate {
  path: string;
  grantee: string;
  owner?: string;
  role: GrantRole;
  expires?: string;
  unit?: string;
}

export interface GrantPatch {
  role?: GrantRole;
  expires?: string;
  unit?: string;
}

export async function list() {
  return fetchJSON<Grant[]>("/api/grants");
}

export async function sharedWithMe() {
  return fetchJSON<Grant[]>("/api/grants/shared-with-me");
}

export async function create(input: GrantCreate) {
  return fetchJSON<Grant>("/api/grants", {
    method: "POST",
    body: JSON.stringify({
      path: removePrefix(input.path),
      grantee: input.grantee,
      ...(input.owner ? { owner: input.owner } : {}),
      role: input.role,
      ...(input.expires ? { expires: input.expires } : {}),
      ...(input.unit ? { unit: input.unit } : {}),
    }),
  });
}

export async function remove(id: number) {
  await fetchURL(`/api/grants/${id}`, {
    method: "DELETE",
  });
}

export async function patch(id: number, input: GrantPatch) {
  return fetchJSON<Grant>(`/api/grants/${id}`, {
    method: "PATCH",
    body: JSON.stringify(input),
  });
}
