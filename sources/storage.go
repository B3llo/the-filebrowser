package sources

import (
	"errors"

	fberrors "github.com/B3llo/the-filebrowser/errors"
)

// StorageBackend is the interface to implement for a sources storage.
type StorageBackend interface {
	GetBy(interface{}) (*Source, error)
	Gets() ([]*Source, error)
	Save(s *Source) error
	Update(s *Source, fields ...string) error
	DeleteByID(uint) error
}

// Store is the public interface consumed by the rest of the application.
type Store interface {
	Get(id interface{}) (*Source, error)
	Gets() ([]*Source, error)
	Save(s *Source) error
	Update(s *Source, fields ...string) error
	Delete(id uint) error
}

// Storage is a sources storage.
type Storage struct {
	back StorageBackend
}

// NewStorage creates a sources storage from a backend.
func NewStorage(back StorageBackend) *Storage {
	return &Storage{back: back}
}

// Get returns a source by its numeric id or its unique name.
func (s *Storage) Get(id interface{}) (*Source, error) {
	src, err := s.back.GetBy(id)
	if err != nil {
		if errors.Is(err, fberrors.ErrNotExist) {
			return nil, err
		}
		return nil, err
	}
	return src, nil
}

// Gets returns all defined sources.
func (s *Storage) Gets() ([]*Source, error) {
	return s.back.Gets()
}

// Save creates or updates a source.
func (s *Storage) Save(src *Source) error {
	return s.back.Save(src)
}

// Update updates specific fields of a source.
func (s *Storage) Update(src *Source, fields ...string) error {
	return s.back.Update(src, fields...)
}

// Delete removes a source by id.
func (s *Storage) Delete(id uint) error {
	return s.back.DeleteByID(id)
}
