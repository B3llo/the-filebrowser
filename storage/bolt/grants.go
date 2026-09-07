package bolt

import (
	"errors"
	"strings"

	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"

	fberrors "github.com/B3llo/the-filebrowser/errors"
	"github.com/B3llo/the-filebrowser/grants"
)

type grantsBackend struct {
	db *storm.DB
}

func (s grantsBackend) All() ([]*grants.Grant, error) {
	var v []*grants.Grant
	err := s.db.All(&v)
	if errors.Is(err, storm.ErrNotFound) {
		return v, fberrors.ErrNotExist
	}

	return v, err
}

func (s grantsBackend) FindByGrantee(id uint) ([]*grants.Grant, error) {
	var v []*grants.Grant
	err := s.db.Select(q.Eq("GranteeID", id)).Find(&v)
	if errors.Is(err, storm.ErrNotFound) {
		return v, fberrors.ErrNotExist
	}

	return v, err
}

func (s grantsBackend) FindByOwner(id uint) ([]*grants.Grant, error) {
	var v []*grants.Grant
	err := s.db.Select(q.Eq("OwnerID", id)).Find(&v)
	if errors.Is(err, storm.ErrNotFound) {
		return v, fberrors.ErrNotExist
	}

	return v, err
}

func (s grantsBackend) Get(id uint) (*grants.Grant, error) {
	var v grants.Grant
	err := s.db.One("ID", id, &v)
	if errors.Is(err, storm.ErrNotFound) {
		return nil, fberrors.ErrNotExist
	}

	return &v, err
}

func (s grantsBackend) Save(g *grants.Grant) error {
	return s.db.Save(g)
}

func (s grantsBackend) Delete(id uint) error {
	err := s.db.DeleteStruct(&grants.Grant{ID: id})
	if errors.Is(err, storm.ErrNotFound) {
		return nil
	}
	return err
}

func (s grantsBackend) DeleteWithPathPrefix(pathPrefix string, ownerID uint) error {
	var items []grants.Grant
	if err := s.db.Prefix("Path", pathPrefix, &items); err != nil {
		if errors.Is(err, storm.ErrNotFound) {
			return nil
		}
		return err
	}

	prefix := strings.TrimRight(pathPrefix, "/")

	var err error
	for _, g := range items {
		if g.OwnerID != ownerID {
			continue
		}

		if g.Path != prefix && !strings.HasPrefix(g.Path, prefix+"/") {
			continue
		}

		err = errors.Join(err, s.db.DeleteStruct(&grants.Grant{ID: g.ID}))
	}
	return err
}

func (s grantsBackend) DeleteByUser(userID uint) error {
	var owned []grants.Grant
	if err := s.db.Select(q.Eq("OwnerID", userID)).Find(&owned); err != nil {
		if !errors.Is(err, storm.ErrNotFound) {
			return err
		}
	}

	var received []grants.Grant
	if err := s.db.Select(q.Eq("GranteeID", userID)).Find(&received); err != nil {
		if !errors.Is(err, storm.ErrNotFound) {
			return err
		}
	}

	var err error
	seen := map[uint]struct{}{}
	for _, g := range append(owned, received...) {
		if _, ok := seen[g.ID]; ok {
			continue
		}
		seen[g.ID] = struct{}{}
		err = errors.Join(err, s.db.DeleteStruct(&grants.Grant{ID: g.ID}))
	}
	return err
}
