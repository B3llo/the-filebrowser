/**
 * Pure helpers for user-to-user grants (shared-with-me view and friends).
 * Kept side-effect free so they are trivially unit-testable.
 */

export function grantDisplayName(path: string): string {
  if (!path || path === "/") return "/";
  const parts = decodeURIComponent(path).split("/").filter(Boolean);
  return parts.length > 0 ? parts[parts.length - 1] : "/";
}

export function filterGrants(grants: Grant[], query: string): Grant[] {
  const q = (query || "").trim().toLowerCase();
  if (!q) return grants;
  return grants.filter((g) => {
    const hay = `${g.path} ${g.ownerUsername ?? ""} ${g.role}`.toLowerCase();
    return hay.includes(q);
  });
}

export function grantTarget(filesBase: string, path: string): string {
  const base = (filesBase || "").replace(/\/$/, "");
  const p = path.startsWith("/") ? path : `/${path}`;
  return `${base}${p}`;
}

export function roleLabel(role: GrantRole, t: (key: string) => string): string {
  return role === "editor" ? t("grants.editor") : t("grants.viewer");
}

export function filterAdminGrants(grants: Grant[], query: string): Grant[] {
  const q = (query || "").trim().toLowerCase();
  if (!q) return grants;
  return grants.filter((g) => {
    const hay =
      `${g.path} ${g.ownerUsername ?? ""} ${g.granteeUsername ?? ""}`.toLowerCase();
    return hay.includes(q);
  });
}
