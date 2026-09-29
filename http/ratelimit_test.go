package fbhttp

import "testing"

func TestAuthRateLimiterBlocksAfterLimit(t *testing.T) {
	t.Parallel()

	l := newAuthRateLimiter(2)
	if !l.allow("10.0.0.1") {
		t.Fatal("first request should be allowed")
	}
	if !l.allow("10.0.0.1") {
		t.Fatal("second request should be allowed")
	}
	if l.allow("10.0.0.1") {
		t.Error("third request should be rate limited")
	}
	if !l.allow("10.0.0.2") {
		t.Error("a different IP gets its own budget")
	}
}

func TestAuthRateLimiterDisabled(t *testing.T) {
	t.Parallel()

	l := newAuthRateLimiter(0)
	for i := 0; i < 100; i++ {
		if !l.allow("10.0.0.1") {
			t.Fatalf("request %d should be allowed when the limiter is disabled", i)
		}
	}
}

func TestAuthRateLimitFromEnv(t *testing.T) {
	cases := map[string]struct {
		env  string
		want int
	}{
		"unset uses default": {env: "", want: defaultAuthRateLimit},
		"positive overrides": {env: "1000", want: 1000},
		"zero disables":      {env: "0", want: 0},
		"invalid keeps default": {
			env:  "many",
			want: defaultAuthRateLimit,
		},
		"negative keeps default": {
			env:  "-5",
			want: defaultAuthRateLimit,
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Setenv("FB_AUTH_RATE_LIMIT", tc.env)
			if got := authRateLimitFromEnv(); got != tc.want {
				t.Errorf("authRateLimitFromEnv() = %d, want %d", got, tc.want)
			}
		})
	}
}
