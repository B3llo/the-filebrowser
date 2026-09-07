package grants

import (
	"errors"
	"testing"
	"time"

	fberrors "github.com/B3llo/the-filebrowser/errors"
)

type fakeBackend struct {
	items   []*Grant
	nextID  uint
	deleted []uint
}

func (f *fakeBackend) All() ([]*Grant, error) { return f.items, nil }
func (f *fakeBackend) FindByGrantee(id uint) ([]*Grant, error) {
	var out []*Grant
	for _, g := range f.items {
		if g.GranteeID == id {
			out = append(out, g)
		}
	}
	return out, nil
}
func (f *fakeBackend) FindByOwner(id uint) ([]*Grant, error) {
	var out []*Grant
	for _, g := range f.items {
		if g.OwnerID == id {
			out = append(out, g)
		}
	}
	return out, nil
}
func (f *fakeBackend) Get(id uint) (*Grant, error) {
	for _, g := range f.items {
		if g.ID == id {
			return g, nil
		}
	}
	return nil, fberrors.ErrNotExist
}
func (f *fakeBackend) Save(g *Grant) error {
	if g.ID == 0 {
		f.nextID++
		g.ID = f.nextID
	}
	f.items = append(f.items, g)
	return nil
}
func (f *fakeBackend) Delete(id uint) error {
	f.deleted = append(f.deleted, id)
	kept := make([]*Grant, 0, len(f.items))
	for _, g := range f.items {
		if g.ID != id {
			kept = append(kept, g)
		}
	}
	f.items = kept
	return nil
}
func (f *fakeBackend) DeleteWithPathPrefix(_ string, _ uint) error { return nil }
func (f *fakeBackend) DeleteByUser(_ uint) error                   { return nil }

func TestFilterExpiredRemovesConsecutive(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()
	// Three consecutive expired grants: the old splice-in-loop bug in
	// share/storage.go would skip the middle one. Filter must drop all.
	be := &fakeBackend{items: []*Grant{
		{ID: 1, Path: "/a", OwnerID: 1, GranteeID: 2, Role: RoleViewer, Expire: now - 10},
		{ID: 2, Path: "/b", OwnerID: 1, GranteeID: 2, Role: RoleViewer, Expire: now - 5},
		{ID: 3, Path: "/c", OwnerID: 1, GranteeID: 2, Role: RoleViewer, Expire: now - 1},
		{ID: 4, Path: "/d", OwnerID: 1, GranteeID: 2, Role: RoleViewer},
	}}
	s := NewStorage(be)

	got, err := s.All()
	if err != nil {
		t.Fatalf("All returned error: %v", err)
	}
	if len(got) != 1 || got[0].ID != 4 {
		t.Fatalf("All = %v, want only grant 4", got)
	}
	if len(be.deleted) != 3 {
		t.Fatalf("deleted = %v, want 3 lazy deletes", be.deleted)
	}
}

func TestGetExpiredDeletesAndReturnsNotExist(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()
	be := &fakeBackend{items: []*Grant{
		{ID: 7, Path: "/x", OwnerID: 1, GranteeID: 2, Role: RoleEditor, Expire: now - 1},
	}}
	s := NewStorage(be)

	if _, err := s.Get(7); !errors.Is(err, fberrors.ErrNotExist) {
		t.Fatalf("Get expired = %v, want ErrNotExist", err)
	}
	if len(be.deleted) != 1 || be.deleted[0] != 7 {
		t.Fatalf("deleted = %v, want [7]", be.deleted)
	}
}

func TestSaveValidates(t *testing.T) {
	t.Parallel()

	s := NewStorage(&fakeBackend{})
	if err := s.Save(&Grant{Path: "/a", OwnerID: 1, GranteeID: 1, Role: RoleViewer}); err == nil {
		t.Fatalf("self-grant save should fail")
	}
	if err := s.Save(&Grant{Path: "/a", OwnerID: 1, GranteeID: 2, Role: "owner"}); err == nil {
		t.Fatalf("bad-role save should fail")
	}
	if err := s.Save(&Grant{Path: "/a", OwnerID: 1, GranteeID: 2, Role: RoleViewer}); err != nil {
		t.Fatalf("valid save failed: %v", err)
	}
}
