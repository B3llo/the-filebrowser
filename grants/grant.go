package grants

import (
	"path"
	"strconv"
	"strings"
	"time"

	fberrors "github.com/B3llo/the-filebrowser/errors"
)

// Role is the access level of a grant.
type Role string

const (
	RoleViewer Role = "viewer"
	RoleEditor Role = "editor"
)

// Grant gives GranteeID access to Path inside OwnerID's scope.
// Path is absolute within the owner's scope (e.g. "/projects/X").
// Expire is a unix timestamp; 0 means permanent.
type Grant struct {
	ID        uint   `storm:"id,increment" json:"id"`
	Path      string `storm:"index" json:"path"`
	OwnerID   uint   `storm:"index" json:"ownerID"`
	GranteeID uint   `storm:"index" json:"granteeID"`
	Role      Role   `json:"role"`
	CreatedBy uint   `json:"createdBy"`
	CreatedAt int64  `json:"createdAt"`
	Expire    int64  `json:"expire"`
}

// CreateBody is the JSON body accepted by POST /api/grants.
// Owner is optional and only honored for admins: it selects the user whose
// scope Path belongs to. Non-admins always create grants on their own scope.
type CreateBody struct {
	Path    string `json:"path"`
	Owner   string `json:"owner"`
	Grantee string `json:"grantee"`
	Role    Role   `json:"role"`
	Expires string `json:"expires"`
	Unit    string `json:"unit"`
}

// IsExpired reports whether the grant is expired at time now.
func (g *Grant) IsExpired(now int64) bool {
	return g.Expire != 0 && g.Expire <= now
}

// IsActive reports whether the grant can be used right now.
func (g *Grant) IsActive() bool {
	return !g.IsExpired(time.Now().Unix())
}

// Validate checks role, path and grantee/owner consistency.
// It normalizes Path in place.
func (g *Grant) Validate() error {
	if g.Role != RoleViewer && g.Role != RoleEditor {
		return fberrors.ErrInvalidRequestParams
	}

	p, err := NormalizePath(g.Path)
	if err != nil {
		return err
	}
	g.Path = p

	if g.OwnerID == 0 || g.GranteeID == 0 {
		return fberrors.ErrInvalidRequestParams
	}
	if g.OwnerID == g.GranteeID {
		return fberrors.ErrInvalidRequestParams
	}

	return nil
}

// NormalizePath cleans a grant path and rejects escapes.
func NormalizePath(p string) (string, error) {
	if p == "" {
		return "", fberrors.ErrInvalidRequestParams
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	cleaned := path.Clean(p)
	if cleaned == "." || !strings.HasPrefix(cleaned, "/") {
		return "", fberrors.ErrInvalidRequestParams
	}
	return cleaned, nil
}

// ExpireFromStrings parses the expires/unit pair used by the share API.
// Empty expires means permanent (0, nil). Mirrors http/share.go behavior:
// seconds, minutes, days, default hours.
func ExpireFromStrings(expires, unit string) (int64, error) {
	if expires == "" {
		return 0, nil
	}

	num, err := strconv.Atoi(expires)
	if err != nil {
		return 0, fberrors.ErrInvalidRequestParams
	}
	if num < 0 {
		return 0, fberrors.ErrInvalidRequestParams
	}

	var add time.Duration
	switch unit {
	case "seconds":
		add = time.Second * time.Duration(num)
	case "minutes":
		add = time.Minute * time.Duration(num)
	case "days":
		add = time.Hour * 24 * time.Duration(num)
	default:
		add = time.Hour * time.Duration(num)
	}

	return time.Now().Add(add).Unix(), nil
}
