//go:generate go-enum --sql --marshal --names --file $GOFILE
package fbhttp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/B3llo/the-filebrowser/files"
	"github.com/B3llo/the-filebrowser/img"
)

/*
ENUM(
thumb
big
)
*/
type PreviewSize int

type ImgService interface {
	FormatFromExtension(ext string) (img.Format, error)
	Resize(ctx context.Context, in io.Reader, width, height int, out io.Writer, options ...img.Option) error
}

type VideoService interface {
	Available() bool
	ExtractFrame(ctx context.Context, realPath string) ([]byte, error)
}

type FileCache interface {
	Store(ctx context.Context, key string, value []byte) error
	Load(ctx context.Context, key string) ([]byte, bool, error)
	Delete(ctx context.Context, key string) error
}

// previewExtensionSupported checks if a file extension is supported for preview
func previewExtensionSupported(ext string) bool {
	supportedExts := map[string]bool{
		// Code files
		".js": true, ".mjs": true, ".cjs": true,
		".ts": true, ".tsx": true, ".jsx": true,
		".py": true,
		".go": true,
		".rs": true,
		".java": true, ".kt": true,
		".c": true, ".cpp": true, ".h": true, ".hpp": true,
		".cs": true,
		".php": true,
		".rb": true,
		".swift": true,
		".dart": true,
		".lua": true,
		".pl": true, ".r": true,
		".scala": true, ".clj": true,
		".sh": true, ".bat": true, ".ps1": true,
		".html": true, ".htm": true, ".css": true, ".scss": true, ".sass": true,
		".json": true, ".xml": true, ".yaml": true, ".yml": true, ".toml": true,
		".sql": true, ".graphql": true,
		// Text files
		".txt": true, ".log": true, ".ini": true, ".cfg": true, ".conf": true,
		".md": true, ".markdown": true,
		".csv": true, ".tsv": true,
		// Documents
		".pdf": true,
	}
	return supportedExts[strings.ToLower(ext)]
}

func previewHandler(imgSvc ImgService, fileCache FileCache, videoSvc VideoService,
	enableThumbnails, resizePreview, enableVideoThumbnails bool) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Download {
			return http.StatusAccepted, nil
		}
		vars := mux.Vars(r)

		previewSize, err := ParsePreviewSize(vars["size"])
		if err != nil {
			return http.StatusBadRequest, err
		}

		file, err := files.NewFileInfo(&files.FileOptions{
			Fs:         d.user.Fs,
			Path:       "/" + vars["path"],
			Modify:     d.user.Perm.Modify,
			Expand:     true,
			ReadHeader: d.server.TypeDetectionByHeader,
			Checker:    d,
		})
		if err != nil {
			return errToStatus(err), err
		}

		setContentDisposition(w, r, file)

		switch file.Type {
		case "image":
			return handleImagePreview(w, r, imgSvc, fileCache, file, previewSize, enableThumbnails, resizePreview)
		case "video":
			return handleVideoPreview(w, r, imgSvc, fileCache, videoSvc, file, previewSize, enableVideoThumbnails)
		case "audio":
			return handleAudioPreview(w, r, file)
		case "pdf":
			return handlePDFPreview(w, r, file)
		case "text", "textImmutable":
			return handleTextPreview(w, r, file)
		default:
			// Check if it's a supported file extension for preview
			if previewExtensionSupported(file.Extension) {
				return handleTextPreview(w, r, file)
			}
			return rawFileHandler(w, r, file)
		}
	})
}

func handleImagePreview(
	w http.ResponseWriter,
	r *http.Request,
	imgSvc ImgService,
	fileCache FileCache,
	file *files.FileInfo,
	previewSize PreviewSize,
	enableThumbnails, resizePreview bool,
) (int, error) {
	if (previewSize == PreviewSizeBig && !resizePreview) ||
		(previewSize == PreviewSizeThumb && !enableThumbnails) {
		return rawFileHandler(w, r, file)
	}

	format, err := imgSvc.FormatFromExtension(file.Extension)
	// Unsupported extensions directly return the raw data
	if errors.Is(err, img.ErrUnsupportedFormat) || format == img.FormatGif {
		return rawFileHandler(w, r, file)
	}
	if err != nil {
		return errToStatus(err), err
	}

	cacheKey := previewCacheKey(file, previewSize)
	resizedImage, ok, err := fileCache.Load(r.Context(), cacheKey)
	if err != nil {
		return errToStatus(err), err
	}
	if !ok {
		resizedImage, err = createPreview(imgSvc, fileCache, file, previewSize)
		if err != nil {
			return errToStatus(err), err
		}
	}

	w.Header().Set("Cache-Control", "private")
	http.ServeContent(w, r, file.Name, file.ModTime, bytes.NewReader(resizedImage))

	return 0, nil
}

func createPreview(imgSvc ImgService, fileCache FileCache,
	file *files.FileInfo, previewSize PreviewSize) ([]byte, error) {
	fd, err := file.Fs.Open(file.Path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	var (
		width   int
		height  int
		options []img.Option
	)

	switch previewSize {
	case PreviewSizeBig:
		width = 1080
		height = 1080
		options = append(options, img.WithMode(img.ResizeModeFit), img.WithQuality(img.QualityMedium))
	case PreviewSizeThumb:
		width = 256
		height = 256
		options = append(options, img.WithMode(img.ResizeModeFill), img.WithQuality(img.QualityLow), img.WithFormat(img.FormatJpeg))
	default:
		return nil, img.ErrUnsupportedFormat
	}

	buf := &bytes.Buffer{}
	if err := imgSvc.Resize(context.Background(), fd, width, height, buf, options...); err != nil {
		return nil, err
	}

	go func() {
		cacheKey := previewCacheKey(file, previewSize)
		if err := fileCache.Store(context.Background(), cacheKey, buf.Bytes()); err != nil {
			fmt.Printf("failed to cache resized image: %v", err)
		}
	}()

	return buf.Bytes(), nil
}

// handleVideoPreview serves a static JPEG thumbnail (extracted via ffmpeg and
// cached) for PreviewSizeThumb requests when video thumbnails are available;
// any other case (Big/player, feature disabled, ffmpeg unavailable, or any
// generation error) falls back to streaming the raw video as before.
func handleVideoPreview(w http.ResponseWriter, r *http.Request, imgSvc ImgService, fileCache FileCache,
	videoSvc VideoService, file *files.FileInfo, previewSize PreviewSize, enableVideoThumbnails bool) (int, error) {
	if previewSize != PreviewSizeThumb || !enableVideoThumbnails || videoSvc == nil || !videoSvc.Available() {
		return serveRawVideo(w, r, file)
	}

	cacheKey := previewCacheKey(file, previewSize)
	thumb, ok, err := fileCache.Load(r.Context(), cacheKey)
	if err != nil {
		return errToStatus(err), err
	}
	if !ok {
		thumb, err = createVideoThumbnail(r.Context(), imgSvc, videoSvc, fileCache, file, previewSize)
		if err != nil {
			log.Printf("[WARN] video thumbnail generation failed for %q: %v — falling back to raw stream", file.RealPath(), err)
			return serveRawVideo(w, r, file)
		}
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "private")
	http.ServeContent(w, r, file.Name, file.ModTime, bytes.NewReader(thumb))
	return 0, nil
}

func createVideoThumbnail(ctx context.Context, imgSvc ImgService, videoSvc VideoService, fileCache FileCache,
	file *files.FileInfo, previewSize PreviewSize) ([]byte, error) {
	realPath := file.RealPath()
	if info, err := os.Stat(realPath); err != nil || info.IsDir() {
		return nil, fmt.Errorf("video source not addressable on local disk: %w", err)
	}

	jpegFrame, err := videoSvc.ExtractFrame(ctx, realPath)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	options := []img.Option{
		img.WithMode(img.ResizeModeFill), img.WithQuality(img.QualityLow), img.WithFormat(img.FormatJpeg),
	}
	if err := imgSvc.Resize(ctx, bytes.NewReader(jpegFrame), 256, 256, buf, options...); err != nil {
		return nil, err
	}

	go func() {
		cacheKey := previewCacheKey(file, previewSize)
		if err := fileCache.Store(context.Background(), cacheKey, buf.Bytes()); err != nil {
			log.Printf("[WARN] failed to cache video thumbnail: %v", err)
		}
	}()

	return buf.Bytes(), nil
}

// serveRawVideo streams a video file with range request support for seeking.
func serveRawVideo(w http.ResponseWriter, r *http.Request, file *files.FileInfo) (int, error) {
	fd, err := file.Fs.Open(file.Path)
	if err != nil {
		return errToStatus(err), err
	}
	defer fd.Close()

	// Get file info for size
	stat, err := fd.Stat()
	if err != nil {
		return errToStatus(err), err
	}

	fileSize := stat.Size()
	contentType := getVideoContentType(file.Extension)

	// Handle range requests for seeking
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		return handleRangeRequest(w, r, fd, fileSize, contentType, rangeHeader)
	}

	// Set headers for full video
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))
	w.Header().Set("Cache-Control", "private")

	// Serve the full video
	http.ServeContent(w, r, file.Name, file.ModTime, fd)
	return 0, nil
}

// handleAudioPreview serves audio files with range request support
func handleAudioPreview(w http.ResponseWriter, r *http.Request, file *files.FileInfo) (int, error) {
	fd, err := file.Fs.Open(file.Path)
	if err != nil {
		return errToStatus(err), err
	}
	defer fd.Close()

	stat, err := fd.Stat()
	if err != nil {
		return errToStatus(err), err
	}

	fileSize := stat.Size()
	contentType := getAudioContentType(file.Extension)

	// Handle range requests
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		return handleRangeRequest(w, r, fd, fileSize, contentType, rangeHeader)
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))
	w.Header().Set("Cache-Control", "private")

	http.ServeContent(w, r, file.Name, file.ModTime, fd)
	return 0, nil
}

// handlePDFPreview serves PDF files for inline viewing
func handlePDFPreview(w http.ResponseWriter, r *http.Request, file *files.FileInfo) (int, error) {
	fd, err := file.Fs.Open(file.Path)
	if err != nil {
		return errToStatus(err), err
	}
	defer fd.Close()

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename=\""+file.Name+"\"")
	w.Header().Set("Cache-Control", "private")

	http.ServeContent(w, r, file.Name, file.ModTime, fd)
	return 0, nil
}

// handleTextPreview serves text-based files (code, markdown, csv, etc.)
func handleTextPreview(w http.ResponseWriter, r *http.Request, file *files.FileInfo) (int, error) {
	fd, err := file.Fs.Open(file.Path)
	if err != nil {
		return errToStatus(err), err
	}
	defer fd.Close()

	stat, err := fd.Stat()
	if err != nil {
		return errToStatus(err), err
	}

	// Determine content type based on extension
	contentType := getTextContentType(file.Extension)

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "private")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))

	http.ServeContent(w, r, file.Name, file.ModTime, fd)
	return 0, nil
}

// handleRangeRequest handles HTTP range requests for media files
func handleRangeRequest(w http.ResponseWriter, _ *http.Request, fd io.ReadSeeker, fileSize int64, contentType, rangeHeader string) (int, error) {
	// Parse range header (e.g., "bytes=0-1023")
	rangeSpec := strings.TrimPrefix(rangeHeader, "bytes=")
	parts := strings.Split(rangeSpec, "-")
	if len(parts) != 2 {
		return http.StatusRequestedRangeNotSatisfiable, fmt.Errorf("invalid range header")
	}

	var start, end int64
	var err error

	if parts[0] == "" {
		// Suffix range: bytes=-500 (last 500 bytes)
		end = fileSize - 1
		start, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return http.StatusRequestedRangeNotSatisfiable, err
		}
		start = fileSize - start
	} else {
		start, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return http.StatusRequestedRangeNotSatisfiable, err
		}
		if parts[1] == "" {
			end = fileSize - 1
		} else {
			end, err = strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return http.StatusRequestedRangeNotSatisfiable, err
			}
		}
	}

	if start < 0 || end >= fileSize || start > end {
		return http.StatusRequestedRangeNotSatisfiable, fmt.Errorf("range not satisfiable")
	}

	// Seek to the start position
	_, err = fd.Seek(start, io.SeekStart)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Calculate content length
	contentLength := end - start + 1

	// Set headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", contentLength))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Cache-Control", "private")
	w.WriteHeader(http.StatusPartialContent)

	// Serve the range
	if _, err := io.CopyN(w, fd, contentLength); err != nil {
		return http.StatusInternalServerError, err
	}
	return 0, nil
}

// getVideoContentType returns the MIME type for video files
func getVideoContentType(ext string) string {
	contentTypes := map[string]string{
		".mp4":  "video/mp4",
		".webm": "video/webm",
		".mov":  "video/quicktime",
		".avi":  "video/x-msvideo",
		".mkv":  "video/x-matroska",
		".m4v":  "video/mp4",
		".flv":  "video/x-flv",
		".wmv":  "video/x-ms-wmv",
	}
	if ct, ok := contentTypes[strings.ToLower(ext)]; ok {
		return ct
	}
	return "video/mp4"
}

// getAudioContentType returns the MIME type for audio files
func getAudioContentType(ext string) string {
	contentTypes := map[string]string{
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".ogg":  "audio/ogg",
		".flac": "audio/flac",
		".m4a":  "audio/mp4",
		".aac":  "audio/aac",
		".wma":  "audio/x-ms-wma",
	}
	if ct, ok := contentTypes[strings.ToLower(ext)]; ok {
		return ct
	}
	return "audio/mpeg"
}

// getTextContentType returns the MIME type for text-based files
func getTextContentType(ext string) string {
	contentTypes := map[string]string{
		// Code files
		".js":  "application/javascript",
		".mjs": "application/javascript",
		".cjs": "application/javascript",
		".ts":  "application/typescript",
		".tsx": "application/typescript",
		".jsx": "application/javascript",
		".py":  "text/x-python",
		".go":  "text/x-go",
		".rs":  "text/x-rust",
		".java": "text/x-java",
		".kt":  "text/x-kotlin",
		".c":   "text/x-c",
		".cpp": "text/x-c++",
		".h":   "text/x-c",
		".hpp": "text/x-c++",
		".cs":  "text/x-csharp",
		".php": "text/x-php",
		".rb":  "text/x-ruby",
		".swift": "text/x-swift",
		".dart": "text/x-dart",
		".lua": "text/x-lua",
		".pl":  "text/x-perl",
		".r":   "text/x-r",
		".scala": "text/x-scala",
		".clj": "text/x-clojure",
		".sh":  "application/x-sh",
		".bat": "application/x-bat",
		".ps1": "text/x-powershell",
		".html": "text/html",
		".htm":  "text/html",
		".css":  "text/css",
		".scss": "text/x-scss",
		".sass": "text/x-sass",
		".json": "application/json",
		".xml":  "application/xml",
		".yaml": "text/yaml",
		".yml":  "text/yaml",
		".toml": "application/toml",
		".sql":  "application/sql",
		".graphql": "application/graphql",
		// Text files
		".txt":       "text/plain",
		".log":       "text/plain",
		".ini":       "text/plain",
		".cfg":       "text/plain",
		".conf":      "text/plain",
		".md":        "text/markdown",
		".markdown":  "text/markdown",
		".csv":       "text/csv",
		".tsv":       "text/tab-separated-values",
	}
	if ct, ok := contentTypes[strings.ToLower(ext)]; ok {
		return ct
	}
	return "text/plain"
}

func previewCacheKey(f *files.FileInfo, previewSize PreviewSize) string {
	return fmt.Sprintf("%x%x%x", f.RealPath(), f.ModTime.Unix(), previewSize)
}
