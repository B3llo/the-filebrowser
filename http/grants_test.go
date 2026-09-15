package fbhttp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/spf13/afero"

	"github.com/B3llo/the-filebrowser/grants"
	"github.com/B3llo/the-filebrowser/settings"
	"github.com/B3llo/the-filebrowser/storage"
	"github.com/B3llo/the-filebrowser/storage/bolt"
	"github.com/B3llo/the-filebrowser/users"
)

type grantFixture struct {
	store *storage.Storage
	key   []byte
	admin users.Permissions
	alice users.Permissions
	bob   users.Permissions
	mall  users.Permissions
}

func signGrantToken(t *testing.T, id uint, username string, perm users.Permissions, key []byte) string {
	t.Helper()
	claims := &authToken{
		User: userInfo{ID: id, Username: username, Perm: perm},
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Minute)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(key)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return signed
}

func newGrantFixture(t *testing.T) *grantFixture {
	t.Helper()

	key := []byte("grant-test-key")
	db, err := storm.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	st, err := bolt.NewStorage(db)
	if err != nil {
		t.Fatalf("failed to get storage: %v", err)
	}

	fx := &grantFixture{
		store: st,
		key:   key,
		admin: users.Permissions{Admin: true, Share: true, Download: true, Modify: true, Delete: true, Create: true},
		alice: users.Permissions{Share: true, Download: true, Modify: true, Create: true},
		bob:   users.Permissions{Download: true},
		mall:  users.Permissions{Download: true},
	}

	// IDs are assigned incrementally: admin=1, alice=2, bob=3, mallory=4.
	for _, u := range []struct {
		name string
		perm users.Permissions
	}{
		{"admin", fx.admin}, {"alice", fx.alice}, {"bob", fx.bob}, {"mallory", fx.mall},
	} {
		if err := st.Users.Save(&users.User{Username: u.name, Password: "pw", Perm: u.perm}); err != nil {
			t.Fatalf("failed to save user %s: %v", u.name, err)
		}
	}
	if err := st.Settings.Save(&settings.Settings{Key: key}); err != nil {
		t.Fatalf("failed to save settings: %v", err)
	}

	// One shared memfs for all users is enough for handler CRUD tests:
	// ownership is tracked by IDs, isolation is covered by grant_access tests.
	mem := afero.NewMemMapFs()
	if err := mem.MkdirAll("/docs", 0o755); err != nil {
		t.Fatalf("failed to mkdir: %v", err)
	}
	if err := afero.WriteFile(mem, "/docs/report.txt", []byte("hello"), 0o644); err != nil {
		t.Fatalf("failed to seed file: %v", err)
	}
	st.Users = &customFSUser{Store: st.Users, fs: mem}

	return fx
}

func doGrantRequest(t *testing.T, fx *grantFixture, handler handleFunc, method, target, body, token string, vars map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, target, reader)
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

func TestGrantsPostHandler(t *testing.T) {
	t.Parallel()

	fx := newGrantFixture(t)
	aliceTok := signGrantToken(t, 2, "alice", fx.alice, fx.key)
	bobTok := signGrantToken(t, 3, "bob", fx.bob, fx.key)
	adminTok := signGrantToken(t, 1, "admin", fx.admin, fx.key)

	// Numeric username: stored as username "123" with a real auto ID (5).
	// Regression test for POST /api/grants 404ing on such users because the
	// ref was parsed as a user ID instead of a username.
	if err := fx.store.Users.Save(&users.User{Username: "123", Password: "pw", Perm: fx.bob}); err != nil {
		t.Fatalf("failed to save numeric user: %v", err)
	}

	cases := map[string]struct {
		token      string
		body       string
		wantStatus int
	}{
		"nonexistent path is 404": {
			token:      aliceTok,
			body:       `{"path":"/nope","grantee":"bob","role":"viewer"}`,
			wantStatus: http.StatusNotFound,
		},
		"user without share perm is 403": {
			token:      bobTok,
			body:       `{"path":"/docs","grantee":"mallory","role":"viewer"}`,
			wantStatus: http.StatusForbidden,
		},
		"self grant is 400": {
			token:      aliceTok,
			body:       `{"path":"/docs","grantee":"alice","role":"viewer"}`,
			wantStatus: http.StatusBadRequest,
		},
		"bad role is 400": {
			token:      aliceTok,
			body:       `{"path":"/docs","grantee":"mallory","role":"owner"}`,
			wantStatus: http.StatusBadRequest,
		},
		"unknown grantee is 404": {
			token:      aliceTok,
			body:       `{"path":"/docs","grantee":"ghost","role":"viewer"}`,
			wantStatus: http.StatusNotFound,
		},
		"non-admin with owner field is 403": {
			token:      aliceTok,
			body:       `{"path":"/docs","owner":"bob","grantee":"mallory","role":"viewer"}`,
			wantStatus: http.StatusForbidden,
		},
		"admin creates on owner scope": {
			token:      adminTok,
			body:       `{"path":"/docs","owner":"alice","grantee":"mallory","role":"editor"}`,
			wantStatus: http.StatusOK,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rec := doGrantRequest(t, fx, grantsPostHandler, http.MethodPost, "/api/grants", tc.body, tc.token, nil)
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d (body=%q)", rec.Code, tc.wantStatus, rec.Body.String())
			}
		})
	}

	// Ordered: create first (200), then the same grant again (409).
	// Kept out of the map loop above because map iteration order is random.
	t.Run("owner creates viewer grant", func(t *testing.T) {
		rec := doGrantRequest(t, fx, grantsPostHandler, http.MethodPost, "/api/grants",
			`{"path":"/docs","grantee":"bob","role":"viewer"}`, aliceTok, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%q)", rec.Code, rec.Body.String())
		}
	})
	t.Run("duplicate is conflict", func(t *testing.T) {
		rec := doGrantRequest(t, fx, grantsPostHandler, http.MethodPost, "/api/grants",
			`{"path":"/docs","grantee":"bob","role":"viewer"}`, aliceTok, nil)
		if rec.Code != http.StatusConflict {
			t.Fatalf("status = %d, want 409 (body=%q)", rec.Code, rec.Body.String())
		}
	})

	// Numeric usernames resolve by username when no user has that ID.
	t.Run("numeric username grant resolves to username", func(t *testing.T) {
		rec := doGrantRequest(t, fx, grantsPostHandler, http.MethodPost, "/api/grants",
			`{"path":"/docs","grantee":"123","role":"viewer"}`, aliceTok, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%q)", rec.Code, rec.Body.String())
		}
		var created grants.Grant
		if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
			t.Fatalf("failed to decode grant: %v", err)
		}
		if created.GranteeID != 5 {
			t.Fatalf("granteeID = %d, want 5 (user \"123\")", created.GranteeID)
		}
	})

	// The admin-created grant must belong to alice, not to the admin.
	list := doGrantRequest(t, fx, grantsListHandler, http.MethodGet, "/api/grants", "", adminTok, nil)
	var all []grants.Grant
	if err := json.Unmarshal(list.Body.Bytes(), &all); err != nil {
		t.Fatalf("failed to decode list: %v", err)
	}
	found := false
	for _, g := range all {
		if g.GranteeID == 4 && g.Role == grants.RoleEditor {
			found = true
			if g.OwnerID != 2 {
				t.Fatalf("admin grant owner = %d, want alice (2)", g.OwnerID)
			}
		}
	}
	if !found {
		t.Fatalf("admin-created editor grant for mallory not found in %v", all)
	}
}

func TestGrantsListAndSharedWithMe(t *testing.T) {
	t.Parallel()

	fx := newGrantFixture(t)
	mustSave := func(g *grants.Grant) {
		t.Helper()
		if err := fx.store.Grants.Save(g); err != nil {
			t.Fatalf("failed to save grant: %v", err)
		}
	}
	// g1: alice -> bob, g2: alice -> mallory, g3: bob -> alice.
	mustSave(&grants.Grant{Path: "/docs", OwnerID: 2, GranteeID: 3, Role: grants.RoleViewer})
	mustSave(&grants.Grant{Path: "/docs", OwnerID: 2, GranteeID: 4, Role: grants.RoleEditor})
	mustSave(&grants.Grant{Path: "/docs", OwnerID: 3, GranteeID: 2, Role: grants.RoleViewer})

	bobTok := signGrantToken(t, 3, "bob", fx.bob, fx.key)
	adminTok := signGrantToken(t, 1, "admin", fx.admin, fx.key)
	mallTok := signGrantToken(t, 4, "mallory", fx.mall, fx.key)

	decode := func(rec *httptest.ResponseRecorder) []grants.Grant {
		t.Helper()
		var out []grants.Grant
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("decode failed: %v body=%q", err, rec.Body.String())
		}
		return out
	}

	// Bob sees g1 (received) + g3 (owned), never g2.
	if got := decode(doGrantRequest(t, fx, grantsListHandler, http.MethodGet, "/api/grants", "", bobTok, nil)); len(got) != 2 {
		t.Fatalf("bob list = %d grants, want 2", len(got))
	}
	// Admin sees everything.
	if got := decode(doGrantRequest(t, fx, grantsListHandler, http.MethodGet, "/api/grants", "", adminTok, nil)); len(got) != 3 {
		t.Fatalf("admin list = %d grants, want 3", len(got))
	}
	// shared-with-me is received-only.
	if got := decode(doGrantRequest(t, fx, grantsSharedWithMeHandler, http.MethodGet, "/api/grants/shared-with-me", "", bobTok, nil)); len(got) != 1 || got[0].OwnerID != 2 {
		t.Fatalf("bob shared-with-me = %v, want 1 grant from alice", got)
	}
	if got := decode(doGrantRequest(t, fx, grantsSharedWithMeHandler, http.MethodGet, "/api/grants/shared-with-me", "", mallTok, nil)); len(got) != 1 {
		t.Fatalf("mallory shared-with-me = %d grants, want 1", len(got))
	}
}

func TestGrantsDeleteAndPatch(t *testing.T) {
	t.Parallel()

	fx := newGrantFixture(t)
	g := &grants.Grant{Path: "/docs", OwnerID: 2, GranteeID: 3, Role: grants.RoleViewer}
	if err := fx.store.Grants.Save(g); err != nil {
		t.Fatalf("failed to save grant: %v", err)
	}
	vars := map[string]string{"id": "1"}

	aliceTok := signGrantToken(t, 2, "alice", fx.alice, fx.key)
	bobTok := signGrantToken(t, 3, "bob", fx.bob, fx.key)
	mallTok := signGrantToken(t, 4, "mallory", fx.mall, fx.key)
	adminTok := signGrantToken(t, 1, "admin", fx.admin, fx.key)

	// Stranger cannot patch or delete.
	if rec := doGrantRequest(t, fx, grantsPatchHandler, http.MethodPatch, "/api/grants/1", `{"role":"editor"}`, mallTok, vars); rec.Code != http.StatusForbidden {
		t.Fatalf("stranger patch = %d, want 403", rec.Code)
	}
	if rec := doGrantRequest(t, fx, grantsDeleteHandler, http.MethodDelete, "/api/grants/1", "", mallTok, vars); rec.Code != http.StatusForbidden {
		t.Fatalf("stranger delete = %d, want 403", rec.Code)
	}
	// Invalid role is rejected.
	if rec := doGrantRequest(t, fx, grantsPatchHandler, http.MethodPatch, "/api/grants/1", `{"role":"owner"}`, aliceTok, vars); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad role patch = %d, want 400", rec.Code)
	}
	// Owner upgrades viewer -> editor.
	if rec := doGrantRequest(t, fx, grantsPatchHandler, http.MethodPatch, "/api/grants/1", `{"role":"editor"}`, aliceTok, vars); rec.Code != http.StatusOK {
		t.Fatalf("owner patch = %d, want 200 body=%q", rec.Code, rec.Body.String())
	}
	updated, err := fx.store.Grants.Get(1)
	if err != nil {
		t.Fatalf("Get after patch failed: %v", err)
	}
	if updated.Role != grants.RoleEditor {
		t.Fatalf("role after patch = %q, want editor", updated.Role)
	}
	// Grantee cannot delete someone else's grant record (only owner/admin).
	if rec := doGrantRequest(t, fx, grantsDeleteHandler, http.MethodDelete, "/api/grants/1", "", bobTok, vars); rec.Code != http.StatusForbidden {
		t.Fatalf("grantee delete = %d, want 403", rec.Code)
	}
	// Admin can revoke.
	if rec := doGrantRequest(t, fx, grantsDeleteHandler, http.MethodDelete, "/api/grants/1", "", adminTok, vars); rec.Code != http.StatusOK {
		t.Fatalf("admin delete = %d, want 200 body=%q", rec.Code, rec.Body.String())
	}
	// Second delete is 404.
	if rec := doGrantRequest(t, fx, grantsDeleteHandler, http.MethodDelete, "/api/grants/1", "", adminTok, vars); rec.Code != http.StatusNotFound {
		t.Fatalf("double delete = %d, want 404", rec.Code)
	}
}
