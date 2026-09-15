package fbhttp

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	gopath "path"
	"strings"

	libErrors "github.com/B3llo/the-filebrowser/errors"
	imgErrors "github.com/B3llo/the-filebrowser/img"
)

// cleanUserPath normalizes a request path so root guards cannot be bypassed
// with "/.", "/..", "//" or "/sub/..". It always returns a path starting
// with "/". Callers must compare the result against "/" — never the raw
// r.URL.Path — before destructive operations.
func cleanUserPath(p string) string {
	if p == "" {
		return "/"
	}
	cleaned := gopath.Clean("/" + strings.TrimPrefix(p, "/"))
	if !strings.HasPrefix(cleaned, "/") {
		return "/" + cleaned
	}
	return cleaned
}

func renderJSON(w http.ResponseWriter, _ *http.Request, data interface{}) (int, error) {
	marsh, err := json.Marshal(data)

	if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := w.Write(marsh); err != nil {
		return http.StatusInternalServerError, err
	}

	return 0, nil
}

func errToStatus(err error) int {
	switch {
	case err == nil:
		return http.StatusOK
	case os.IsPermission(err):
		return http.StatusForbidden
	case os.IsNotExist(err), errors.Is(err, libErrors.ErrNotExist):
		return http.StatusNotFound
	case os.IsExist(err), errors.Is(err, libErrors.ErrExist):
		return http.StatusConflict
	case errors.Is(err, libErrors.ErrPermissionDenied):
		return http.StatusForbidden
	case errors.Is(err, libErrors.ErrInvalidRequestParams):
		return http.StatusBadRequest
	case errors.Is(err, libErrors.ErrRootUserDeletion):
		return http.StatusForbidden
	case errors.Is(err, imgErrors.ErrImageTooLarge):
		return http.StatusRequestEntityTooLarge
	default:
		return http.StatusInternalServerError
	}
}

// This is an adaptation if http.StripPrefix in which we don't
// return 404 if the page doesn't have the needed prefix.
func stripPrefix(prefix string, h http.Handler) http.Handler {
	if prefix == "" || prefix == "/" {
		return h
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, prefix)
		rp := strings.TrimPrefix(r.URL.RawPath, prefix)

		// If the path is exactly the prefix (no trailing slash), redirect to
		// the prefix with a trailing slash so the router receives "/" instead
		// of "", which would otherwise cause a redirect to the site root.
		if p == "" {
			http.Redirect(w, r, prefix+"/", http.StatusMovedPermanently)
			return
		}

		r2 := new(http.Request)
		*r2 = *r
		r2.URL = new(url.URL)
		*r2.URL = *r.URL
		r2.URL.Path = p
		r2.URL.RawPath = rp
		h.ServeHTTP(w, r2)
	})
}
