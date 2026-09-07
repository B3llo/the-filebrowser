package bolt

import (
	"os"
	"sort"
	"testing"

	"github.com/asdine/storm/v3"

	"github.com/B3llo/the-filebrowser/grants"
)

func newTestGrantsBackend(t *testing.T) grantsBackend {
	t.Helper()

	f, err := os.CreateTemp(t.TempDir(), "grants-*.db")
	if err != nil {
		t.Fatalf("failed to create temp db: %v", err)
	}
	_ = f.Close()

	db, err := storm.Open(f.Name())
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	return grantsBackend{db: db}
}

func remainingGrantIDs(t *testing.T, s grantsBackend) []uint {
	t.Helper()

	list, err := s.All()
	if err != nil {
		t.Fatalf("All returned error: %v", err)
	}

	ids := make([]uint, 0, len(list))
	for _, g := range list {
		ids = append(ids, g.ID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func saveGrant(t *testing.T, s grantsBackend, g *grants.Grant) uint {
	t.Helper()
	if err := s.Save(g); err != nil {
		t.Fatalf("failed to save grant: %v", err)
	}
	if g.ID == 0 {
		t.Fatalf("save did not assign an ID")
	}
	return g.ID
}

func TestGrantsDeleteWithPathPrefix(t *testing.T) {
	t.Parallel()

	s := newTestGrantsBackend(t)

	keepABC := saveGrant(t, s, &grants.Grant{Path: "/abc", OwnerID: 1, GranteeID: 2, Role: grants.RoleViewer})
	keepOther := saveGrant(t, s, &grants.Grant{Path: "/other", OwnerID: 1, GranteeID: 2, Role: grants.RoleViewer})
	u2a := saveGrant(t, s, &grants.Grant{Path: "/a", OwnerID: 2, GranteeID: 3, Role: grants.RoleViewer})
	u2child := saveGrant(t, s, &grants.Grant{Path: "/a/child.txt", OwnerID: 2, GranteeID: 3, Role: grants.RoleViewer})
	saveGrant(t, s, &grants.Grant{Path: "/a", OwnerID: 1, GranteeID: 2, Role: grants.RoleViewer})
	saveGrant(t, s, &grants.Grant{Path: "/a/child.txt", OwnerID: 1, GranteeID: 2, Role: grants.RoleEditor})

	// Owner 1 deletes /a: only owner 1's /a and descendants go away.
	// /abc (byte-prefix sibling) and all of owner 2's grants must remain.
	if err := s.DeleteWithPathPrefix("/a", 1); err != nil {
		t.Fatalf("DeleteWithPathPrefix returned error: %v", err)
	}

	got := remainingGrantIDs(t, s)
	want := []uint{keepABC, keepOther, u2a, u2child}
	sort.Slice(want, func(i, j int) bool { return want[i] < want[j] })
	if len(got) != len(want) {
		t.Fatalf("remaining ids = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("remaining ids = %v, want %v", got, want)
		}
	}
}

func TestGrantsDeleteWithPathPrefixNoMatch(t *testing.T) {
	t.Parallel()

	s := newTestGrantsBackend(t)

	if err := s.DeleteWithPathPrefix("/a", 1); err != nil {
		t.Fatalf("DeleteWithPathPrefix on empty store returned error: %v", err)
	}
}

func TestGrantsDeleteByUser(t *testing.T) {
	t.Parallel()

	s := newTestGrantsBackend(t)

	owned := saveGrant(t, s, &grants.Grant{Path: "/a", OwnerID: 1, GranteeID: 2, Role: grants.RoleViewer})
	received := saveGrant(t, s, &grants.Grant{Path: "/b", OwnerID: 2, GranteeID: 1, Role: grants.RoleEditor})
	unrelated := saveGrant(t, s, &grants.Grant{Path: "/c", OwnerID: 2, GranteeID: 3, Role: grants.RoleViewer})

	if err := s.DeleteByUser(1); err != nil {
		t.Fatalf("DeleteByUser returned error: %v", err)
	}

	got := remainingGrantIDs(t, s)
	if len(got) != 1 || got[0] != unrelated {
		t.Fatalf("remaining ids = %v, want [%d] (owned=%d received=%d should be gone)", got, unrelated, owned, received)
	}
}

func TestGrantsFindByOwnerAndGrantee(t *testing.T) {
	t.Parallel()

	s := newTestGrantsBackend(t)

	g1 := saveGrant(t, s, &grants.Grant{Path: "/a", OwnerID: 1, GranteeID: 2, Role: grants.RoleViewer})
	g2 := saveGrant(t, s, &grants.Grant{Path: "/b", OwnerID: 1, GranteeID: 3, Role: grants.RoleEditor})
	g3 := saveGrant(t, s, &grants.Grant{Path: "/c", OwnerID: 2, GranteeID: 2, Role: grants.RoleViewer})

	owned, err := s.FindByOwner(1)
	if err != nil {
		t.Fatalf("FindByOwner returned error: %v", err)
	}
	if len(owned) != 2 {
		t.Fatalf("FindByOwner(1) = %d grants, want 2 (g1=%d g2=%d)", len(owned), g1, g2)
	}

	received, err := s.FindByGrantee(2)
	if err != nil {
		t.Fatalf("FindByGrantee returned error: %v", err)
	}
	if len(received) != 2 {
		t.Fatalf("FindByGrantee(2) = %d grants, want 2 (g1=%d g3=%d)", len(received), g1, g3)
	}
}
