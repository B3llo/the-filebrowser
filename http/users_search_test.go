package fbhttp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/asdine/storm/v3"

	"github.com/B3llo/the-filebrowser/settings"
	"github.com/B3llo/the-filebrowser/storage/bolt"
	"github.com/B3llo/the-filebrowser/users"
)

func TestUsersSearchHandler(t *testing.T) {
	t.Parallel()

	key := []byte("search-test-key")
	db, err := storm.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	st, err := bolt.NewStorage(db)
	if err != nil {
		t.Fatalf("failed to get storage: %v", err)
	}
	// alice=1 (requester), bob=2, bobby=3 + filler users for the limit check.
	names := []string{"alice", "bob", "bobby"}
	for i := 0; i < 12; i++ {
		names = append(names, fmt.Sprintf("zxuser%02d", i))
	}
	perm := users.Permissions{Download: true, Share: true}
	for _, n := range names {
		if err := st.Users.Save(&users.User{Username: n, Password: "pw", Perm: perm}); err != nil {
			t.Fatalf("failed to save %s: %v", n, err)
		}
	}
	if err := st.Settings.Save(&settings.Settings{Key: key}); err != nil {
		t.Fatalf("failed to save settings: %v", err)
	}

	aliceTok := signGrantToken(t, 1, "alice", perm, key)
	doSearch := func(q, token string) *httptest.ResponseRecorder {
		t.Helper()
		req, _ := http.NewRequest(http.MethodGet, "/api/users/search?q="+q, http.NoBody)
		req.Header.Set("X-Auth", token)
		rec := httptest.NewRecorder()
		handle(usersSearchHandler, "", st, &settings.Server{}).ServeHTTP(rec, req)
		return rec
	}

	t.Run("finds by substring", func(t *testing.T) {
		rec := doSearch("bo", aliceTok)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 body=%q", rec.Code, rec.Body.String())
		}
		var out []map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if len(out) != 2 {
			t.Fatalf("results = %v, want bob + bobby", out)
		}
	})

	t.Run("self excluded", func(t *testing.T) {
		rec := doSearch("ali", aliceTok)
		var out []map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if len(out) != 0 {
			t.Fatalf("results = %v, want empty (self excluded)", out)
		}
	})

	t.Run("short query rejected", func(t *testing.T) {
		if rec := doSearch("a", aliceTok); rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
		if rec := doSearch("", aliceTok); rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("no sensitive fields leak", func(t *testing.T) {
		rec := doSearch("bo", aliceTok)
		var out []map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		for _, item := range out {
			for k := range item {
				if k != "id" && k != "username" && k != "displayName" {
					t.Fatalf("leaked field %q in %v", k, item)
				}
			}
		}
	})

	t.Run("results capped at 10", func(t *testing.T) {
		rec := doSearch("zx", aliceTok)
		var out []map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if len(out) != 10 {
			t.Fatalf("results = %d, want cap of 10", len(out))
		}
	})
}
