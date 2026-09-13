package sources

import (
	"strings"
	"testing"
)

func TestNormalizePath(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr string
	}{
		{"/data/media", "/data/media", ""},
		{"  /data/media  \n", "/data/media", ""},
		{"\"/data/media\"", "/data/media", ""},
		{"'/data/media'", "/data/media", ""},
		{"  \"/data/media\"  ", "/data/media", ""},
		{"/tmp/../tmp/x//", "/tmp/x", ""},
		{"data/relative", "", "must be absolute"},
		{"", "", "is required"},
		{"   ", "", "is required"},
	}
	for _, tc := range cases {
		got, err := NormalizePath(tc.in)
		if tc.wantErr == "" && err != nil {
			t.Errorf("NormalizePath(%q) unexpected error: %v", tc.in, err)
			continue
		}
		if tc.wantErr != "" {
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("NormalizePath(%q) expected error containing %q, got %v", tc.in, tc.wantErr, err)
			}
			continue
		}
		if got != tc.want {
			t.Errorf("NormalizePath(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}

	// ~ and $HOME expand to absolute paths (value depends on test env home).
	for _, in := range []string{"~", "~/media", "$HOME/media", "${HOME}/media"} {
		got, err := NormalizePath(in)
		if err != nil {
			t.Errorf("NormalizePath(%q) unexpected error: %v", in, err)
			continue
		}
		if got == "" || got[0] != '/' {
			t.Errorf("NormalizePath(%q) = %q, want absolute path", in, got)
		}
	}
}
