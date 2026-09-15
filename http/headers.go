//go:build !dev

package fbhttp

// global headers to append to every response
var globalHeaders = map[string]string{
	"Cache-Control":          "no-cache, no-store, must-revalidate",
	"X-Frame-Options":        "SAMEORIGIN",
	"Referrer-Policy":        "same-origin",
	"X-Content-Type-Options": "nosniff",
}
