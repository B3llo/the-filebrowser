package fbhttp

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/asdine/storm/v3"
	"github.com/spf13/afero"

	"github.com/B3llo/the-filebrowser/diskcache"
	"github.com/B3llo/the-filebrowser/files"
	"github.com/B3llo/the-filebrowser/grants"
	"github.com/B3llo/the-filebrowser/rules"
	"github.com/B3llo/the-filebrowser/settings"
	"github.com/B3llo/the-filebrowser/storage"
	"github.com/B3llo/the-filebrowser/storage/bolt"
	"github.com/B3llo/the-filebrowser/users"
)

// perUserFSStore gives every user their own on-disk scope, mirroring
// production isolation (unlike customFSUser, which shares one fs).
type perUserFSStore struct {
	users.Store
	dirs map[uint]string
}

func (p *perUserFSStore) Get(baseScope string, followExternalSymlinks bool, id interface{}) (*users.User, error) {
	u, err := p.Store.Get(baseScope, followExternalSymlinks, id)
	if err != nil {
		return nil, err
	}
	if dir, ok := p.dirs[u.ID]; ok {
		u.Fs = files.NewScopedFs(afero.NewBasePathFs(afero.NewOsFs(), dir), "/")
	}
	return u, nil
}

type accessFixture struct {
	store    *storage.Storage
	key      []byte
	aliceDir string
	bobDir   string
	alice    users.Permissions
	bob      users.Permissions
	mall     users.Permissions
}

func newAccessFixture(t *testing.T, aliceRules []rules.Rule) *accessFixture {
	t.Helper()

	root := t.TempDir()
	fx := &accessFixture{
		key:      []byte("access-test-key"),
		aliceDir: filepath.Join(root, "alice"),
		bobDir:   filepath.Join(root, "bob"),
		alice:    users.Permissions{Share: true, Download: true, Create: true, Modify: true, Rename: true, Delete: true},
		bob:      users.Permissions{Download: true, Create: true, Modify: true, Rename: true},
		mall:     users.Permissions{Download: true},
	}
	for _, dir := range []string{fx.aliceDir, fx.bobDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// Only the owner has /docs: the grantee's scope stays empty so grant
	// fallback (not own-scope shadowing) is what the tests exercise.
	if err := os.MkdirAll(filepath.Join(fx.aliceDir, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(p, content string) {
		t.Helper()
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(fx.aliceDir, "docs", "report.txt"), "alice-report")
	write(filepath.Join(fx.aliceDir, "docs2-other.txt"), "sibling")
	if err := os.MkdirAll(filepath.Join(fx.aliceDir, "docs", "private"), 0o755); err != nil {
		t.Fatal(err)
	}
	write(filepath.Join(fx.aliceDir, "docs", "private", "secret.txt"), "top-secret")

	db, err := storm.Open(filepath.Join(root, "db"))
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	st, err := bolt.NewStorage(db)
	if err != nil {
		t.Fatalf("failed to get storage: %v", err)
	}
	// IDs: alice=1, bob=2, mallory=3.
	alice := &users.User{Username: "alice", Password: "pw", Perm: fx.alice, Rules: aliceRules}
	if err := st.Users.Save(alice); err != nil {
		t.Fatalf("failed to save alice: %v", err)
	}
	if err := st.Users.Save(&users.User{Username: "bob", Password: "pw", Perm: fx.bob}); err != nil {
		t.Fatalf("failed to save bob: %v", err)
	}
	if err := st.Users.Save(&users.User{Username: "mallory", Password: "pw", Perm: fx.mall}); err != nil {
		t.Fatalf("failed to save mallory: %v", err)
	}
	if err := st.Settings.Save(&settings.Settings{Key: fx.key}); err != nil {
		t.Fatalf("failed to save settings: %v", err)
	}
	st.Users = &perUserFSStore{Store: st.Users, dirs: map[uint]string{1: fx.aliceDir, 2: fx.bobDir, 3: t.TempDir()}}
	fx.store = st
	return fx
}

func (fx *accessFixture) token(t *testing.T, id uint, name string, perm users.Permissions) string {
	t.Helper()
	return signGrantToken(t, id, name, perm, fx.key)
}

func (fx *accessFixture) do(t *testing.T, handler handleFunc, method, target, body, token string) *httptest.ResponseRecorder {
	t.Helper()
	req, err := http.NewRequest(method, target, strings.NewReader(body))
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	req.Header.Set("X-Auth", token)
	rec := httptest.NewRecorder()
	handle(handler, "", fx.store, &settings.Server{}).ServeHTTP(rec, req)
	return rec
}

func (fx *accessFixture) grant(t *testing.T, role grants.Role) {
	t.Helper()
	g := &grants.Grant{Path: "/docs", OwnerID: 1, GranteeID: 2, Role: role}
	if err := fx.store.Grants.Save(g); err != nil {
		t.Fatalf("failed to save grant: %v", err)
	}
}

func TestGrantReadAccess(t *testing.T) {
	t.Parallel()

	fx := newAccessFixture(t, nil)
	fx.grant(t, grants.RoleViewer)
	bob := fx.token(t, 2, "bob", fx.bob)
	mall := fx.token(t, 3, "mallory", fx.mall)

	if rec := fx.do(t, resourceGetHandler, http.MethodGet, "/docs/report.txt", "", bob); rec.Code != http.StatusOK {
		t.Fatalf("bob resource get = %d, want 200 body=%q", rec.Code, rec.Body.String())
	}
	if rec := fx.do(t, rawHandler, http.MethodGet, "/docs/report.txt", "", bob); rec.Code != http.StatusOK || rec.Body.String() != "alice-report" {
		t.Fatalf("bob raw = %d body=%q, want 200 alice-report", rec.Code, rec.Body.String())
	}
	// Directory listing through the grant works.
	if rec := fx.do(t, resourceGetHandler, http.MethodGet, "/docs", "", bob); rec.Code != http.StatusOK {
		t.Fatalf("bob dir listing = %d, want 200", rec.Code)
	}
	// Stranger without a grant sees nothing (404, not 403: no existence leak).
	if rec := fx.do(t, resourceGetHandler, http.MethodGet, "/docs/report.txt", "", mall); rec.Code != http.StatusNotFound {
		t.Fatalf("mallory get = %d, want 404", rec.Code)
	}
	// Outside the grant subtree, even for the grantee.
	if rec := fx.do(t, resourceGetHandler, http.MethodGet, "/docs2-other.txt", "", bob); rec.Code != http.StatusNotFound {
		t.Fatalf("bob outside grant = %d, want 404", rec.Code)
	}
	// Byte-prefix sibling (/docs vs /docs-other) must not match.
	if err := os.WriteFile(filepath.Join(fx.aliceDir, "docs-other.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if rec := fx.do(t, resourceGetHandler, http.MethodGet, "/docs-other.txt", "", bob); rec.Code != http.StatusNotFound {
		t.Fatalf("bob sibling-prefix = %d, want 404", rec.Code)
	}
}

func TestGrantWriteRoleGate(t *testing.T) {
	t.Parallel()

	fx := newAccessFixture(t, nil)
	fx.grant(t, grants.RoleViewer)
	bob := fx.token(t, 2, "bob", fx.bob)

	// Viewer cannot overwrite.
	if rec := fx.do(t, resourcePutHandler, http.MethodPut, "/docs/report.txt", "hacked", bob); rec.Code != http.StatusForbidden {
		t.Fatalf("viewer put = %d, want 403", rec.Code)
	}
	// Viewer cannot upload.
	if rec := fx.do(t, resourcePostHandler(diskcache.NewNoOp()), http.MethodPost, "/docs/new.txt", "new", bob); rec.Code != http.StatusForbidden {
		t.Fatalf("viewer post = %d, want 403", rec.Code)
	}
	// Viewer cannot delete (also lacks the global Delete perm).
	if rec := fx.do(t, resourceDeleteHandler(diskcache.NewNoOp()), http.MethodDelete, "/docs/report.txt", "", bob); rec.Code != http.StatusForbidden {
		t.Fatalf("viewer delete = %d, want 403", rec.Code)
	}

	// Upgrade to editor: overwrite + new upload work (bob has Modify/Create).
	g, err := fx.store.Grants.Get(1)
	if err != nil {
		t.Fatalf("Get grant failed: %v", err)
	}
	g.Role = grants.RoleEditor
	if err := fx.store.Grants.Save(g); err != nil {
		t.Fatalf("upgrade failed: %v", err)
	}
	if rec := fx.do(t, resourcePutHandler, http.MethodPut, "/docs/report.txt", "bob-edit", bob); rec.Code != http.StatusOK {
		t.Fatalf("editor put = %d, want 200 body=%q", rec.Code, rec.Body.String())
	}
	if data, _ := os.ReadFile(filepath.Join(fx.aliceDir, "docs", "report.txt")); string(data) != "bob-edit" {
		t.Fatalf("owner file = %q, want bob-edit", string(data))
	}
	if rec := fx.do(t, resourcePostHandler(diskcache.NewNoOp()), http.MethodPost, "/docs/fresh.txt", "fresh", bob); rec.Code != http.StatusOK {
		t.Fatalf("editor post = %d, want 200 body=%q", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(fx.aliceDir, "docs", "fresh.txt")); err != nil {
		t.Fatalf("uploaded file missing in owner scope: %v", err)
	}
	// Global perm ceiling still applies: bob has no Delete perm.
	if rec := fx.do(t, resourceDeleteHandler(diskcache.NewNoOp()), http.MethodDelete, "/docs/fresh.txt", "", bob); rec.Code != http.StatusForbidden {
		t.Fatalf("editor-without-delete-perm delete = %d, want 403", rec.Code)
	}
}

func TestGrantRenameScope(t *testing.T) {
	t.Parallel()

	fx := newAccessFixture(t, nil)
	fx.grant(t, grants.RoleEditor)
	bob := fx.token(t, 2, "bob", fx.bob)

	rename := func(src, dst string) int {
		target := "/api/resources" + src + "?action=rename&destination=" + url.QueryEscape(dst)
		req, _ := http.NewRequest(http.MethodPatch, target, nil)
		req.Header.Set("X-Auth", bob)
		rec := httptest.NewRecorder()
		handle(resourcePatchHandler(diskcache.NewNoOp()), "/api/resources", fx.store, &settings.Server{}).ServeHTTP(rec, req)
		return rec.Code
	}

	// Rename inside the grant works.
	if code := rename("/docs/report.txt", "/docs/renamed.txt"); code != http.StatusOK {
		t.Fatalf("rename inside grant = %d, want 200", code)
	}
	if _, err := os.Stat(filepath.Join(fx.aliceDir, "docs", "renamed.txt")); err != nil {
		t.Fatalf("renamed file missing: %v", err)
	}
	// Rename escaping the grant is forbidden.
	if code := rename("/docs/renamed.txt", "/escaped.txt"); code != http.StatusForbidden {
		t.Fatalf("rename escape = %d, want 403", code)
	}
	if _, err := os.Stat(filepath.Join(fx.aliceDir, "escaped.txt")); err == nil {
		t.Fatalf("VULNERABLE: file escaped the grant scope")
	}
}

func TestGrantRevocation(t *testing.T) {
	t.Parallel()

	fx := newAccessFixture(t, nil)
	fx.grant(t, grants.RoleViewer)
	bob := fx.token(t, 2, "bob", fx.bob)

	if rec := fx.do(t, resourceGetHandler, http.MethodGet, "/docs/report.txt", "", bob); rec.Code != http.StatusOK {
		t.Fatalf("before revoke = %d, want 200", rec.Code)
	}
	if err := fx.store.Grants.Delete(1); err != nil {
		t.Fatalf("revoke failed: %v", err)
	}
	if rec := fx.do(t, resourceGetHandler, http.MethodGet, "/docs/report.txt", "", bob); rec.Code != http.StatusNotFound {
		t.Fatalf("after revoke = %d, want 404", rec.Code)
	}
}

func TestGrantDiesWhenOwnerLosesShare(t *testing.T) {
	t.Parallel()

	fx := newAccessFixture(t, nil)
	fx.grant(t, grants.RoleViewer)
	bob := fx.token(t, 2, "bob", fx.bob)

	alice, err := fx.store.Users.Get("", false, uint(1))
	if err != nil {
		t.Fatalf("get alice failed: %v", err)
	}
	alice.Perm.Share = false
	if err := fx.store.Users.Update(alice, "Perm"); err != nil {
		t.Fatalf("update alice failed: %v", err)
	}

	if rec := fx.do(t, resourceGetHandler, http.MethodGet, "/docs/report.txt", "", bob); rec.Code != http.StatusNotFound {
		t.Fatalf("after owner lost share = %d, want 404", rec.Code)
	}
}

func TestGrantOwnerRulesApply(t *testing.T) {
	t.Parallel()

	fx := newAccessFixture(t, []rules.Rule{{Allow: false, Path: "/docs/private"}})
	fx.grant(t, grants.RoleViewer)
	bob := fx.token(t, 2, "bob", fx.bob)

	if rec := fx.do(t, resourceGetHandler, http.MethodGet, "/docs/private/secret.txt", "", bob); rec.Code != http.StatusForbidden {
		t.Fatalf("blocked path via grant = %d, want 403", rec.Code)
	}
	if rec := fx.do(t, resourceGetHandler, http.MethodGet, "/docs/report.txt", "", bob); rec.Code != http.StatusOK {
		t.Fatalf("allowed path via grant = %d, want 200", rec.Code)
	}
}

func TestGrantSymlinkEscapeBlocked(t *testing.T) {
	t.Parallel()

	fx := newAccessFixture(t, nil)
	outside := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(outside, []byte("OUT-OF-SCOPE"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(fx.aliceDir, "docs", "link.txt")); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}
	fx.grant(t, grants.RoleViewer)
	bob := fx.token(t, 2, "bob", fx.bob)

	rec := fx.do(t, rawHandler, http.MethodGet, "/docs/link.txt", "", bob)
	if rec.Code == http.StatusOK && strings.Contains(rec.Body.String(), "OUT-OF-SCOPE") {
		t.Fatalf("VULNERABLE: out-of-scope content exfiltrated via grant symlink")
	}
}

func TestGrantDeleteCascade(t *testing.T) {
	t.Parallel()

	fx := newAccessFixture(t, nil)
	fx.grant(t, grants.RoleViewer)
	alice := fx.token(t, 1, "alice", fx.alice)
	bob := fx.token(t, 2, "bob", fx.bob)

	if rec := fx.do(t, resourceDeleteHandler(diskcache.NewNoOp()), http.MethodDelete, "/docs", "", alice); rec.Code != http.StatusNoContent {
		t.Fatalf("owner delete dir = %d, want 204 body=%q", rec.Code, rec.Body.String())
	}
	remaining, err := fx.store.Grants.FindByOwner(1)
	if err == nil && len(remaining) != 0 {
		t.Fatalf("grants after owner delete = %d, want 0", len(remaining))
	}
	if rec := fx.do(t, resourceGetHandler, http.MethodGet, "/docs/report.txt", "", bob); rec.Code != http.StatusNotFound {
		t.Fatalf("bob after owner delete = %d, want 404", rec.Code)
	}
}
