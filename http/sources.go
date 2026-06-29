package fbhttp

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/afero"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/sources"
	"github.com/filebrowser/filebrowser/v2/users"
)

const (
	// sourceHeader overrides the active source per request.
	sourceHeader = "X-Source"
	// sourceCookie carries the active source on browser-initiated requests
	// (downloads, previews) that cannot send a custom header.
	sourceCookie = "source"
)

// resolveActiveSource rebinds d.user.Fs to the active source's ScopedFs. The
// active source is read from the X-Source header, falling back to the "source"
	// cookie. When it resolves to the implicit source (id 0 / unknown /
// inaccessible), the filesystem built by User.Clean — the classic user scope —
// is left untouched, preserving legacy behavior.
func resolveActiveSource(r *http.Request, d *data) {
	id := activeSourceID(r)
	if id == sources.ImplicitSourceID || !userCanAccessSource(d.user, id) {
		return
	}

	src, err := d.store.Sources.Get(id)
	if err != nil {
		return
	}

	d.user.Fs = files.NewFs(afero.NewOsFs(), src.Path, d.server.FollowExternalSymlinks)
}

// activeSourceID parses the active source id from the request (header first,
// then cookie). It returns 0 (the implicit source) when absent or unparseable.
func activeSourceID(r *http.Request) uint {
	raw := r.Header.Get(sourceHeader)
	if raw == "" {
		if c, err := r.Cookie(sourceCookie); err == nil {
			raw = c.Value
		}
	}
	if raw == "" {
		return sources.ImplicitSourceID
	}
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return sources.ImplicitSourceID
	}
	return uint(id)
}

// userCanAccessSource reports whether a user may browse the given source. When
// the user has no explicit SourceRef list, every defined source is accessible;
// otherwise only the assigned subset is.
func userCanAccessSource(u *users.User, id uint) bool {
	if len(u.Sources) == 0 {
		return true
	}
	for _, s := range u.Sources {
		if s.ID == id {
			return true
		}
	}
	return false
}

// assignedSourceIDs returns the ids a user may browse, or nil when the user has
// no explicit assignment (meaning "all defined sources").
func assignedSourceIDs(u *users.User) []uint {
	if len(u.Sources) == 0 {
		return nil
	}
	ids := make([]uint, 0, len(u.Sources))
	for _, r := range u.Sources {
		ids = append(ids, r.ID)
	}
	return ids
}

var sourcesListHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	defined, err := d.store.Sources.Gets()
	if err != nil && !errors.Is(err, fberrors.ErrNotExist) {
		return http.StatusInternalServerError, err
	}

	// With no defined sources, expose the legacy scope as a single implicit
	// source so the application behaves exactly as before.
	if len(defined) == 0 {
		return renderJSON(w, r, []sources.Source{{
			ID:   sources.ImplicitSourceID,
			Name: sources.ImplicitName,
		}})
	}

	// Restrict to the user's assigned subset when one is set.
	if ids := assignedSourceIDs(d.user); ids != nil {
		allowed := make(map[uint]bool, len(ids))
		for _, id := range ids {
			allowed[id] = true
		}
		filtered := make([]*sources.Source, 0, len(defined))
		for _, s := range defined {
			if allowed[s.ID] {
				filtered = append(filtered, s)
			}
		}
		defined = filtered
	}

	// Non-admins never see the on-disk path.
	out := make([]sources.Source, 0, len(defined))
	for _, s := range defined {
		entry := sources.Source{ID: s.ID, Name: s.Name, CreatedAt: s.CreatedAt}
		if d.user.Perm.Admin {
			entry.Path = s.Path
		}
		out = append(out, entry)
	}
	return renderJSON(w, r, out)
})

var sourceGetHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	id, err := getSourceID(r)
	if err != nil {
		return http.StatusBadRequest, err
	}
	src, err := d.store.Sources.Get(id)
	if err != nil {
		return errToStatus(err), err
	}
	if !userCanAccessSource(d.user, src.ID) {
		return http.StatusForbidden, nil
	}
	if !d.user.Perm.Admin {
		src.Path = ""
	}
	return renderJSON(w, r, src)
})

type modifySourceRequest struct {
	modifyRequest
	Data *sources.Source `json:"data"`
}

var sourcePostHandler = withAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	req := &modifySourceRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return http.StatusBadRequest, err
	}
	if req.Data == nil || req.Data.Name == "" || req.Data.Path == "" {
		return http.StatusBadRequest, fberrors.ErrInvalidRequestParams
	}

	src := &sources.Source{
		Name:      req.Data.Name,
		Path:      filepath.Clean(req.Data.Path),
		CreatedAt: time.Now(),
	}
	if !filepath.IsAbs(src.Path) {
		return http.StatusBadRequest, errors.New("source path must be absolute")
	}
	if err := os.MkdirAll(src.Path, 0o750); err != nil {
		return http.StatusInternalServerError, err
	}

	if err := d.store.Sources.Save(src); err != nil {
		return errToStatus(err), err
	}
	w.Header().Set("Location", "/api/sources/"+strconv.FormatUint(uint64(src.ID), 10))
	w.WriteHeader(http.StatusCreated)
	return renderJSON(w, r, src)
})

var sourcePutHandler = withAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	id, err := getSourceID(r)
	if err != nil {
		return http.StatusBadRequest, err
	}
	req := &modifySourceRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return http.StatusBadRequest, err
	}
	if req.Data == nil || req.Data.ID == 0 {
		return http.StatusBadRequest, fberrors.ErrInvalidRequestParams
	}

	src, err := d.store.Sources.Get(id)
	if err != nil {
		return errToStatus(err), err
	}

	which := req.Which
	if len(which) == 0 || (len(which) == 1 && which[0] == "all") {
		which = []string{"Name", "Path"}
	}

	for _, field := range which {
		switch field {
		case "Name":
			if req.Data.Name == "" {
				return http.StatusBadRequest, fberrors.ErrInvalidRequestParams
			}
			src.Name = req.Data.Name
		case "Path":
			p := filepath.Clean(req.Data.Path)
			if !filepath.IsAbs(p) {
				return http.StatusBadRequest, errors.New("source path must be absolute")
			}
			src.Path = p
		}
	}

	if err := d.store.Sources.Update(src, which...); err != nil {
		return errToStatus(err), err
	}
	return renderJSON(w, r, src)
})

var sourceDeleteHandler = withAdmin(func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
	id, err := getSourceID(r)
	if err != nil {
		return http.StatusBadRequest, err
	}
	if err := d.store.Sources.Delete(id); err != nil {
		return errToStatus(err), err
	}
	return http.StatusOK, nil
})

func getSourceID(r *http.Request) (uint, error) {
	vars := mux.Vars(r)
	i, err := strconv.ParseUint(vars["id"], 10, 0)
	if err != nil {
		return 0, err
	}
	return uint(i), nil
}
