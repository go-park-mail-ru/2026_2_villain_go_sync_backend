package profile

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/httputil"
)

const maxProfileImageSize int64 = 10 << 20

func openProfileImage(w http.ResponseWriter, r *http.Request, field string) (multipart.File, string, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxProfileImageSize+(1<<20))
	file, header, err := r.FormFile(field)
	if err != nil {
		httputil.WriteAppError(
			w,
			fmt.Errorf("%w: %w", apperrors.ErrInvalidInput, err),
			"invalid image upload",
		)
		return nil, "", false
	}

	if header.Size > maxProfileImageSize {
		file.Close()
		httputil.WriteError(w, http.StatusRequestEntityTooLarge, "image is too large")
		return nil, "", false
	}
	if header.Size == 0 {
		file.Close()
		httputil.WriteError(w, http.StatusBadRequest, "image is empty")
		return nil, "", false
	}

	buf := make([]byte, 512)
	n, err := io.ReadFull(file, buf)
	if err != nil &&
		!errors.Is(err, io.EOF) &&
		!errors.Is(err, io.ErrUnexpectedEOF) {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to read file")
		file.Close()
		return nil, "", false
	}
	if n == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "image is empty")
		file.Close()
		return nil, "", false
	}

	contentType := http.DetectContentType(buf[:n])
	switch contentType {
	case "image/png", "image/jpeg", "image/webp":
	default:
		httputil.WriteError(w, http.StatusUnsupportedMediaType, "unsupported image format")
		file.Close()
		return nil, "", false
	}
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to reset image file")
		file.Close()
		return nil, "", false
	}
	return file, contentType, true
}
