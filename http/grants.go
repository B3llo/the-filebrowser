package fbhttp

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	fberrors "github.com/B3llo/the-filebrowser/errors"
	"github.com/B3llo/the-filebrowser/grants"
	"github.com/B3llo/the-filebrowser/users"
)

// grantJSON is a grant enriched with usernames for display. The frontend
// cannot resolve usernames itself for non-admins (GET /api/users is
// admin-only), so the backend includes them best-effort.
type grantJSON struct {
	*grants.Grant
	OwnerUsername   string `json:"ownerUsername,omitempty"`
	GranteeUsername string `json:"granteeUsername,omitempty"`
}

func enrichGrants(d *data, list []*grants.Grant) []grantJSON {
	out := make([]grantJSON, 0, len(list))
	if len(list) == 0 {
		return out
	}

	names := map[uint]string{}
	if all, err := d.store.Users.Gets(d.server.Root, d.server.FollowExternalSymlinks); err == nil {
		for _, u := range all {
			names[u.ID] = u.Username
		}
	}

	for _, g := range list {
		out = append(out, grantJSON{
			Grant:           g,
			OwnerUsername:   names[g.OwnerID],
			GranteeUsername: names[g.GranteeID],
		})
	}
	return out
}

// resolveUserRef resolves a grantee/owner reference that may be a numeric
// ID or a username.
func resolveUserRef(d *data, ref string) (*users.User, error) {
	if ref == "" {
		return nil, fberrors.ErrInvalidRequestParams
	}
	if id, err := strconv.ParseUint(ref, 10, 0); err == nil {
		return d.store.Users.Get(d.server.Root, d.server.FollowExternalSymlinks, uint(id))
	}
	return d.store.Users.Get(d.server.Root, d.server.FollowExternalSymlinks, ref)
}

func getGrantID(r *http.Request) (uint, error) {
	vars := mux.Vars(r)
	i, err := strconv.ParseUint(vars["id"], 10, 0)
	if err != nil {
		return 0, err
	}
	return uint(i), nil
}

var grantsListHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	var list []*grants.Grant
	if d.user.Perm.Admin {
		all, err := d.store.Grants.All()
		if errors.Is(err, fberrors.ErrNotExist) {
			return renderJSON(w, r, []grantJSON{})
		}
		if err != nil {
			return http.StatusInternalServerError, err
		}
		list = all
	} else {
		owned, err := d.store.Grants.FindByOwner(d.user.ID)
		if err != nil && !errors.Is(err, fberrors.ErrNotExist) {
			return http.StatusInternalServerError, err
		}
		received, err := d.store.Grants.FindByGrantee(d.user.ID)
		if err != nil && !errors.Is(err, fberrors.ErrNotExist) {
			return http.StatusInternalServerError, err
		}
		seen := map[uint]struct{}{}
		for _, g := range append(owned, received...) {
			if _, ok := seen[g.ID]; ok {
				continue
			}
			seen[g.ID] = struct{}{}
			list = append(list, g)
		}
	}

	return renderJSON(w, r, enrichGrants(d, list))
})

var grantsSharedWithMeHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	list, err := d.store.Grants.FindActive(d.user.ID)
	if errors.Is(err, fberrors.ErrNotExist) {
		return renderJSON(w, r, []grantJSON{})
	}
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, enrichGrants(d, list))
})

var grantsPostHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if r.Body == nil {
		return http.StatusBadRequest, fberrors.ErrEmptyRequest
	}

	var body grants.CreateBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return http.StatusBadRequest, err
	}

	// The owner defaults to the requester. Only admins may create grants
	// on another user's scope.
	owner := d.user
	if body.Owner != "" {
		if !d.user.Perm.Admin {
			return http.StatusForbidden, nil
		}
		resolved, err := resolveUserRef(d, body.Owner)
		if err != nil {
			return errToStatus(err), err
		}
		owner = resolved
	} else if !d.user.Perm.Admin && !d.user.Perm.Share {
		return http.StatusForbidden, nil
	}

	grantee, err := resolveUserRef(d, body.Grantee)
	if err != nil {
		return errToStatus(err), err
	}

	cleanPath, err := grants.NormalizePath(body.Path)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// Only allow granting paths that currently exist in the owner's scope.
	// owner.Fs is scoped, so Stat also refuses to follow a symlink whose
	// target escapes the scope — same guarantee as sharePostHandler.
	if _, err := owner.Fs.Stat(cleanPath); err != nil {
		return errToStatus(err), err
	}

	expire, err := grants.ExpireFromStrings(body.Expires, body.Unit)
	if err != nil {
		return http.StatusBadRequest, err
	}

	g := &grants.Grant{
		Path:      cleanPath,
		OwnerID:   owner.ID,
		GranteeID: grantee.ID,
		Role:      body.Role,
		CreatedBy: d.user.ID,
		CreatedAt: time.Now().Unix(),
		Expire:    expire,
	}
	if err := g.Validate(); err != nil {
		return http.StatusBadRequest, err
	}

	if existing, err := d.store.Grants.FindByOwner(owner.ID); err == nil {
		for _, e := range existing {
			if e.Path == g.Path && e.GranteeID == g.GranteeID {
				return http.StatusConflict, fberrors.ErrExist
			}
		}
	} else if !errors.Is(err, fberrors.ErrNotExist) {
		return http.StatusInternalServerError, err
	}

	if err := d.store.Grants.Save(g); err != nil {
		return errToStatus(err), err
	}

	log.Printf("grant: user %d granted %s on %s (owner %d) role %s", d.user.ID, grantee.Username, cleanPath, owner.ID, g.Role)
	return renderJSON(w, r, g)
})

var grantsDeleteHandler = withUser(func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
	id, err := getGrantID(r)
	if err != nil {
		return http.StatusBadRequest, nil
	}

	g, err := d.store.Grants.Get(id)
	if err != nil {
		return errToStatus(err), err
	}

	if g.OwnerID != d.user.ID && !d.user.Perm.Admin {
		return http.StatusForbidden, nil
	}

	if err := d.store.Grants.Delete(id); err != nil {
		return errToStatus(err), err
	}

	log.Printf("grant: user %d revoked grant %d (owner %d grantee %d path %s)", d.user.ID, id, g.OwnerID, g.GranteeID, g.Path)
	return http.StatusOK, nil
})

type grantPatchBody struct {
	Role    *grants.Role `json:"role"`
	Expires *string      `json:"expires"`
	Unit    string       `json:"unit"`
}

var grantsPatchHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	id, err := getGrantID(r)
	if err != nil {
		return http.StatusBadRequest, nil
	}

	g, err := d.store.Grants.Get(id)
	if err != nil {
		return errToStatus(err), err
	}

	if g.OwnerID != d.user.ID && !d.user.Perm.Admin {
		return http.StatusForbidden, nil
	}

	if r.Body == nil {
		return http.StatusBadRequest, fberrors.ErrEmptyRequest
	}
	var body grantPatchBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return http.StatusBadRequest, err
	}

	if body.Role != nil {
		g.Role = *body.Role
	}
	if body.Expires != nil {
		expire, err := grants.ExpireFromStrings(*body.Expires, body.Unit)
		if err != nil {
			return http.StatusBadRequest, err
		}
		g.Expire = expire
	}

	if err := g.Validate(); err != nil {
		return http.StatusBadRequest, err
	}
	if err := d.store.Grants.Save(g); err != nil {
		return errToStatus(err), err
	}

	log.Printf("grant: user %d patched grant %d (role %s)", d.user.ID, id, g.Role)
	return renderJSON(w, r, g)
})
