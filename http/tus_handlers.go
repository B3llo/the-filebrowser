package fbhttp

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/B3llo/the-filebrowser/files"
	"github.com/spf13/afero"
)

// tusWriteLocks serializes concurrent PATCH writes per destination file.
// O_WRONLY (without O_APPEND) honors Seek, so overlapping chunks must not
// interleave; the lock key is the resolved RealPath.
var tusWriteLocks sync.Map // string -> *sync.Mutex

func tusLockFor(realPath string) *sync.Mutex {
	mu, _ := tusWriteLocks.LoadOrStore(realPath, &sync.Mutex{})
	return mu.(*sync.Mutex)
}

// keepUploadActive periodically touches the cache entry to prevent eviction during transfer
func keepUploadActive(cache UploadCache, filePath string) func() {
	stop := make(chan bool)

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				cache.Touch(filePath)
			}
		}
	}()

	return func() {
		close(stop)
	}
}

func tusPostHandler(cache UploadCache) handleFunc {
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

		file, err := files.NewFileInfo(&files.FileOptions{
			Fs:         d.user.Fs,
			Path:       effPath,
			Modify:     d.user.Perm.Modify,
			Expand:     false,
			ReadHeader: d.server.TypeDetectionByHeader,
			Checker:    d,
		})
		switch {
		case errors.Is(err, afero.ErrFileNotFound):
			dirPath := filepath.Dir(effPath)
			if _, statErr := d.user.Fs.Stat(dirPath); os.IsNotExist(statErr) {
				if mkdirErr := d.user.Fs.MkdirAll(dirPath, d.settings.DirMode); mkdirErr != nil {
					return http.StatusInternalServerError, mkdirErr
				}
			}
		case err != nil:
			return errToStatus(err), err
		}

		fileFlags := os.O_CREATE | os.O_WRONLY

		// if file exists
		if file != nil {
			if file.IsDir {
				return http.StatusBadRequest, fmt.Errorf("cannot upload to a directory %s", file.RealPath())
			}

			// Existing files will remain untouched unless explicitly instructed to override
			if r.URL.Query().Get("override") != "true" {
				return http.StatusConflict, nil
			}

			// Permission for overwriting the file
			if !d.user.Perm.Modify {
				return http.StatusForbidden, nil
			}

			fileFlags |= os.O_TRUNC
		}

		openFile, err := d.user.Fs.OpenFile(effPath, fileFlags, d.settings.FileMode)
		if err != nil {
			return errToStatus(err), err
		}
		defer openFile.Close()

		file, err = files.NewFileInfo(&files.FileOptions{
			Fs:         d.user.Fs,
			Path:       effPath,
			Modify:     d.user.Perm.Modify,
			Expand:     false,
			ReadHeader: false,
			Checker:    d,
			Content:    false,
		})
		if err != nil {
			return errToStatus(err), err
		}

		uploadLength, err := getUploadLength(r)
		if err != nil || uploadLength < 0 {
			return http.StatusBadRequest, fmt.Errorf("invalid upload length: %w", err)
		}

		// Enables the user to utilize the PATCH endpoint for uploading file data
		cache.Register(file.RealPath(), uploadLength)

		basePath := "/" + strings.Trim(strings.TrimSpace(d.server.BaseURL), "/")
		if basePath == "/" {
			basePath = ""
		}

		w.Header().Set("Location", basePath+"/api/tus"+r.URL.EscapedPath())
		return http.StatusCreated, nil
	})
}

func tusHeadHandler(cache UploadCache) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		w.Header().Set("Cache-Control", "no-store")

		effPath := r.URL.Path
		if rel, gg, ok := tryGrantScope(effPath, d, false); ok {
			if status, allow := grantWriteStatus(gg); !allow {
				return status, nil
			}
			effPath = rel
		}

		if !d.user.Perm.Create || !d.Check(effPath) {
			return http.StatusForbidden, nil
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

		uploadLength, err := cache.GetLength(file.RealPath())
		if err != nil {
			return http.StatusNotFound, err
		}
		cache.Touch(file.RealPath())

		w.Header().Set("Upload-Offset", strconv.FormatInt(file.Size, 10))
		w.Header().Set("Upload-Length", strconv.FormatInt(uploadLength, 10))

		return http.StatusOK, nil
	})
}

func tusPatchHandler(cache UploadCache, fileCaches ...FileCache) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		effPath := r.URL.Path
		if rel, gg, ok := tryGrantScope(effPath, d, false); ok {
			if status, allow := grantWriteStatus(gg); !allow {
				return status, nil
			}
			effPath = rel
		}

		if !d.user.Perm.Create || !d.Check(effPath) {
			return http.StatusForbidden, nil
		}
		if r.Header.Get("Content-Type") != "application/offset+octet-stream" {
			return http.StatusUnsupportedMediaType, nil
		}

		uploadOffset, err := getUploadOffset(r)
		if err != nil {
			return http.StatusBadRequest, fmt.Errorf("invalid upload offset")
		}

		file, err := files.NewFileInfo(&files.FileOptions{
			Fs:         d.user.Fs,
			Path:       effPath,
			Modify:     d.user.Perm.Modify,
			Expand:     false,
			ReadHeader: d.server.TypeDetectionByHeader,
			Checker:    d,
		})

		switch {
		case errors.Is(err, afero.ErrFileNotFound):
			return http.StatusNotFound, nil
		case err != nil:
			return errToStatus(err), err
		}

		uploadLength, err := cache.GetLength(file.RealPath())
		if err != nil {
			return http.StatusNotFound, err
		}
		cache.Touch(file.RealPath())

		// Prevent the upload from being evicted during the transfer
		stop := keepUploadActive(cache, file.RealPath())
		defer stop()

		switch {
		case file.IsDir:
			return http.StatusBadRequest, fmt.Errorf("cannot upload to a directory %s", file.RealPath())
		case file.Size != uploadOffset:
			return http.StatusConflict, fmt.Errorf(
				"%s file size doesn't match the provided offset: %d",
				file.RealPath(),
				uploadOffset,
			)
		case uploadOffset > uploadLength:
			return http.StatusRequestEntityTooLarge, fmt.Errorf(
				"%s upload offset exceeds declared length: %d > %d",
				file.RealPath(),
				uploadOffset,
				uploadLength,
			)
		}

		// Serialize writers per destination: O_WRONLY honors Seek (O_APPEND
		// would ignore it), so concurrent chunks must not interleave.
		mu := tusLockFor(file.RealPath())
		mu.Lock()
		defer mu.Unlock()

		openFile, err := d.user.Fs.OpenFile(effPath, os.O_WRONLY, d.settings.FileMode)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("could not open file: %w", err)
		}
		defer openFile.Close()

		_, err = openFile.Seek(uploadOffset, io.SeekStart)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("could not seek file: %w", err)
		}

		// Never write past the declared length: cap the body at the
		// remaining bytes (+1 to detect an overflowing chunk).
		remaining := uploadLength - uploadOffset
		defer r.Body.Close()
		bytesWritten, err := io.Copy(openFile, io.LimitReader(r.Body, remaining+1))
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("could not write to file: %w", err)
		}

		// Sync the file to ensure all data is written to storage
		// to prevent file corruption.
		if err := openFile.Sync(); err != nil {
			return http.StatusInternalServerError, fmt.Errorf("could not sync file: %w", err)
		}

		newOffset := uploadOffset + bytesWritten
		if newOffset > uploadLength {
			// Best effort: discard the overflowing byte(s) so the file
			// never exceeds the declared length.
			_ = openFile.Truncate(uploadLength)
			_ = openFile.Sync()
			return http.StatusRequestEntityTooLarge, fmt.Errorf(
				"%s upload exceeds declared length: %d > %d",
				file.RealPath(),
				newOffset,
				uploadLength,
			)
		}
		w.Header().Set("Upload-Offset", strconv.FormatInt(newOffset, 10))

		if newOffset >= uploadLength {
			cache.Complete(file.RealPath())
			if len(fileCaches) > 0 && fileCaches[0] != nil {
				_ = delThumbs(r.Context(), fileCaches[0], file)
			}
			_ = d.RunHook(func() error { return nil }, "upload", effPath, "", d.user)
		}

		return http.StatusNoContent, nil
	})
}

func tusDeleteHandler(cache UploadCache) handleFunc {
	return withUser(func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
		cleaned := cleanUserPath(r.URL.Path)
		if cleaned == "/" || !d.user.Perm.Delete {
			return http.StatusForbidden, nil
		}

		effPath := cleaned
		if rel, gg, ok := tryGrantScope(effPath, d, false); ok {
			if status, allow := grantWriteStatus(gg); !allow {
				return status, nil
			}
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

		_, err = cache.GetLength(file.RealPath())
		if err != nil {
			return http.StatusNotFound, err
		}

		err = d.user.Fs.RemoveAll(effPath)
		if err != nil {
			return errToStatus(err), err
		}

		cache.Complete(file.RealPath())

		return http.StatusNoContent, nil
	})
}

func getUploadLength(r *http.Request) (int64, error) {
	uploadOffset, err := strconv.ParseInt(r.Header.Get("Upload-Length"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid upload length: %w", err)
	}
	return uploadOffset, nil
}

func getUploadOffset(r *http.Request) (int64, error) {
	uploadOffset, err := strconv.ParseInt(r.Header.Get("Upload-Offset"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid upload offset: %w", err)
	}
	return uploadOffset, nil
}
