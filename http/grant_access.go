package fbhttp

import (
	"net/http"
	"path"
	"strings"

	"github.com/B3llo/the-filebrowser/files"
	"github.com/B3llo/the-filebrowser/grants"
	"github.com/B3llo/the-filebrowser/users"
)

// Grant confinement mirrors the public-share logic in http/public.go: the
// filesystem is rebased onto the grant root with symlink confinement, the
// owner's rules keep applying via checkerPrefix, and access dies as soon as
// the owner loses Share/Download permission.

// cleanReqPath normalizes an owner-style request path.
func cleanReqPath(p string) string {
	if p == "" || p[0] != '/' {
		p = "/" + p
	}
	return path.Clean(p)
}

// matchGrant returns the longest active grant covering clean (boundary-aware)
// whose owner still has Share+Download permission. Pure: it never mutates d.
func matchGrant(clean string, d *data) (*grants.Grant, *users.User) {
	list, err := d.store.Grants.FindActive(d.user.ID)
	if err != nil || len(list) == 0 {
		return nil, nil
	}

	var best *grants.Grant
	for _, cand := range list {
		if cand.Path != clean && !strings.HasPrefix(clean, cand.Path+"/") {
			continue
		}
		if best == nil || len(cand.Path) > len(best.Path) {
			best = cand
		}
	}
	if best == nil {
		return nil, nil
	}

	owner, err := d.store.Users.Get(d.server.Root, d.server.FollowExternalSymlinks, best.OwnerID)
	if err != nil {
		return nil, nil
	}
	if !owner.Perm.Share || !owner.Perm.Download {
		return nil, nil
	}

	return best, owner
}

// applyGrant rebases the requester's filesystem onto the grant root and
// returns the grant-relative path for all downstream filesystem operations.
// The owner is recorded so its rules keep applying (see data.Check).
func applyGrant(g *grants.Grant, owner *users.User, clean string, d *data) string {
	d.user.Fs = files.NewFs(owner.Fs, g.Path, d.server.FollowExternalSymlinks)
	d.checkerPrefix = g.Path
	d.grantOwner = owner

	rest := strings.TrimPrefix(clean, g.Path)
	if rest == "" {
		return "/"
	}
	return "/" + strings.TrimLeft(rest, "/")
}

// tryGrantScope resolves an owner-style request path through active grants.
//
//   - forCreate=false (reads, overwrites, deletes): the full path must exist
//     in the owner's scope.
//   - forCreate=true (new uploads): the parent directory may exist instead,
//     so editors can create new files inside a granted folder.
//
// Own-scope wins whenever it can serve the request: an existing path, or
// (for creates) an existing parent directory, always resolves locally and
// returns active=false. Returns the grant-relative path on a match.
func tryGrantScope(reqPath string, d *data, forCreate bool) (rel string, g *grants.Grant, active bool) {
	clean := cleanReqPath(reqPath)

	if _, err := d.user.Fs.Stat(clean); err == nil {
		return "", nil, false
	}
	if forCreate {
		if _, err := d.user.Fs.Stat(path.Dir(clean)); err == nil {
			return "", nil, false
		}
	}

	g, owner := matchGrant(clean, d)
	if g == nil {
		return "", nil, false
	}

	// owner.Fs is scoped, so Stat refuses to follow symlinks escaping the
	// owner's scope — same guarantee as sharePostHandler.
	if _, err := owner.Fs.Stat(clean); err == nil {
		return applyGrant(g, owner, clean, d), g, true
	}
	if forCreate {
		if _, err := owner.Fs.Stat(path.Dir(clean)); err == nil {
			return applyGrant(g, owner, clean, d), g, true
		}
	}

	return "", nil, false
}

// grantWriteStatus enforces the role gate for mutating handlers. The
// grantee's global permissions (Create/Modify/Delete/...) keep applying
// through each handler's own checks; the grant only adds the viewer/editor
// ceiling on top.
func grantWriteStatus(g *grants.Grant) (int, bool) {
	if g.Role != grants.RoleEditor {
		return http.StatusForbidden, false
	}
	return 0, true
}

// grantOwnerPath maps a grant-relative path back to owner coordinates.
// Used for cascade cleanup, which is tracked in the owner's scope.
func grantOwnerPath(g *grants.Grant, rel string) string {
	if rel == "/" {
		return g.Path
	}
	return strings.TrimRight(g.Path, "/") + "/" + strings.TrimLeft(rel, "/")
}
