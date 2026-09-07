package grants

import (
	"testing"
	"time"

	fberrors "github.com/B3llo/the-filebrowser/errors"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		grant   Grant
		wantErr bool
		wantPth string
	}{
		"viewer ok":         {Grant{Path: "/a", OwnerID: 1, GranteeID: 2, Role: RoleViewer}, false, "/a"},
		"editor ok":         {Grant{Path: "/a/b", OwnerID: 1, GranteeID: 2, Role: RoleEditor}, false, "/a/b"},
		"bad role":          {Grant{Path: "/a", OwnerID: 1, GranteeID: 2, Role: "admin"}, true, ""},
		"empty role":        {Grant{Path: "/a", OwnerID: 1, GranteeID: 2}, true, ""},
		"empty path":        {Grant{OwnerID: 1, GranteeID: 2, Role: RoleViewer}, true, ""},
		"self grant":        {Grant{Path: "/a", OwnerID: 1, GranteeID: 1, Role: RoleViewer}, true, ""},
		"zero owner":        {Grant{Path: "/a", GranteeID: 2, Role: RoleViewer}, true, ""},
		"zero grantee":      {Grant{Path: "/a", OwnerID: 1, Role: RoleViewer}, true, ""},
		"missing slash":     {Grant{Path: "a", OwnerID: 1, GranteeID: 2, Role: RoleViewer}, false, "/a"},
		"dotdot normalized": {Grant{Path: "/a/../b", OwnerID: 1, GranteeID: 2, Role: RoleViewer}, false, "/b"},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := tc.grant.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if tc.grant.Path != tc.wantPth {
					t.Fatalf("path = %q, want %q", tc.grant.Path, tc.wantPth)
				}
			}
		})
	}
}

func TestExpireFromStrings(t *testing.T) {
	t.Parallel()

	if exp, err := ExpireFromStrings("", "hours"); err != nil || exp != 0 {
		t.Fatalf("empty should be permanent: exp=%d err=%v", exp, err)
	}
	before := time.Now().Unix()
	exp, err := ExpireFromStrings("2", "hours")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exp-before < 7190 || exp-before > 7210 {
		t.Fatalf("2 hours expiry off: delta=%d", exp-before)
	}
	if _, err := ExpireFromStrings("abc", "hours"); err == nil {
		t.Fatalf("expected error for non-numeric expires")
	}
	if _, err := ExpireFromStrings("-1", "hours"); err == nil {
		t.Fatalf("expected error for negative expires")
	}
	// Unknown unit defaults to hours (mirrors share behavior).
	expDays, _ := ExpireFromStrings("1", "days")
	expHours, _ := ExpireFromStrings("24", "hours")
	if expDays-expHours > 5 || expHours-expDays > 5 {
		t.Fatalf("days/hours mismatch: %d vs %d", expDays, expHours)
	}
}

func TestIsExpired(t *testing.T) {
	t.Parallel()
	now := time.Now().Unix()
	if (&Grant{Expire: 0}).IsExpired(now) {
		t.Fatalf("permanent grant should not be expired")
	}
	if !(&Grant{Expire: now - 1}).IsExpired(now) {
		t.Fatalf("past grant should be expired")
	}
	if (&Grant{Expire: now + 100}).IsExpired(now) {
		t.Fatalf("future grant should not be expired")
	}
	if err := (&Grant{}).Validate(); err != fberrors.ErrInvalidRequestParams {
		t.Fatalf("expected ErrInvalidRequestParams, got %v", err)
	}
}
