package sources

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// NormalizePath cleans a user-supplied source path the way an admin expects
// when pasting from a terminal, file manager, or chat message.
//
// It handles the common paste mistakes that previously surfaced as a bare
// "source path must be absolute" (400) even though the path "looked" absolute:
//   - leading/trailing whitespace and newlines ("  /data/media\n")
//   - surrounding single or double quotes ('"/data/media"', "'/data/media'")
//   - leading ~ (~/media -> $HOME/media)
//   - $VAR / ${VAR} environment expansion ($HOME/media)
//
// On success it returns filepath.Clean(path). When the result is still not
// absolute it returns an error that echoes the received value plus an example,
// so the UI can show something actionable instead of a generic 400.
func NormalizePath(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("source path is required: use an absolute path such as %q", "/data/media")
	}

	// Strip one layer of matching surrounding quotes left by copy-paste from
	// terminals/docs: '"/data/media"' -> /data/media.
	if len(trimmed) >= 2 {
		first, last := trimmed[0], trimmed[len(trimmed)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			trimmed = strings.TrimSpace(trimmed[1 : len(trimmed)-1])
		}
	}
	if trimmed == "" {
		return "", fmt.Errorf("source path is required: use an absolute path such as %q", "/data/media")
	}

	// Expand a leading ~ to the server's home directory. filepath.Clean alone
	// leaves "~/x" untouched and IsAbs reports false, which confused users.
	if trimmed == "~" || strings.HasPrefix(trimmed, "~/") {
		home, err := os.UserHomeDir()
		if err != nil || home == "" {
			return "", fmt.Errorf("source path %q uses ~ but the server home directory is unknown: use an absolute path such as %q", raw, "/data/media")
		}
		if trimmed == "~" {
			trimmed = home
		} else {
			trimmed = filepath.Join(home, strings.TrimPrefix(trimmed, "~/"))
		}
	} else if strings.Contains(trimmed, "$") {
		// Expand $HOME-style variables pasted from docs/shell. Harmless when
		// no $ is present.
		trimmed = os.ExpandEnv(trimmed)
	}

	cleaned := filepath.Clean(trimmed)
	if !filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("source path must be absolute (e.g. %q): got %q", "/data/media", raw)
	}
	return cleaned, nil
}
