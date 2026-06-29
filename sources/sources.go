package sources

import "time"

// Source is an independent filesystem location that FileBrowser can browse.
// Each Source is mounted as its own ScopedFs (see files/scoped.go), so the
// per-root symlink confinement is inherited unchanged. Path is absolute on the
// host filesystem and is trusted only because it is defined by an admin.
type Source struct {
	ID        uint      `storm:"id,increment" json:"id"`
	Name      string    `storm:"unique" json:"name"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"createdAt"`
}

// ImplicitSourceID is the synthetic id of the virtual source that represents a
// legacy user's classic Scope when no real sources are defined. It lets the
// application behave exactly as before until the administrator defines real
// sources.
const ImplicitSourceID uint = 0

// ImplicitName is the display name of the virtual legacy source.
const ImplicitName = "My Files"
