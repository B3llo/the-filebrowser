package storage

import (
	"github.com/B3llo/the-filebrowser/auth"
	"github.com/B3llo/the-filebrowser/settings"
	"github.com/B3llo/the-filebrowser/share"
	"github.com/B3llo/the-filebrowser/sources"
	"github.com/B3llo/the-filebrowser/users"
)

// Storage is a storage powered by a Backend which makes the necessary
// verifications when fetching and saving data to ensure consistency.
type Storage struct {
	Users    users.Store
	Share    *share.Storage
	Auth     *auth.Storage
	Settings *settings.Storage
	Sources  sources.Store
}
