package fbhttp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	fberrors "github.com/B3llo/the-filebrowser/errors"
	"github.com/B3llo/the-filebrowser/files"
	"github.com/B3llo/the-filebrowser/fileutils"
	"github.com/B3llo/the-filebrowser/grants"
	"github.com/mholt/archives"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/spf13/afero"
)

// trashDir is the per-source trash directory. It is hidden from normal
// listings, searches, recursive walks, size calculations and archives, but
// stays directly accessible so the trash UI keeps working.
const trashDir = "/.Trash"

// isTrashPath reports whether p is the trash directory itself or lives under it.
func isTrashPath(p string) bool {
	return p == trashDir || strings.HasPrefix(p, trashDir+"/")
}

// hideTrash reports whether trash entries must be filtered out of a view
// rooted at rootPath. Views rooted inside the trash itself (the trash UI)
// keep their contents.
func hideTrash(rootPath string) bool {
	return !isTrashPath(path.Clean("/" + strings.TrimPrefix(rootPath, "/")))
}

// filterTrashItems drops trash entries from an expanded directory listing,
// keeping the dir/file counters consistent. Listings of the trash itself
// are left untouched.
func filterTrashItems(file *files.FileInfo, rootPath string) {
	if file == nil || file.Listing == nil || !file.IsDir || !hideTrash(rootPath) {
		return
	}
	kept := make([]*files.FileInfo, 0, len(file.Items))
	dirs, filesCount := 0, 0
	for _, item := range file.Items {
		if isTrashPath(item.Path) {
			log.Printf("[DEBUG] hiding trash entry from listing: %s", item.Path)
			continue
		}
		kept = append(kept, item)
		if item.IsDir {
			dirs++
		} else {
			filesCount++
		}
	}
	file.Items = kept
	file.NumDirs = dirs
	file.NumFiles = filesCount
}

var resourceGetHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	effPath := r.URL.Path
	if rel, _, ok := tryGrantScope(effPath, d, false); ok {
		effPath = rel
	}

	file, err := files.NewFileInfo(&files.FileOptions{
		Fs:         d.user.Fs,
		Path:       effPath,
		Modify:     d.user.Perm.Modify,
		Expand:     true,
		ReadHeader: d.server.TypeDetectionByHeader,
		Checker:    d,
		Content:    d.user.Perm.Download,
	})
	if err != nil {
		return errToStatus(err), err
	}

	encoding := r.Header.Get("X-Encoding")
	if file.IsDir {
		file.Sorting = d.user.Sorting
		filterTrashItems(file, effPath)
		file.ApplySort()
		return renderJSON(w, r, file)
	} else if encoding == "true" {
		if !d.user.Perm.Download {
			return http.StatusAccepted, nil
		}
		if file.Type != "text" {
			return renderJSON(w, r, file)
		}

		f, err := d.user.Fs.Open(effPath)
		if err != nil {
			return errToStatus(err), err
		}
		defer f.Close()

		const maxInlineTextBytes = 10 << 20
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, err = io.CopyN(w, f, maxInlineTextBytes+1)
		if err != nil && !errors.Is(err, io.EOF) {
			return http.StatusInternalServerError, err
		}
		return 0, nil
	}

	if checksum := r.URL.Query().Get("checksum"); checksum != "" {
		err := file.ChecksumWithContext(r.Context(), checksum)
		if errors.Is(err, context.Canceled) {
			return 0, nil
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return http.StatusGatewayTimeout, err
		}
		if errors.Is(err, fberrors.ErrInvalidOption) {
			return http.StatusBadRequest, nil
		} else if err != nil {
			return http.StatusInternalServerError, err
		}

		// do not waste bandwidth if we just want the checksum
		file.Content = ""
	}

	return renderJSON(w, r, file)
})

func resourceDeleteHandler(fileCache FileCache) handleFunc {
	return withUser(func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
		cleaned := cleanUserPath(r.URL.Path)
		if cleaned == "/" || !d.user.Perm.Delete {
			return http.StatusForbidden, nil
		}

		effPath := cleaned
		var g *grants.Grant
		if rel, gg, ok := tryGrantScope(effPath, d, false); ok {
			if status, allow := grantWriteStatus(gg); !allow {
				return status, nil
			}
			g = gg
			effPath = rel
		}

		file, err := files.NewFileInfo(&files.FileOptions{
			Fs:         d.user.Fs,
			Path:       effPath,
			Modify:     d.user.Perm.Modify,
			Expand:     false,
			ReadHeader: d.server.TypeDetectionByHeader,
			Checker:    d,
		})
		if err != nil {
			return errToStatus(err), err
		}

		// Cascades are tracked in owner coordinates: in a grant context the
		// deleted path belongs to the grant owner's scope.
		cascadePath, cascadeUser := file.Path, d.user.ID
		if g != nil {
			cascadePath, cascadeUser = grantOwnerPath(g, effPath), g.OwnerID
		}

		err = d.store.Share.DeleteWithPathPrefix(cascadePath, cascadeUser)
		if err != nil {
			log.Printf("WARNING: Error(s) occurred while deleting associated shares with file: %s", err)
		}
		if err := d.store.Grants.DeleteWithPathPrefix(cascadePath, cascadeUser); err != nil {
			log.Printf("WARNING: Error(s) occurred while deleting associated grants with file: %s", err)
		}

		// delete thumbnails
		err = delThumbs(r.Context(), fileCache, file)
		if err != nil {
			return errToStatus(err), err
		}

		err = d.RunHook(func() error {
			return d.user.Fs.RemoveAll(effPath)
		}, "delete", effPath, "", d.user)

		if err != nil {
			return errToStatus(err), err
		}

		return http.StatusNoContent, nil
	})
}

func resourcePostHandler(fileCache FileCache) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		effPath := r.URL.Path
		if rel, gg, ok := tryGrantScope(effPath, d, true); ok {
			if status, allow := grantWriteStatus(gg); !allow {
				return status, nil
			}
			effPath = rel
		}

		if !d.user.Perm.Create || !d.Check(effPath) {
			return http.StatusForbidden, nil
		}

		// Directories creation on POST.
		if strings.HasSuffix(effPath, "/") {
			err := d.user.Fs.MkdirAll(effPath, d.settings.DirMode)
			return errToStatus(err), err
		}

		file, err := files.NewFileInfo(&files.FileOptions{
			Fs:         d.user.Fs,
			Path:       effPath,
			Modify:     d.user.Perm.Modify,
			Expand:     false,
			ReadHeader: d.server.TypeDetectionByHeader,
			Checker:    d,
		})
		if err == nil {
			if r.URL.Query().Get("override") != "true" {
				return http.StatusConflict, nil
			}

			// Permission for overwriting the file
			if !d.user.Perm.Modify {
				return http.StatusForbidden, nil
			}

			err = delThumbs(r.Context(), fileCache, file)
			if err != nil {
				return errToStatus(err), err
			}
		}

		err = d.RunHook(func() error {
			info, writeErr := writeFile(d.user.Fs, effPath, r.Body, d.settings.FileMode, d.settings.DirMode)
			if writeErr != nil {
				return writeErr
			}

			etag := fmt.Sprintf(`"%x%x"`, info.ModTime().UnixNano(), info.Size())
			w.Header().Set("ETag", etag)
			return nil
		}, "upload", effPath, "", d.user)

		if err != nil {
			_ = d.user.Fs.RemoveAll(effPath)
		}

		return errToStatus(err), err
	})
}

var resourcePutHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	effPath := r.URL.Path
	if rel, gg, ok := tryGrantScope(effPath, d, false); ok {
		if status, allow := grantWriteStatus(gg); !allow {
			return status, nil
		}
		effPath = rel
	}

	if !d.user.Perm.Modify || !d.Check(effPath) {
		return http.StatusForbidden, nil
	}

	// Only allow PUT for files.
	if strings.HasSuffix(effPath, "/") {
		return http.StatusMethodNotAllowed, nil
	}

	exists, err := afero.Exists(d.user.Fs, effPath)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if !exists {
		return http.StatusNotFound, nil
	}

	err = d.RunHook(func() error {
		info, writeErr := writeFile(d.user.Fs, effPath, r.Body, d.settings.FileMode, d.settings.DirMode)
		if writeErr != nil {
			return writeErr
		}

		etag := fmt.Sprintf(`"%x%x"`, info.ModTime().UnixNano(), info.Size())
		w.Header().Set("ETag", etag)
		return nil
	}, "save", effPath, "", d.user)

	return errToStatus(err), err
})

func resourcePatchHandler(fileCache FileCache) handleFunc {
	return withUser(func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
		src := path.Clean("/" + r.URL.Path)
		action := r.URL.Query().Get("action")

		// Grant scoping for the source (exact: these actions target existing paths).
		var srcGrant *grants.Grant
		if rel, gg, ok := tryGrantScope(src, d, false); ok {
			if status, allow := grantWriteStatus(gg); !allow {
				return status, nil
			}
			srcGrant = gg
			src = rel
			r.URL.Path = rel
		}

		// extract has no destination: it unpacks the archive alongside src.
		if action == "extract" {
			if src == "/" || !d.Check(src) {
				return http.StatusForbidden, nil
			}

			err := d.RunHook(func() error {
				return patchAction(r.Context(), action, src, "", d, fileCache)
			}, action, src, "", d.user)

			return errToStatus(err), err
		}

		dst := r.URL.Query().Get("destination")
		dst, err := url.QueryUnescape(dst)
		if err != nil {
			return errToStatus(err), err
		}
		dst = path.Clean("/" + dst)

		if srcGrant != nil {
			// In a grant context the destination must stay inside the same
			// grant (addressed in owner coordinates, like the source was).
			// Cross-scope copies are rejected.
			if dst != srcGrant.Path && !strings.HasPrefix(dst, srcGrant.Path+"/") {
				return http.StatusForbidden, nil
			}
			dst = strings.TrimPrefix(dst, srcGrant.Path)
			if dst == "" {
				dst = "/"
			}
			q := r.URL.Query()
			q.Set("destination", dst)
			r.URL.RawQuery = q.Encode()
		} else if _, statErr := d.user.Fs.Stat(dst); statErr != nil {
			if _, pErr := d.user.Fs.Stat(path.Dir(dst)); pErr != nil {
				// The destination only resolves inside a grant scope:
				// cross-scope copies are rejected.
				if gg, owner := matchGrant(dst, d); gg != nil {
					if _, serr := owner.Fs.Stat(dst); serr == nil {
						return http.StatusForbidden, nil
					}
					if _, serr := owner.Fs.Stat(path.Dir(dst)); serr == nil {
						return http.StatusForbidden, nil
					}
				}
			}
		}
		if !d.Check(src) || !d.Check(dst) {
			return http.StatusForbidden, nil
		}
		if dst == "/" || src == "/" {
			return http.StatusForbidden, nil
		}

		err = checkParent(src, dst)
		if err != nil {
			return http.StatusBadRequest, err
		}

		srcInfo, _ := d.user.Fs.Stat(src)
		dstInfo, _ := d.user.Fs.Stat(dst)
		same := os.SameFile(srcInfo, dstInfo)

		if action != "rename" || !same {
			override := r.URL.Query().Get("override") == "true"
			rename := r.URL.Query().Get("rename") == "true"
			if !override && !rename {
				if _, err = d.user.Fs.Stat(dst); err == nil {
					return http.StatusConflict, nil
				}
			}
			if rename {
				dst = addVersionSuffix(dst, d.user.Fs)
			}

			if override && !d.user.Perm.Modify {
				return http.StatusForbidden, nil
			}
		}

		err = d.RunHook(func() error {
			return patchAction(r.Context(), action, src, dst, d, fileCache)
		}, action, src, dst, d.user)

		return errToStatus(err), err
	})
}

func checkParent(src, dst string) error {
	rel, err := filepath.Rel(src, dst)
	if err != nil {
		return err
	}

	rel = filepath.ToSlash(rel)
	if !strings.HasPrefix(rel, "../") && rel != ".." && rel != "." {
		return fberrors.ErrSourceIsParent
	}

	return nil
}

func addVersionSuffix(source string, afs afero.Fs) string {
	counter := 1
	dir, name := path.Split(source)
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)

	for {
		if _, err := afs.Stat(source); err != nil {
			break
		}
		renamed := fmt.Sprintf("%s(%d)%s", base, counter, ext)
		source = path.Join(dir, renamed)
		counter++
	}

	return source
}

func writeFile(afs afero.Fs, dst string, in io.Reader, fileMode, dirMode fs.FileMode) (os.FileInfo, error) {
	dir, _ := path.Split(dst)
	err := afs.MkdirAll(dir, dirMode)
	if err != nil {
		return nil, err
	}

	file, err := afs.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileMode)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// P1: global upload ceiling (10GB). The HTTP layer cannot use
	// http.MaxBytesReader here (no w/r in scope), so enforce with a
	// LimitReader CopyN teto: read at most max+1 bytes and reject overflow.
	// No per-user quota exists on the user model, so the ceiling doubles
	// as the quota check.
	const maxUploadSize int64 = 10 << 30 // 10GB
	n, err := io.Copy(file, io.LimitReader(in, maxUploadSize+1))
	if err != nil {
		return nil, err
	}
	if n > maxUploadSize {
		return nil, fmt.Errorf("upload exceeds maximum size of %d bytes", maxUploadSize)
	}

	// Sync the file to ensure all data is written to storage.
	// to prevent file corruption.
	if err := file.Sync(); err != nil {
		return nil, err
	}

	// Gets the info about the file.
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	return info, nil
}

func delThumbs(ctx context.Context, fileCache FileCache, file *files.FileInfo) error {
	for _, previewSizeName := range PreviewSizeNames() {
		size, _ := ParsePreviewSize(previewSizeName)
		if err := fileCache.Delete(ctx, previewCacheKey(file, size)); err != nil {
			return err
		}
	}

	return nil
}

func patchAction(ctx context.Context, action, src, dst string, d *data, fileCache FileCache) error {
	switch action {
	case "copy":
		if !d.user.Perm.Create {
			return fberrors.ErrPermissionDenied
		}

		return fileutils.Copy(d.user.Fs, src, dst, d.settings.FileMode, d.settings.DirMode)
	case "extract":
		if !d.user.Perm.Modify && !d.user.Perm.Create {
			return fberrors.ErrPermissionDenied
		}
		return extractArchive(d.user.Fs, src, d.settings.FileMode, d.settings.DirMode)
	case "rename":
		if !d.user.Perm.Rename {
			return fberrors.ErrPermissionDenied
		}
		src = path.Clean("/" + src)
		dst = path.Clean("/" + dst)

		file, err := files.NewFileInfo(&files.FileOptions{
			Fs:         d.user.Fs,
			Path:       src,
			Modify:     d.user.Perm.Modify,
			Expand:     false,
			ReadHeader: false,
			Checker:    d,
		})
		if err != nil {
			return err
		}

		// delete thumbnails
		err = delThumbs(ctx, fileCache, file)
		if err != nil {
			return err
		}

		return fileutils.MoveFile(d.user.Fs, src, dst, d.settings.FileMode, d.settings.DirMode)
	default:
		return fmt.Errorf("unsupported action %s: %w", action, fberrors.ErrInvalidRequestParams)
	}
}

// extractArchive extracts a supported archive file (zip, tar, etc.) into the
// same directory the archive lives in. It uses mholt/archives for format
// detection and extraction.
func extractArchive(afs afero.Fs, src string, fileMode, dirMode os.FileMode) error {
	file, err := afs.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	// Identify the archive format from the filename and stream.
	// afero.File implements io.ReadAt, io.Seeker, io.Reader, so we can
	// use the file directly as a ReaderAtSeeker.
	format, _, err := archives.Identify(context.Background(), stat.Name(), file)
	if err != nil {
		return fmt.Errorf("unsupported archive format: %w", err)
	}

	extractor, ok := format.(archives.Extractor)
	if !ok {
		return fmt.Errorf("format does not support extraction")
	}

	file.Seek(0, io.SeekStart) //nolint: errcheck

	dstDir := path.Clean(path.Dir(src))

	return extractor.Extract(context.Background(), file, func(_ context.Context, info archives.FileInfo) error {
		if info.LinkTarget != "" || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing to extract link entry: %q", info.NameInArchive)
		}

		raw := info.NameInArchive
		if strings.HasPrefix(raw, "/") {
			return fmt.Errorf("illegal file path in archive: %q", raw)
		}
		for _, seg := range strings.Split(raw, "/") {
			if seg == ".." {
				return fmt.Errorf("illegal file path in archive: %q", raw)
			}
		}

		name := path.Clean(raw)
		if name == "." || name == "" {
			return nil
		}

		destPath := path.Join(dstDir, name)
		if rel, err := filepath.Rel(dstDir, destPath); err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return fmt.Errorf("illegal file path in archive: %q", raw)
		}

		if info.IsDir() {
			return afs.MkdirAll(destPath, dirMode)
		}

		parentDir := path.Dir(destPath)
		if err := afs.MkdirAll(parentDir, dirMode); err != nil {
			return err
		}

		srcContent, err := info.Open()
		if err != nil {
			return err
		}
		defer srcContent.Close()

		dstFile, err := afs.OpenFile(destPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileMode)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcContent)
		return err
	})
}

// Guards for expensive tree walks: cap per-request entries, bound wall time,
// and expose basic offset/limit pagination. Frontend callers that omit the
// query params keep receiving a plain JSON array (truncated at the cap).
const (
	maxRecursiveEntries  = 10000
	maxDirSizeEntries    = 10000
	recursiveWalkTimeout = 30 * time.Second
	dirSizeWalkTimeout   = 30 * time.Second
)

var errWalkPageFilled = errors.New("page filled")

// parseOffsetLimit parses ?offset=&limit= for the recursive listing.
// limit <= 0 means "no explicit limit" (caller caps to maxRecursiveEntries).
func parseOffsetLimit(r *http.Request) (offset, limit int, err error) {
	q := r.URL.Query()
	if s := q.Get("offset"); s != "" {
		offset, err = strconv.Atoi(s)
		if err != nil || offset < 0 {
			return 0, 0, fmt.Errorf("invalid offset: %w", fberrors.ErrInvalidRequestParams)
		}
	}
	if s := q.Get("limit"); s != "" {
		limit, err = strconv.Atoi(s)
		if err != nil || limit < 0 {
			return 0, 0, fmt.Errorf("invalid limit: %w", fberrors.ErrInvalidRequestParams)
		}
	}
	if limit <= 0 || limit > maxRecursiveEntries {
		limit = maxRecursiveEntries
	}
	return offset, limit, nil
}

// RecursiveEntry is a single file/directory entry returned by the recursive listing endpoint.
type RecursiveEntry struct {
	Path    string    `json:"path"`
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modified"`
	IsDir   bool      `json:"isDir"`
}

// resourceGetRecursiveHandler returns a flat list of every file and directory
// under the requested path, walking the tree recursively on the server side
// so the client only needs a single HTTP call.
var resourceGetRecursiveHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	rootPath := r.URL.Path
	if rootPath == "" {
		rootPath = "/"
	}
	if rel, _, ok := tryGrantScope(rootPath, d, false); ok {
		rootPath = rel
	}

	// Make sure the root itself exists and is a directory.
	info, err := d.user.Fs.Stat(rootPath)
	if err != nil {
		return errToStatus(err), err
	}
	if !info.IsDir() {
		return http.StatusBadRequest, fmt.Errorf("path is not a directory")
	}

	offset, limit, err := parseOffsetLimit(r)
	if err != nil {
		return http.StatusBadRequest, err
	}
	paginated := r.URL.Query().Get("offset") != "" || r.URL.Query().Get("limit") != ""

	// Timeout per request + abort on client disconnect.
	ctx, cancel := context.WithTimeout(r.Context(), recursiveWalkTimeout)
	defer cancel()

	entries := make([]RecursiveEntry, 0, min(limit, 1024))

	var matched int64
	// The flat trash view walks the trash itself; every other recursive
	// listing skips it.
	skipTrash := hideTrash(rootPath)
	err = afero.Walk(d.user.Fs, rootPath, func(fPath string, info os.FileInfo, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err != nil {
			return nil // skip entries we cannot read
		}

		// Skip the root directory itself.
		if fPath == rootPath {
			return nil
		}

		if skipTrash && isTrashPath(fPath) {
			log.Printf("[DEBUG] skipping trash path in recursive listing: %s", fPath)
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Respect user rules.
		if !d.Check(fPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if matched < int64(offset) {
			matched++
			return nil
		}
		matched++

		if len(entries) >= limit {
			return errWalkPageFilled
		}

		entries = append(entries, RecursiveEntry{
			Path:    fPath,
			Name:    info.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
		})
		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, errWalkPageFilled):
			// Page filled: not an error, stop the walk early.
		case errors.Is(err, context.DeadlineExceeded) || ctx.Err() == context.DeadlineExceeded:
			return http.StatusGatewayTimeout, fmt.Errorf("recursive listing timed out")
		case errors.Is(err, context.Canceled) || ctx.Err() == context.Canceled:
			return 0, nil
		default:
			return http.StatusInternalServerError, err
		}
	}
	if ctx.Err() == context.DeadlineExceeded {
		return http.StatusGatewayTimeout, fmt.Errorf("recursive listing timed out")
	}
	if ctx.Err() == context.Canceled {
		return 0, nil
	}

	if !paginated && len(entries) >= maxRecursiveEntries {
		w.Header().Set("X-Truncated", "true")
		log.Printf("WARNING: recursive listing of %q truncated at %d entries", rootPath, maxRecursiveEntries)
	}

	return renderJSON(w, r, entries)
})

type DiskUsageResponse struct {
	Total uint64 `json:"total"`
	Used  uint64 `json:"used"`
}

var diskUsage = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	file, err := files.NewFileInfo(&files.FileOptions{
		Fs:         d.user.Fs,
		Path:       r.URL.Path,
		Modify:     d.user.Perm.Modify,
		Expand:     false,
		ReadHeader: false,
		Checker:    d,
		Content:    false,
	})
	if err != nil {
		return errToStatus(err), err
	}
	fPath := file.RealPath()
	if !file.IsDir {
		return renderJSON(w, r, &DiskUsageResponse{
			Total: 0,
			Used:  0,
		})
	}

	usage, err := disk.UsageWithContext(r.Context(), fPath)
	if err != nil {
		return errToStatus(err), err
	}
	return renderJSON(w, r, &DiskUsageResponse{
		Total: usage.Total,
		Used:  usage.Used,
	})
})

// DirSizeResponse is the JSON payload for the directory size endpoint.
type DirSizeResponse struct {
	Size int64 `json:"size"`
}

// resourceDirSizeHandler walks the directory tree rooted at the request path
// and returns the total recursive size of all files within it.
var resourceDirSizeHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	rootPath := r.URL.Path
	if rootPath == "" {
		rootPath = "/"
	}
	if rel, _, ok := tryGrantScope(rootPath, d, false); ok {
		rootPath = rel
	}

	info, err := d.user.Fs.Stat(rootPath)
	if err != nil {
		return errToStatus(err), err
	}
	if !info.IsDir() {
		return http.StatusBadRequest, fmt.Errorf("path is not a directory")
	}

	// Timeout per request + abort on client disconnect.
	ctx, cancel := context.WithTimeout(r.Context(), dirSizeWalkTimeout)
	defer cancel()

	var totalSize int64
	var visited int64

	// Dir sizes exclude the trash, unless the walk is rooted inside it.
	skipTrash := hideTrash(rootPath)
	err = afero.Walk(d.user.Fs, rootPath, func(fPath string, info os.FileInfo, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err != nil {
			return nil // skip entries we cannot read
		}

		if skipTrash && isTrashPath(fPath) && fPath != rootPath {
			log.Printf("[DEBUG] skipping trash path in dirsize: %s", fPath)
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.Check(fPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.IsDir() {
			visited++
			if visited > maxDirSizeEntries {
				return fmt.Errorf("directory too large (>%d entries): %w", maxDirSizeEntries, fberrors.ErrInvalidRequestParams)
			}
			totalSize += info.Size()
		}
		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded) || ctx.Err() == context.DeadlineExceeded:
			return http.StatusGatewayTimeout, fmt.Errorf("dirsize timed out")
		case errors.Is(err, context.Canceled) || ctx.Err() == context.Canceled:
			return 0, nil
		}
		if errors.Is(err, fberrors.ErrInvalidRequestParams) {
			return http.StatusRequestEntityTooLarge, err
		}
		return http.StatusInternalServerError, err
	}
	if ctx.Err() == context.DeadlineExceeded {
		return http.StatusGatewayTimeout, fmt.Errorf("dirsize timed out")
	}
	if ctx.Err() == context.Canceled {
		return 0, nil
	}

	return renderJSON(w, r, &DirSizeResponse{Size: totalSize})
})
