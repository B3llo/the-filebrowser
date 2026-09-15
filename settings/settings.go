package settings

import (
	"crypto/rand"
	"fmt"
	"io/fs"
	"log"
	"strings"
	"time"

	"github.com/B3llo/the-filebrowser/rules"
)

const DefaultUsersHomeBasePath = "/users"
const DefaultLogoutPage = "/login"
const DefaultMinimumPasswordLength = 12
const DefaultFileMode = 0640
const DefaultDirMode = 0750

// DefaultTrashEnabled is the default for Trash.Enabled on fresh installs
// and on databases that predate the trash settings.
const DefaultTrashEnabled = true

// DefaultTrashRetentionDays is the default number of days trashed items are
// kept before permanent deletion. 0 means "keep forever".
const DefaultTrashRetentionDays = 30

// MaxTrashRetentionDays caps Trash.RetentionDays (10 years).
const MaxTrashRetentionDays = 3650

// AuthMethod describes an authentication method.
type AuthMethod string

// Trash holds the trash-bin settings of the application.
type Trash struct {
	// Enabled controls whether deleted files go to the per-source .Trash
	// directory (and whether the trash UI is shown). Defaults to true.
	Enabled bool `json:"enabled"`
	// RetentionDays is how many days trashed items are kept before they may
	// be permanently deleted. 0 means "keep forever". Valid range: 0-3650.
	RetentionDays int `json:"retentionDays"`
}

// Validate rejects out-of-range retention settings.
func (t Trash) Validate() error {
	if t.RetentionDays < 0 || t.RetentionDays > MaxTrashRetentionDays {
		return fmt.Errorf("trash retentionDays must be between 0 and %d", MaxTrashRetentionDays)
	}
	return nil
}

// Settings contain the main settings of the application.
type Settings struct {
	Key                   []byte              `json:"key"`
	Signup                bool                `json:"signup"`
	HideLoginButton       bool                `json:"hideLoginButton"`
	CreateUserDir         bool                `json:"createUserDir"`
	UserHomeBasePath      string              `json:"userHomeBasePath"`
	Defaults              UserDefaults        `json:"defaults"`
	AuthMethod            AuthMethod          `json:"authMethod"`
	LogoutPage            string              `json:"logoutPage"`
	Branding              Branding            `json:"branding"`
	Tus                   Tus                 `json:"tus"`
	Commands              map[string][]string `json:"commands"`
	Shell                 []string            `json:"shell"`
	Rules                 []rules.Rule        `json:"rules"`
	MinimumPasswordLength uint                `json:"minimumPasswordLength"`
	FileMode              fs.FileMode         `json:"fileMode"`
	DirMode               fs.FileMode         `json:"dirMode"`
	HideDotfiles          bool                `json:"hideDotfiles"`
	Trash                 Trash               `json:"trash"`
}

// GetRules implements rules.Provider.
func (s *Settings) GetRules() []rules.Rule {
	return s.Rules
}

// Server specific settings.
type Server struct {
	Root                   string `json:"root"`
	BaseURL                string `json:"baseURL"`
	Socket                 string `json:"socket"`
	TLSKey                 string `json:"tlsKey"`
	TLSCert                string `json:"tlsCert"`
	Port                   string `json:"port"`
	Address                string `json:"address"`
	Log                    string `json:"log"`
	EnableThumbnails       bool   `json:"enableThumbnails"`
	ResizePreview          bool   `json:"resizePreview"`
	EnableVideoThumbnails  bool   `json:"enableVideoThumbnails"`
	EnableExec             bool   `json:"enableExec"`
	TypeDetectionByHeader  bool   `json:"typeDetectionByHeader"`
	ImageResolutionCal     bool   `json:"imageResolutionCalculation"`
	AuthHook               string `json:"authHook"`
	TokenExpirationTime    string `json:"tokenExpirationTime"`
	FollowExternalSymlinks bool   `json:"followExternalSymlinks"`
}

// Clean cleans any variables that might need cleaning.
func (s *Server) Clean() {
	s.BaseURL = strings.TrimSuffix(s.BaseURL, "/")
}

func (s *Server) GetTokenExpirationTime(fallback time.Duration) time.Duration {
	if s.TokenExpirationTime == "" {
		return fallback
	}

	duration, err := time.ParseDuration(s.TokenExpirationTime)
	if err != nil {
		log.Printf("[WARN] Failed to parse tokenExpirationTime: %v", err)
		return fallback
	}
	return duration
}

// GenerateKey generates a key of 512 bits.
func GenerateKey() ([]byte, error) {
	b := make([]byte, 64)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}
