package users

import "testing"

// Interface is implemented by storage
var _ Store = &Storage{}

// mockBackend is a no-op StorageBackend good enough to exercise the
// in-memory session revocation bookkeeping in Storage.
type mockBackend struct{}

func (mockBackend) GetBy(interface{}) (*User, error) { return nil, nil }
func (mockBackend) Gets() ([]*User, error)           { return nil, nil }
func (mockBackend) Save(*User) error                 { return nil }
func (mockBackend) Update(*User, ...string) error    { return nil }
func (mockBackend) DeleteByID(uint) error            { return nil }
func (mockBackend) DeleteByUsername(string) error    { return nil }
func (mockBackend) CountAdmins() (int, error)        { return 0, nil }

func TestUpdateRevokesSessionsOnlyForSensitiveFields(t *testing.T) {
	t.Parallel()

	preferences := []string{"Theme", "sorting", "ViewMode", "FolderColors", "Avatar", "Locale", "DisplayName"}
	for _, field := range preferences {
		field := field
		t.Run("preference_"+field, func(t *testing.T) {
			t.Parallel()

			st := NewStorage(mockBackend{})
			u := &User{ID: 7, Username: "alice", Password: "hash"}

			if err := st.Update(u, field); err != nil {
				t.Fatalf("Update(%q): %v", field, err)
			}
			if got := st.LastUpdate(7); got != 0 {
				t.Errorf("LastUpdate after Update(%q) = %d, want 0", field, got)
			}
		})
	}

	sensitive := []string{"Password", "Username", "Perm", "Scope", "Sources", "LockPassword", "Commands", "Rules"}
	for _, field := range sensitive {
		field := field
		t.Run("sensitive_"+field, func(t *testing.T) {
			t.Parallel()

			st := NewStorage(mockBackend{})
			u := &User{ID: 7, Username: "alice", Password: "hash"}

			if err := st.Update(u, field); err != nil {
				t.Fatalf("Update(%q): %v", field, err)
			}
			if got := st.LastUpdate(7); got == 0 {
				t.Errorf("LastUpdate after Update(%q) = 0, want a revocation timestamp", field)
			}
		})
	}
}

func TestUpdateWithoutFieldsRevokesSessions(t *testing.T) {
	t.Parallel()

	st := NewStorage(mockBackend{})
	u := &User{ID: 9, Username: "bob", Password: "hash"}

	if err := st.Update(u); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got := st.LastUpdate(9); got == 0 {
		t.Error("LastUpdate after full Update = 0, want a revocation timestamp")
	}
}
