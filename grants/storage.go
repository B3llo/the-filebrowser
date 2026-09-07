package grants

import (
	"time"

	fberrors "github.com/B3llo/the-filebrowser/errors"
)

// StorageBackend is the interface to implement for a grant storage.
type StorageBackend interface {
	All() ([]*Grant, error)
	FindByGrantee(id uint) ([]*Grant, error)
	FindByOwner(id uint) ([]*Grant, error)
	Get(id uint) (*Grant, error)
	Save(g *Grant) error
	Delete(id uint) error
	DeleteWithPathPrefix(path string, ownerID uint) error
	DeleteByUser(userID uint) error
}

// Storage is a grant storage with lazy expiration.
type Storage struct {
	back StorageBackend
}

// NewStorage creates a grant storage from a backend.
func NewStorage(back StorageBackend) *Storage {
	return &Storage{back: back}
}

// filterExpired removes expired grants (deleting them) and returns the
// active ones. Unlike a splice-in-loop, it never skips consecutive entries.
func (s *Storage) filterExpired(list []*Grant) ([]*Grant, error) {
	now := time.Now().Unix()
	active := make([]*Grant, 0, len(list))
	for _, g := range list {
		if g.IsExpired(now) {
			if err := s.Delete(g.ID); err != nil {
				return nil, err
			}
			continue
		}
		active = append(active, g)
	}
	return active, nil
}

// All returns all active grants.
func (s *Storage) All() ([]*Grant, error) {
	list, err := s.back.All()
	if err != nil {
		return nil, err
	}
	return s.filterExpired(list)
}

// FindByGrantee returns active grants where id is the grantee.
func (s *Storage) FindByGrantee(id uint) ([]*Grant, error) {
	list, err := s.back.FindByGrantee(id)
	if err != nil {
		return nil, err
	}
	return s.filterExpired(list)
}

// FindByOwner returns active grants where id is the owner.
func (s *Storage) FindByOwner(id uint) ([]*Grant, error) {
	list, err := s.back.FindByOwner(id)
	if err != nil {
		return nil, err
	}
	return s.filterExpired(list)
}

// FindActive returns active grants for a grantee (alias kept explicit for call sites).
func (s *Storage) FindActive(granteeID uint) ([]*Grant, error) {
	return s.FindByGrantee(granteeID)
}

// Get returns a single grant, deleting it if expired.
func (s *Storage) Get(id uint) (*Grant, error) {
	g, err := s.back.Get(id)
	if err != nil {
		return nil, err
	}
	if g.IsExpired(time.Now().Unix()) {
		if err := s.Delete(g.ID); err != nil {
			return nil, err
		}
		return nil, fberrors.ErrNotExist
	}
	return g, nil
}

// Save validates then persists a grant.
func (s *Storage) Save(g *Grant) error {
	if err := g.Validate(); err != nil {
		return err
	}
	return s.back.Save(g)
}

// Delete removes a grant by id.
func (s *Storage) Delete(id uint) error {
	return s.back.Delete(id)
}

// DeleteWithPathPrefix removes owner grants at path or below it.
func (s *Storage) DeleteWithPathPrefix(path string, ownerID uint) error {
	return s.back.DeleteWithPathPrefix(path, ownerID)
}

// DeleteByUser removes every grant where the user is owner or grantee.
func (s *Storage) DeleteByUser(userID uint) error {
	return s.back.DeleteByUser(userID)
}
