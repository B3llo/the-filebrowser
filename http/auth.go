package fbhttp

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang-jwt/jwt/v5/request"

	fbAuth "github.com/B3llo/the-filebrowser/auth"
	fberrors "github.com/B3llo/the-filebrowser/errors"
	"github.com/B3llo/the-filebrowser/settings"
	"github.com/B3llo/the-filebrowser/users"
)

const (
	DefaultTokenExpirationTime = time.Hour * 2

	// Short-lived access sessions (P1). DefaultTokenExpirationTime is kept
	// for backwards compatibility, but any effective expiration above
	// maxTokenTTL is clamped down to it.
	accessTokenTTL = time.Minute * 10
	maxTokenTTL    = time.Minute * 15

	// renewMaxAge bounds the refresh window: a token (even expired) can
	// only be renewed within 7 days of its IssuedAt.
	renewMaxAge = time.Hour * 24 * 7

	authCookieName = "auth"

	maxAuthBodySize = 1 << 20 // 1 MiB
)

type userInfo struct {
	ID                       uint              `json:"id"`
	Locale                   string            `json:"locale"`
	ViewMode                 users.ViewMode    `json:"viewMode"`
	SingleClick              bool              `json:"singleClick"`
	RedirectAfterCopyMove    bool              `json:"redirectAfterCopyMove"`
	Perm                     users.Permissions `json:"perm"`
	Commands                 []string          `json:"commands"`
	LockPassword             bool              `json:"lockPassword"`
	HideDotfiles             bool              `json:"hideDotfiles"`
	PreferHighQualityPreview bool              `json:"preferHighQualityPreview"`
	DateFormat               bool              `json:"dateFormat"`
	Username                 string            `json:"username"`
	DisplayName              string            `json:"displayName"`
	AceEditorTheme           string            `json:"aceEditorTheme"`
	FolderColors             map[string]string `json:"folderColors"`
	Theme                    string            `json:"theme"`
	Avatar                   string            `json:"avatar"`
}

type authToken struct {
	User userInfo `json:"user"`
	jwt.RegisteredClaims
}

type extractor []string

func (e extractor) ExtractToken(r *http.Request) (string, error) {
	token, _ := request.HeaderExtractor{"X-Auth"}.ExtractToken(r)

	// Checks if the token isn't empty and if it contains two dots.
	// The former prevents incompatibility with URLs that previously
	// used basic auth.
	if token != "" && strings.Count(token, ".") == 2 {
		return token, nil
	}

	// HttpOnly refresh cookie. Read on every method (not just GET) so
	// POST /api/renew and /api/logout work after a page reload, when the
	// in-memory JWT is gone and only the cookie remains.
	cookie, _ := r.Cookie(authCookieName)
	if cookie != nil && strings.Count(cookie.Value, ".") == 2 {
		return cookie.Value, nil
	}

	return "", request.ErrNoTokenInRequest
}

func renewableErr(err error, d *data) bool {
	if d.settings.AuthMethod != fbAuth.MethodProxyAuth || err == nil {
		return false
	}

	if d.settings.LogoutPage == settings.DefaultLogoutPage {
		return false
	}

	if !errors.Is(err, jwt.ErrTokenExpired) {
		return false
	}

	return true
}

func withUser(fn handleFunc) handleFunc {
	return func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		keyFunc := func(_ *jwt.Token) (interface{}, error) {
			return d.settings.Key, nil
		}

		var tk authToken
		p := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
		token, err := request.ParseFromRequest(r, &extractor{}, keyFunc, request.WithClaims(&tk), request.WithParser(p))
		if (err != nil || !token.Valid) && !renewableErr(err, d) {
			return http.StatusUnauthorized, nil
		}

		// Revoked (logged out / rotated) tokens are rejected outright, not
		// just flagged for renewal.
		if tk.IssuedAt != nil && tk.IssuedAt.Unix() < d.store.Users.LastUpdate(tk.User.ID) {
			return http.StatusUnauthorized, nil
		}

		expiresSoon := tk.ExpiresAt != nil && time.Until(tk.ExpiresAt.Time) < time.Minute*5
		if expiresSoon {
			w.Header().Add("X-Renew-Token", "true")
		}

		d.user, err = d.store.Users.Get(d.server.Root, d.server.FollowExternalSymlinks, tk.User.ID)
		if err != nil {
			// Unknown user (deleted): unauthenticated, not a server error.
			return http.StatusUnauthorized, nil
		}
		resolveActiveSource(r, d)
		return fn(w, r, d)
	}
}

func withAdmin(fn handleFunc) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Admin {
			return http.StatusForbidden, nil
		}

		return fn(w, r, d)
	})
}

func loginHandler(tokenExpireTime time.Duration) handleFunc {
	return func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxAuthBodySize)
		}

		auther, err := d.store.Auth.Get(d.settings.AuthMethod)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		user, err := auther.Auth(r, d.store.Users, d.settings, d.server)
		switch {
		case errors.Is(err, os.ErrPermission):
			return http.StatusForbidden, nil
		case err != nil:
			return http.StatusInternalServerError, err
		}

		return printToken(w, r, d, user, tokenExpireTime)
	}
}

type signupBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var signupHandler = func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.settings.Signup {
		return http.StatusMethodNotAllowed, nil
	}

	if r.Body == nil {
		return http.StatusBadRequest, nil
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxAuthBodySize)

	info := &signupBody{}
	err := json.NewDecoder(r.Body).Decode(info)
	if err != nil {
		return http.StatusBadRequest, err
	}

	if info.Password == "" || info.Username == "" {
		return http.StatusBadRequest, nil
	}

	user := &users.User{
		Username: info.Username,
	}

	d.settings.Defaults.Apply(user)

	// Users signed up via the signup handler should never become admins, even
	// if that is the default permission.
	user.Perm.Admin = false

	// Self-registered users should not inherit execution capabilities from
	// default settings, regardless of what the administrator has configured
	// as the default. Execution rights must be explicitly granted by an admin.
	user.Perm.Execute = false
	user.Commands = []string{}

	pwd, err := users.ValidateAndHashPwd(info.Password, d.settings.MinimumPasswordLength)
	if err != nil {
		return http.StatusBadRequest, err
	}

	user.Password = pwd
	if d.settings.CreateUserDir {
		user.Scope = ""
	}

	userHome, err := d.settings.MakeUserDir(user.Username, user.Scope, d.server.Root)
	if err != nil {
		log.Printf("create user: failed to mkdir user home dir: [%s]", userHome)
		return http.StatusInternalServerError, err
	}
	user.Scope = userHome
	log.Printf("new user: %s, home dir: [%s].", user.Username, userHome)

	err = d.store.Users.Save(user)
	if errors.Is(err, fberrors.ErrExist) {
		return http.StatusConflict, err
	} else if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func renewHandler(tokenExpireTime time.Duration) handleFunc {
	return func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		raw, err := (&extractor{}).ExtractToken(r)
		if err != nil || raw == "" {
			return http.StatusUnauthorized, nil
		}

		// Verify the signature but allow an expired access token: the
		// refresh window is bounded by IssuedAt (renewMaxAge), not by
		// the short access expiry.
		var tk authToken
		keyFunc := func(_ *jwt.Token) (interface{}, error) {
			return d.settings.Key, nil
		}
		p := jwt.NewParser(
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
			jwt.WithoutClaimsValidation(),
		)
		token, err := p.ParseWithClaims(raw, &tk, keyFunc)
		if err != nil || !token.Valid {
			return http.StatusUnauthorized, nil
		}

		if tk.IssuedAt == nil {
			return http.StatusUnauthorized, nil
		}
		now := time.Now()
		issued := tk.IssuedAt.Time
		if issued.After(now.Add(time.Minute)) {
			return http.StatusUnauthorized, nil
		}
		if now.Sub(issued) > renewMaxAge {
			return http.StatusUnauthorized, nil
		}
		if tk.IssuedAt.Unix() < d.store.Users.LastUpdate(tk.User.ID) {
			return http.StatusUnauthorized, nil
		}

		user, err := d.store.Users.Get(d.server.Root, d.server.FollowExternalSymlinks, tk.User.ID)
		if err != nil {
			return http.StatusUnauthorized, nil
		}
		d.user = user
		resolveActiveSource(r, d)

		w.Header().Set("X-Renew-Token", "false")
		return printToken(w, r, d, d.user, tokenExpireTime)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	// Best effort: invalidate the presenting token by bumping the user's
	// LastUpdate, so any copy of it fails the IssuedAt check in withUser
	// and renewHandler. Signature is still verified; expiry is ignored so
	// logout works even with an expired access token.
	if raw, err := (&extractor{}).ExtractToken(r); err == nil && raw != "" {
		var tk authToken
		keyFunc := func(_ *jwt.Token) (interface{}, error) {
			return d.settings.Key, nil
		}
		p := jwt.NewParser(
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
			jwt.WithoutClaimsValidation(),
		)
		if token, err := p.ParseWithClaims(raw, &tk, keyFunc); err == nil && token.Valid && tk.User.ID != 0 {
			if user, err := d.store.Users.Get(d.server.Root, d.server.FollowExternalSymlinks, tk.User.ID); err == nil {
				_ = d.store.Users.Update(user)
				// LastUpdate has 1s resolution; guarantee it lands strictly
				// after IssuedAt so the token is actually revoked.
				if tk.IssuedAt != nil && d.store.Users.LastUpdate(user.ID) <= tk.IssuedAt.Unix() {
					if wait := time.Until(tk.IssuedAt.Time.Add(time.Second + 50*time.Millisecond)); wait > 0 && wait < 2*time.Second {
						time.Sleep(wait)
					} else {
						time.Sleep(time.Second)
					}
					_ = d.store.Users.Update(user)
				}
			}
		}
	}

	clearAuthCookie(w)
	return http.StatusOK, nil
}

// clampTokenExpiration keeps DefaultTokenExpirationTime for compatibility
// but caps any effective expiration above maxTokenTTL (15min); non-positive
// values fall back to the 10min access TTL.
func clampTokenExpiration(d time.Duration) time.Duration {
	if d <= 0 {
		return accessTokenTTL
	}
	if d > maxTokenTTL {
		return maxTokenTTL
	}
	return d
}

func setAuthCookie(w http.ResponseWriter, signed string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    signed,
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		Expires:  time.Now().Add(ttl),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

func clearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

func printToken(w http.ResponseWriter, _ *http.Request, d *data, user *users.User, tokenExpirationTime time.Duration) (int, error) {
	tokenExpirationTime = clampTokenExpiration(tokenExpirationTime)
	claims := &authToken{
		User: userInfo{
			ID:                       user.ID,
			Locale:                   user.Locale,
			ViewMode:                 user.ViewMode,
			SingleClick:              user.SingleClick,
			RedirectAfterCopyMove:    user.RedirectAfterCopyMove,
			Perm:                     user.Perm,
			LockPassword:             user.LockPassword,
			Commands:                 user.Commands,
			HideDotfiles:             user.HideDotfiles,
			PreferHighQualityPreview: user.PreferHighQualityPreview,
			DateFormat:               user.DateFormat,
			Username:                 user.Username,
			DisplayName:              user.DisplayName,
			AceEditorTheme:           user.AceEditorTheme,
			FolderColors:             user.FolderColors,
			Theme:                    user.Theme,
			Avatar:                   user.Avatar,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpirationTime)),
			Issuer:    "File Browser",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(d.settings.Key)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	setAuthCookie(w, signed, tokenExpirationTime)
	w.Header().Set("Content-Type", "text/plain")
	if _, err := w.Write([]byte(signed)); err != nil {
		return http.StatusInternalServerError, err
	}
	return 0, nil
}
