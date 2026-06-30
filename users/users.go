package users

import (
	"path/filepath"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/rules"
	"github.com/spf13/afero"
)

// ViewMode describes a view mode.
type ViewMode string

const (
	ListViewMode   ViewMode = "list"
	MosaicViewMode ViewMode = "mosaic"
)

// SourceRef assigns a source to a user. When a user has at least one SourceRef,
// only the listed sources are accessible; otherwise every defined source is
// accessible. Default marks the source selected on first visit.
type SourceRef struct {
	ID      uint `json:"id"`
	Default bool `json:"default"`
}

// StarredFile is a user-starred file/folder entry persisted per-user.
type StarredFile struct {
	URL       string `json:"url"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	StarredAt int64  `json:"starredAt"`
}

// RecentFile is a recently-visited file/folder entry persisted per-user.
type RecentFile struct {
	URL  string `json:"url"`
	Name string `json:"name"`
	Type string `json:"type"`
	At   int64  `json:"at"`
}

// User describes a user.
type User struct {
	ID                    uint              `storm:"id,increment" json:"id"`
	Username              string            `storm:"unique" json:"username"`
	DisplayName           string            `json:"displayName"`
	Password              string            `json:"password"`
	Scope                 string            `json:"scope"`
	Sources               []SourceRef       `json:"sources"`
	Locale                string            `json:"locale"`
	LockPassword          bool              `json:"lockPassword"`
	ViewMode              ViewMode          `json:"viewMode"`
	SingleClick           bool              `json:"singleClick"`
	RedirectAfterCopyMove bool              `json:"redirectAfterCopyMove"`
	Perm                  Permissions       `json:"perm"`
	Commands              []string          `json:"commands"`
	Sorting               files.Sorting     `json:"sorting"`
	Fs                    afero.Fs          `json:"-" yaml:"-"`
	Rules                 []rules.Rule      `json:"rules"`
	HideDotfiles          bool              `json:"hideDotfiles"`
	DateFormat            bool              `json:"dateFormat"`
	AceEditorTheme        string            `json:"aceEditorTheme"`
	FolderColors          map[string]string `json:"folderColors"`
	Theme                 string            `json:"theme"`
	Starred               []StarredFile     `json:"starred"`
	Recents               []RecentFile      `json:"recents"`
}

// GetRules implements rules.Provider.
func (u *User) GetRules() []rules.Rule {
	return u.Rules
}

var checkableFields = []string{
	"Username",
	"Password",
	"Scope",
	"Sources",
	"ViewMode",
	"Commands",
	"Sorting",
	"Rules",
}

// Clean cleans up a user and verifies if all its fields
// are alright to be saved.
func (u *User) Clean(baseScope string, followExternalSymlinks bool, fields ...string) error {
	if len(fields) == 0 {
		fields = checkableFields
	}

	for _, field := range fields {
		switch field {
		case "Username":
			if u.Username == "" {
				return fberrors.ErrEmptyUsername
			}
		case "Password":
			if u.Password == "" {
				return fberrors.ErrEmptyPassword
			}
		case "ViewMode":
			if u.ViewMode == "" {
				u.ViewMode = ListViewMode
			}
		case "Sources":
			if u.Sources == nil {
				u.Sources = []SourceRef{}
			}
		case "Commands":
			if u.Commands == nil {
				u.Commands = []string{}
			}
		case "Sorting":
			if u.Sorting.By == "" {
				u.Sorting.By = "name"
			}
		case "Rules":
			if u.Rules == nil {
				u.Rules = []rules.Rule{}
			}
		case "FolderColors":
			if u.FolderColors == nil {
				u.FolderColors = map[string]string{}
			}
		}
	}

	if u.Fs == nil {
		scope := u.Scope
		scope = filepath.Join(baseScope, filepath.Join("/", scope))
		u.Fs = files.NewFs(afero.NewOsFs(), scope, followExternalSymlinks)
	}

	return nil
}

// FullPath gets the full path for a user's relative path.
func (u *User) FullPath(path string) string {
	return afero.FullBaseFsPath(files.BasePath(u.Fs), path)
}
