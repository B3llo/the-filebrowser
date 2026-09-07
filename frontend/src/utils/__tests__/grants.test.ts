import { describe, it, expect } from "vitest";
import {
  filterAdminGrants,
  filterGrants,
  grantDisplayName,
  grantTarget,
  roleLabel,
} from "@/utils/grants";

const grants = [
  {
    id: 1,
    path: "/docs/report.txt",
    ownerID: 1,
    granteeID: 2,
    role: "viewer",
    ownerUsername: "alice",
  },
  {
    id: 2,
    path: "/media",
    ownerID: 3,
    granteeID: 2,
    role: "editor",
    ownerUsername: "carol",
  },
] as Grant[];

describe("utils/grants", () => {
  it("grantDisplayName returns the basename", () => {
    expect(grantDisplayName("/docs/report.txt")).toBe("report.txt");
    expect(grantDisplayName("/media")).toBe("media");
    expect(grantDisplayName("/")).toBe("/");
  });

  it("filterGrants matches path, owner and role", () => {
    expect(filterGrants(grants, "")).toHaveLength(2);
    expect(filterGrants(grants, "report")).toEqual([grants[0]]);
    expect(filterGrants(grants, "CAROL")).toEqual([grants[1]]);
    expect(filterGrants(grants, "editor")).toEqual([grants[1]]);
    expect(filterGrants(grants, "nobody")).toEqual([]);
  });

  it("grantTarget joins filesBase and path", () => {
    expect(grantTarget("/files/0", "/docs")).toBe("/files/0/docs");
    expect(grantTarget("/files/0/", "/docs")).toBe("/files/0/docs");
  });

  it("roleLabel falls back to viewer", () => {
    const t = (k: string) => k;
    expect(roleLabel("editor", t)).toBe("grants.editor");
    expect(roleLabel("viewer", t)).toBe("grants.viewer");
    expect(roleLabel("anything" as GrantRole, t)).toBe("grants.viewer");
  });

  it("filterAdminGrants matches path and usernames", () => {
    const all = [
      { ...grants[0], granteeUsername: "bob" },
      { ...grants[1], granteeUsername: "dave" },
    ] as Grant[];
    expect(filterAdminGrants(all, "")).toHaveLength(2);
    expect(filterAdminGrants(all, "bob")).toEqual([all[0]]);
    expect(filterAdminGrants(all, "CAROL")).toEqual([all[1]]);
    expect(filterAdminGrants(all, "/media")).toEqual([all[1]]);
    expect(filterAdminGrants(all, "nobody")).toEqual([]);
  });
});
