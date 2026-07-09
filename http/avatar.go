package fbhttp

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

	"github.com/B3llo/the-filebrowser/img"
)

// maxAvatarUploadSize caps the raw image body accepted for an avatar upload,
// well above what a 256x256 output needs but small enough to avoid abuse.
const maxAvatarUploadSize = 8 << 20 // 8 MiB

const avatarSize = 256

func avatarCacheKey(id uint) string {
	return fmt.Sprintf("user_%d", id)
}

func avatarPostHandler(imgSvc ImgService, avatarStore FileCache) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if r.Body == nil {
			return http.StatusBadRequest, nil
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxAvatarUploadSize)

		buf := &bytes.Buffer{}
		err := imgSvc.Resize(r.Context(), r.Body, avatarSize, avatarSize, buf,
			img.WithMode(img.ResizeModeFill), img.WithFormat(img.FormatPng), img.WithQuality(img.QualityHigh))
		if errors.Is(err, img.ErrUnsupportedFormat) {
			return http.StatusBadRequest, nil
		}
		if err != nil {
			return errToStatus(err), err
		}

		if err := avatarStore.Store(r.Context(), avatarCacheKey(d.user.ID), buf.Bytes()); err != nil {
			return http.StatusInternalServerError, err
		}

		d.user.Avatar = "custom"
		if err := d.store.Users.Update(d.user, "Avatar"); err != nil {
			return http.StatusInternalServerError, err
		}

		return http.StatusOK, nil
	})
}

func avatarDeleteHandler(avatarStore FileCache) handleFunc {
	return withUser(func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if err := avatarStore.Delete(r.Context(), avatarCacheKey(d.user.ID)); err != nil {
			return http.StatusInternalServerError, err
		}

		d.user.Avatar = ""
		if err := d.store.Users.Update(d.user, "Avatar"); err != nil {
			return http.StatusInternalServerError, err
		}

		return http.StatusOK, nil
	})
}

func avatarGetHandler(avatarStore FileCache) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
		id, err := getUserID(r)
		if err != nil {
			return http.StatusBadRequest, err
		}

		value, exist, err := avatarStore.Load(r.Context(), avatarCacheKey(id))
		if err != nil {
			return http.StatusInternalServerError, err
		}
		if !exist {
			return http.StatusNotFound, nil
		}

		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "private, max-age=60")
		if _, err := w.Write(value); err != nil {
			return http.StatusInternalServerError, err
		}
		return 0, nil
	})
}
