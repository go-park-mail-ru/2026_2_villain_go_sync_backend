package httputil

import (
	"math"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/pagination"
)

func ReadPagination(w http.ResponseWriter, r *http.Request) (page, limit int, ok bool) {
	page, limit = 1, 12
	query := r.URL.Query()
	if query.Has("page") {
		value, err := strconv.Atoi(query.Get("page"))
		if err != nil || value < 1 {
			WriteError(w, http.StatusBadRequest, "invalid page")
			return 0, 0, false
		}
		page = value
	}
	if query.Has("limit") {
		value, err := strconv.Atoi(query.Get("limit"))
		if err != nil || value < 1 || value > 100 {
			WriteError(w, http.StatusBadRequest, "invalid limit")
			return 0, 0, false
		}
		limit = value
	}
	if int64(page-1) > math.MaxInt64/int64(limit) {
		WriteError(w, http.StatusBadRequest, "invalid page")
		return 0, 0, false
	}
	return page, limit, true
}

func WritePage[T any](w http.ResponseWriter, items []T, page, limit int, total int64) {
	if items == nil {
		items = []T{}
	}
	totalPages := total / int64(limit)
	if total%int64(limit) != 0 {
		totalPages++
	}
	WriteOK(w, http.StatusOK, pagination.Response[T]{
		Items: items,
		Pagination: pagination.Metadata{
			Page: page, Limit: limit, Total: total, TotalPages: totalPages,
		},
	})
}
