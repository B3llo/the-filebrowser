// Package video provides server-side video thumbnail extraction via ffmpeg.
package video

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/marusama/semaphore/v2"
)

// ErrNotAvailable is returned when ffmpeg is not configured or was not found on the system.
var ErrNotAvailable = errors.New("ffmpeg not available")

// Service extracts thumbnail frames from video files using ffmpeg.
type Service struct {
	ffmpegPath  string
	sem         semaphore.Semaphore
	execTimeout time.Duration
}

// New builds a video thumbnail service. Pass an empty ffmpegPath to disable it.
func New(ffmpegPath string, workers int, execTimeout time.Duration) *Service {
	return &Service{
		ffmpegPath:  ffmpegPath,
		sem:         semaphore.New(workers),
		execTimeout: execTimeout,
	}
}

// Available reports whether the service has a usable ffmpeg binary.
func (s *Service) Available() bool {
	return s != nil && s.ffmpegPath != ""
}

// Detect resolves the ffmpeg binary to use: the explicit configured path if
// it is executable, otherwise a PATH lookup for "ffmpeg". Returns "" if
// neither is usable, in which case the caller should disable the feature.
func Detect(configuredPath string) string {
	if configuredPath != "" {
		if p, err := exec.LookPath(configuredPath); err == nil {
			return p
		}
		log.Printf("[WARN] configured ffmpegPath %q is not executable, falling back to PATH lookup", configuredPath)
	}
	if p, err := exec.LookPath("ffmpeg"); err == nil {
		return p
	}
	return ""
}

// ExtractFrame extracts a single frame from the video at realPath and
// returns it as JPEG-encoded bytes, downscaled to a moderate size while
// preserving aspect ratio. Callers that need an exact thumbnail size should
// resize the returned bytes further (e.g. via img.Service.Resize).
func (s *Service) ExtractFrame(ctx context.Context, realPath string) ([]byte, error) {
	if !s.Available() {
		return nil, ErrNotAvailable
	}

	if err := s.sem.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	defer s.sem.Release(1)

	cctx, cancel := context.WithTimeout(ctx, s.execTimeout)
	defer cancel()

	// Try seeking a bit into the video first (avoids all-black opening
	// frames); fall back to the very first frame for very short clips.
	frame, err := s.run(cctx, realPath, time.Second)
	if err != nil || len(frame) == 0 {
		frame, err = s.run(cctx, realPath, 0)
	}
	if err != nil {
		return nil, err
	}
	if len(frame) == 0 {
		return nil, errors.New("ffmpeg produced no frame data")
	}
	return frame, nil
}

func (s *Service) run(ctx context.Context, realPath string, seek time.Duration) ([]byte, error) {
	args := []string{
		"-y",
		"-ss", fmt.Sprintf("%.2f", seek.Seconds()), // seek BEFORE -i: fast keyframe seek, avoids decoding the whole file
		"-i", realPath,
		"-vframes", "1",
		"-an",
		"-vf", "scale='min(512,iw)':'min(512,ih)':force_original_aspect_ratio=decrease",
		"-f", "image2pipe",
		"-vcodec", "mjpeg",
		"pipe:1",
	}

	cmd := exec.CommandContext(ctx, s.ffmpegPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg: %w (stderr: %s)", err, stderr.String())
	}

	return stdout.Bytes(), nil
}
