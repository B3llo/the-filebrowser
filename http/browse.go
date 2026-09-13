package fbhttp

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/B3llo/the-filebrowser/sources"
)

// browseHandler lists immediate subdirectories at the given host path.
// Admin-only: admins can already type arbitrary paths when creating sources,
// so this adds discoverability, not new capability.
var browseHandler = withAdmin(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
	rawPath := strings.TrimSpace(r.URL.Query().Get("path"))
	if rawPath == "" {
		rawPath = "/"
	}
	absPath, err := sources.NormalizePath(rawPath)
	if err != nil {
		// Keep browse forgiving: invalid/pasted paths just list nothing so the
		// picker stays usable instead of surfacing a 400.
		return renderJSON(w, r, []browseDirEntry{})
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		// Return empty list instead of 500 for permission errors / missing paths
		return renderJSON(w, r, []browseDirEntry{})
	}

	result := make([]browseDirEntry, 0)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		childPath := filepath.Join(absPath, e.Name())
		hasSub := dirHasSubdirs(childPath)
		result = append(result, browseDirEntry{
			Name:        e.Name(),
			Path:        childPath,
			HasChildren: hasSub,
		})
	}

	return renderJSON(w, r, result)
})

type browseDirEntry struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	HasChildren bool   `json:"hasChildren"`
}

func dirHasSubdirs(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() {
			return true
		}
	}
	return false
}
