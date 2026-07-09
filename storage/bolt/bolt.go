package bolt

import (
	"github.com/asdine/storm/v3"

	"github.com/B3llo/the-filebrowser/auth"
	"github.com/B3llo/the-filebrowser/settings"
	"github.com/B3llo/the-filebrowser/share"
	"github.com/B3llo/the-filebrowser/sources"
	"github.com/B3llo/the-filebrowser/storage"
	"github.com/B3llo/the-filebrowser/users"
)

// NewStorage creates a storage.Storage based on Bolt DB.
func NewStorage(db *storm.DB) (*storage.Storage, error) {
	userStore := users.NewStorage(usersBackend{db: db})
	shareStore := share.NewStorage(shareBackend{db: db})
	settingsStore := settings.NewStorage(settingsBackend{db: db})
	authStore := auth.NewStorage(authBackend{db: db}, userStore)
	sourcesStore := sources.NewStorage(sourcesBackend{db: db})

	err := save(db, "version", 2)
	if err != nil {
		return nil, err
	}

	return &storage.Storage{
		Auth:     authStore,
		Users:    userStore,
		Share:    shareStore,
		Settings: settingsStore,
		Sources:  sourcesStore,
	}, nil
}
