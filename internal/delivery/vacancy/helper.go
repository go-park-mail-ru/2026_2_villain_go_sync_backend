package vacancy

import "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/sanitizer"

func ptrIfNotNil(s *sanitizer.SafeString) *string {
	if s == nil {
		return nil
	}
	return s.Ptr()
}
