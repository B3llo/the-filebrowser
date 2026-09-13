package fbhttp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/asdine/storm/v3"
	"github.com/gorilla/mux"

	"github.com/B3llo/the-filebrowser/settings"
	"github.com/B3llo/the-filebrowser/sources"
	"github.com/B3llo/the-filebrowser/storage"
	"github.com/B3llo/the-filebrowser/storage/bolt"
	"github.com/B3llo/the-filebrowser/users"
)

type sourceFixture struct {
	store *storage.Storage
	key   []byte
	admin users.Permissions
}

func newSourceFixture(t *testing.T) *sourceFixture {
	t.Helper()

	key := []byte("source-test-key")
	db, err := storm.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	st, err := bolt.NewStorage(db)
	if err != nil {
		t.Fatalf("failed to get storage: %v", err)
	}

	fx := &sourceFixture{
		store: st,
		key:   key,
		admin: users.Permissions{Admin: true, Share: true, Download: true, Modify: true, Delete: true, Create: true},
	}
	if err := st.Users.Save(&users.User{Username: "admin", Password: "pw", Perm: fx.admin}); err != nil {
		t.Fatalf("failed to save admin: %v", err)
	}
	if err := st.Settings.Save(&settings.Settings{Key: key}); err != nil {
		t.Fatalf("failed to save settings: %v", err)
	}
	return fx
}

func doSourceRequest(t *testing.T, fx *sourceFixture, handler handleFunc, method, target, body, token string, vars map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	req, err := http.NewRequest(method, target, strings.NewReader(body))
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	req.Header.Set("X-Auth", token)
	if vars != nil {
		req = mux.SetURLVars(req, vars)
	}
	rec := httptest.NewRecorder()
	handle(handler, "", fx.store, &settings.Server{}).ServeHTTP(rec, req)
	return rec
}

func TestSourcePostAcceptsPastedPaths(t *testing.T) {
	t.Parallel()

	fx := newSourceFixture(t)
	adminTok := signGrantToken(t, 1, "admin", fx.admin, fx.key)

	// Pasted values as they arrive from the UI: leading/trailing whitespace,
	// newlines and surrounding quotes must not trigger a 400.
	rawBase := filepath.Join(t.TempDir(), "media")
	pasted := []string{
		"  " + rawBase + "  \n",
		"\"" + rawBase + "\"",
		"'" + rawBase + "'",
		rawBase,
	}
	for i, p := range pasted {
		name := "src-paste-" + strings.Repeat("x", i+1)
		body, _ := json.Marshal(map[string]any{
			"what": "source",
			"data": map[string]any{"name": name, "path": p},
		})
		rec := doSourceRequest(t, fx, sourcePostHandler, http.MethodPost, "/api/sources", string(body), adminTok, nil)
		if rec.Code != http.StatusCreated {
			t.Fatalf("POST pasted path %q: got status %d body %q", p, rec.Code, rec.Body.String())
		}
		var got sources.Source
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("POST pasted path %q: invalid json: %v", p, err)
		}
		if got.Path != rawBase {
			t.Errorf("POST pasted path %q: stored path = %q, want %q", p, got.Path, rawBase)
		}
		if st, err := os.Stat(rawBase); err != nil || !st.IsDir() {
			t.Errorf("POST pasted path %q: source dir was not created: %v", p, err)
		}
	}
}

func TestSourcePostRejectsBadPaths(t *testing.T) {
	t.Parallel()

	fx := newSourceFixture(t)
	adminTok := signGrantToken(t, 1, "admin", fx.admin, fx.key)

	// A regular file is not a valid source root.
	filePath := filepath.Join(t.TempDir(), "afile")
	if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name     string
		path     string
		wantCode int
	}{
		{"relative", "data/relative", http.StatusBadRequest},
		{"notadir", filePath, http.StatusBadRequest},
	}
	for _, tc := range cases {
		body, _ := json.Marshal(map[string]any{
			"what": "source",
			"data": map[string]any{"name": "bad-" + tc.name, "path": tc.path},
		})
		rec := doSourceRequest(t, fx, sourcePostHandler, http.MethodPost, "/api/sources", string(body), adminTok, nil)
		if rec.Code != tc.wantCode {
			t.Errorf("%s: got status %d, want %d (body %q)", tc.name, rec.Code, tc.wantCode, rec.Body.String())
		}
	}
}

func TestSourcePutNormalizesPath(t *testing.T) {
	t.Parallel()

	fx := newSourceFixture(t)
	adminTok := signGrantToken(t, 1, "admin", fx.admin, fx.key)

	dir := filepath.Join(t.TempDir(), "vids")
	body, _ := json.Marshal(map[string]any{
		"what": "source",
		"data": map[string]any{"name": "vids", "path": dir},
	})
	rec := doSourceRequest(t, fx, sourcePostHandler, http.MethodPost, "/api/sources", string(body), adminTok, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup POST: got %d body %q", rec.Code, rec.Body.String())
	}
	var created sources.Source
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("setup decode: %v", err)
	}

	// Update with a pasted path (leading space + quotes): must normalize, not 400.
	newDir := filepath.Join(t.TempDir(), "movies")
	upd, _ := json.Marshal(map[string]any{
		"what":  "source",
		"which": []string{"Name", "Path"},
		"data":  map[string]any{"id": created.ID, "name": "vids", "path": "  \"" + newDir + "\"  "},
	})
	vars := map[string]string{"id": strconv.FormatUint(uint64(created.ID), 10)}
	rec2 := doSourceRequest(t, fx, sourcePutHandler, http.MethodPut, "/api/sources/"+vars["id"], string(upd), adminTok, vars)
	if rec2.Code != http.StatusOK {
		t.Fatalf("PUT pasted path: got status %d body %q", rec2.Code, rec2.Body.String())
	}
	var updated sources.Source
	if err := json.Unmarshal(rec2.Body.Bytes(), &updated); err != nil {
		t.Fatalf("PUT decode: %v", err)
	}
	if updated.Path != newDir {
		t.Errorf("PUT pasted path: stored = %q, want %q", updated.Path, newDir)
	}
}
