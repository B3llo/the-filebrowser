package bolt

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/asdine/storm/v3"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/sources"
)

type sourcesBackend struct {
	db *storm.DB
}

func (st sourcesBackend) GetBy(i interface{}) (src *sources.Source, err error) {
	src = &sources.Source{}

	var arg string
	switch i.(type) {
	case uint:
		arg = "ID"
	case string:
		arg = "Name"
	default:
		return nil, fberrors.ErrInvalidDataType
	}

	err = st.db.One(arg, i, src)
	if err != nil {
		if errors.Is(err, storm.ErrNotFound) {
			return nil, fberrors.ErrNotExist
		}
		return nil, err
	}

	return src, nil
}

func (st sourcesBackend) Gets() ([]*sources.Source, error) {
	var all []*sources.Source
	err := st.db.All(&all)
	if errors.Is(err, storm.ErrNotFound) {
		return nil, fberrors.ErrNotExist
	}
	if err != nil {
		return all, err
	}
	return all, err
}

func (st sourcesBackend) Update(src *sources.Source, fields ...string) error {
	if len(fields) == 0 {
		return st.Save(src)
	}

	for _, field := range fields {
		srcField := reflect.ValueOf(src).Elem().FieldByName(field)
		if !srcField.IsValid() {
			return fmt.Errorf("invalid field: %s", field)
		}
		val := srcField.Interface()
		if err := st.db.UpdateField(src, field, val); err != nil {
			return err
		}
	}

	return nil
}

func (st sourcesBackend) Save(src *sources.Source) error {
	err := st.db.Save(src)
	if errors.Is(err, storm.ErrAlreadyExists) {
		return fberrors.ErrExist
	}
	return err
}

func (st sourcesBackend) DeleteByID(id uint) error {
	return st.db.DeleteStruct(&sources.Source{ID: id})
}
