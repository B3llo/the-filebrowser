package fbhttp

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/B3llo/the-filebrowser/files"
	"github.com/B3llo/the-filebrowser/settings"
	"github.com/B3llo/the-filebrowser/users"
)

// doCookieRawRequest performs a download exactly like a browser navigation
// does: no X-Auth header, only the HttpOnly auth cookie.
func doCookieRawRequest(t *testing.T, fx *grantFixture, token string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/api/raw/docs/report.txt", http.NoBody)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	rec := httptest.NewRecorder()
	handle(rawHandler, "/api/raw", fx.store, &settings.Server{}).ServeHTTP(rec, req)
	return rec
}

// Regression test: persisting a user preference (sorting, view mode, theme,
// folder colors) through PUT /api/users/:id used to bump LastUpdate and
// revoke the caller's own token, so the next request — including every
// cookie-authenticated download — returned 401 and logged the user out.
func TestDownloadCookieSurvivesPreferenceUpdate(t *testing.T) {
	t.Parallel()

	fx := newGrantFixture(t)
	token := signGrantToken(t, 1, "admin", fx.admin, fx.key)

	if rec := doCookieRawRequest(t, fx, token); rec.Code != http.StatusOK {
		t.Fatalf("download before update = %d, want 200", rec.Code)
	}

	if err := fx.store.Users.Update(&users.User{
		ID:      1,
		Sorting: files.Sorting{By: "size", Asc: true},
	}, "Sorting"); err != nil {
		t.Fatalf("preference update: %v", err)
	}
	if rec := doCookieRawRequest(t, fx, token); rec.Code != http.StatusOK {
		t.Errorf("download after preference update = %d, want 200", rec.Code)
	}

	// Credential changes must still revoke every outstanding token.
	if err := fx.store.Users.Update(&users.User{ID: 1, Password: "new-hash"}, "Password"); err != nil {
		t.Fatalf("password update: %v", err)
	}
	if rec := doCookieRawRequest(t, fx, token); rec.Code != http.StatusUnauthorized {
		t.Errorf("download after password update = %d, want 401", rec.Code)
	}
}

// The auth cookie is the refresh credential: it must outlive the access
// token so a reload or a download can still renew after the JWT expires.
func TestAuthCookieOutlivesAccessToken(t *testing.T) {
	t.Parallel()

	fx := newGrantFixture(t)
	u, err := fx.store.Users.Get("", false, uint(1))
	if err != nil {
		t.Fatalf("failed to load user: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/login", http.NoBody)
	if _, err := printToken(rec, req, &data{settings: &settings.Settings{Key: fx.key}}, u, time.Hour); err != nil {
		t.Fatalf("printToken: %v", err)
	}

	var cookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == authCookieName {
			cookie = c
		}
	}
	if cookie == nil {
		t.Fatal("auth cookie was not set")
	}
	if want := int(refreshCookieTTL.Seconds()); cookie.MaxAge != want {
		t.Errorf("cookie MaxAge = %d, want %d", cookie.MaxAge, want)
	}

	var tk authToken
	if _, err := jwt.NewParser(jwt.WithoutClaimsValidation()).
		ParseWithClaims(rec.Body.String(), &tk, func(*jwt.Token) (interface{}, error) {
			return fx.key, nil
		}); err != nil {
		t.Fatalf("failed to parse issued token: %v", err)
	}
	if tk.ExpiresAt == nil || tk.IssuedAt == nil {
		t.Fatal("issued token is missing iat/exp")
	}
	if got := tk.ExpiresAt.Sub(tk.IssuedAt.Time); got != time.Hour {
		t.Errorf("access token TTL = %v, want 1h", got)
	}
	if refreshCookieTTL <= time.Hour {
		t.Errorf("refreshCookieTTL = %v, want it to outlive the access token", refreshCookieTTL)
	}
}

// WebKit refuses to store a Secure cookie from a plain-HTTP origin, so the
// cookie must only be marked Secure when the request actually came over TLS
// (directly or through a TLS-terminating proxy).
func TestAuthCookieSecureFlagFollowsScheme(t *testing.T) {
	t.Parallel()

	fx := newGrantFixture(t)
	u, err := fx.store.Users.Get("", false, uint(1))
	if err != nil {
		t.Fatalf("failed to load user: %v", err)
	}

	issue := func(t *testing.T, mutate func(*http.Request)) *http.Cookie {
		t.Helper()

		req := httptest.NewRequest(http.MethodPost, "/api/login", http.NoBody)
		mutate(req)

		rec := httptest.NewRecorder()
		if _, err := printToken(rec, req, &data{settings: &settings.Settings{Key: fx.key}}, u, time.Hour); err != nil {
			t.Fatalf("printToken: %v", err)
		}
		for _, c := range rec.Result().Cookies() {
			if c.Name == authCookieName {
				return c
			}
		}
		t.Fatal("auth cookie was not set")
		return nil
	}

	plain := issue(t, func(*http.Request) {})
	if plain.Secure {
		t.Error("plain-HTTP request got a Secure cookie; WebKit will drop it")
	}

	viaTLS := issue(t, func(r *http.Request) { r.TLS = &tls.ConnectionState{} })
	if !viaTLS.Secure {
		t.Error("TLS request should get a Secure cookie")
	}

	viaProxy := issue(t, func(r *http.Request) {
		r.Header.Set("X-Forwarded-Proto", "https")
	})
	if !viaProxy.Secure {
		t.Error("X-Forwarded-Proto=https should get a Secure cookie")
	}
}

func TestClampTokenExpiration(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		in   time.Duration
		want time.Duration
	}{
		"non-positive falls back to access TTL": {0, accessTokenTTL},
		"within bounds is kept":                 {time.Hour, time.Hour},
		"above max is clamped to max":           {30 * 24 * time.Hour, maxTokenTTL},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := clampTokenExpiration(tc.in); got != tc.want {
				t.Errorf("clampTokenExpiration(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func signSessionTokenAt(t *testing.T, id uint, username string, perm users.Permissions, key []byte, issuedAt time.Time, ttl time.Duration) string {
	t.Helper()

	claims := &authToken{
		User: userInfo{ID: id, Username: username, Perm: perm},
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(issuedAt.Add(ttl)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(key)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return signed
}

func doCookieRenew(t *testing.T, fx *grantFixture, token string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/renew", http.NoBody)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	rec := httptest.NewRecorder()
	handle(renewHandler(DefaultTokenExpirationTime), "", fx.store, &settings.Server{}).ServeHTTP(rec, req)
	return rec
}

// The refresh cookie may carry an access token that already expired: renew
// must accept it within renewMaxAge so reloads restore the session instead of
// forcing a login (and instead of leaving cookie-only downloads broken).
func TestRenewAcceptsExpiredTokenWithinWindow(t *testing.T) {
	t.Parallel()

	fx := newGrantFixture(t)

	expired := signSessionTokenAt(t, 1, "admin", fx.admin, fx.key, time.Now().Add(-24*time.Hour), time.Hour)
	if rec := doCookieRenew(t, fx, expired); rec.Code != http.StatusOK {
		t.Errorf("renew with 1-day-old expired token = %d, want 200", rec.Code)
	}

	stale := signSessionTokenAt(t, 1, "admin", fx.admin, fx.key, time.Now().Add(-renewMaxAge-time.Hour), time.Hour)
	if rec := doCookieRenew(t, fx, stale); rec.Code != http.StatusUnauthorized {
		t.Errorf("renew with token older than renewMaxAge = %d, want 401", rec.Code)
	}
}

func doCookieLogout(t *testing.T, fx *grantFixture, token string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/logout", http.NoBody)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	rec := httptest.NewRecorder()
	handle(logoutHandler, "", fx.store, &settings.Server{}).ServeHTTP(rec, req)
	return rec
}

// An idle tab's inactivity timer calls /api/logout with an already-expired
// token. That must clear the cookie but not bump LastUpdate: revoking there
// killed every fresh session in other tabs and devices.
func TestExpiredTokenLogoutDoesNotRevokeFreshSessions(t *testing.T) {
	t.Parallel()

	fx := newGrantFixture(t)
	fresh := signGrantToken(t, 1, "admin", fx.admin, fx.key)
	expired := signSessionTokenAt(t, 1, "admin", fx.admin, fx.key, time.Now().Add(-3*time.Hour), time.Hour)

	if rec := doCookieLogout(t, fx, expired); rec.Code != http.StatusOK {
		t.Fatalf("expired-token logout = %d, want 200", rec.Code)
	}
	if rec := doCookieRawRequest(t, fx, fresh); rec.Code != http.StatusOK {
		t.Errorf("fresh session after expired-token logout = %d, want 200", rec.Code)
	}

	// Explicit logout with a valid token must still revoke it.
	if rec := doCookieLogout(t, fx, fresh); rec.Code != http.StatusOK {
		t.Fatalf("valid-token logout = %d, want 200", rec.Code)
	}
	if rec := doCookieRawRequest(t, fx, fresh); rec.Code != http.StatusUnauthorized {
		t.Errorf("revoked session after valid-token logout = %d, want 401", rec.Code)
	}
}
