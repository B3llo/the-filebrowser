package rules

import (
	"path/filepath"
	"regexp"
	"strings"
)

// Checker is a Rules checker.
type Checker interface {
	Check(path string) bool
}

// Rule is a allow/disallow rule.
type Rule struct {
	Regex  bool    `json:"regex"`
	Allow  bool    `json:"allow"`
	Path   string  `json:"path"`
	Regexp *Regexp `json:"regexp"`
}

// MatchHidden matches paths with a basename
// that begins with a dot.
func MatchHidden(path string) bool {
	return path != "" && strings.HasPrefix(filepath.Base(path), ".")
}

// Matches matches a path against a rule.
func (r *Rule) Matches(path string) bool {
	if r.Regex {
		return r.Regexp.MatchString(path)
	}

	if path == r.Path {
		return true
	}

	prefix := r.Path
	if prefix != "/" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	return strings.HasPrefix(path, prefix)
}

// Regexp is a wrapper to the native regexp type where we
// save the raw expression.
type Regexp struct {
	Raw    string `json:"raw"`
	regexp *regexp.Regexp
}

// MatchString checks if a string matches the regexp.
func (r *Regexp) MatchString(s string) bool {
	if r.regexp == nil {
		compiled, err := regexp.Compile(r.Raw)
		if err != nil {
			return false
		}
		r.regexp = compiled
	}

	return r.regexp.MatchString(s)
}

// Validate compiles the raw expression eagerly so invalid regexes are
// rejected at save time instead of panicking on first match.
func (r *Regexp) Validate() error {
	if r == nil {
		return nil
	}
	_, err := regexp.Compile(r.Raw)
	return err
}

// Validate checks rule fields, including regex compilation.
func (r *Rule) Validate() error {
	if r.Regex {
		if r.Regexp == nil {
			return nil
		}
		return r.Regexp.Validate()
	}
	return nil
}
