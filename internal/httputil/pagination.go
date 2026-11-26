package httputil

import (
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/pagination"
)

func ReadPagination(w http.ResponseWriter, r *http.Request) (pagination.Params, bool) {
	var limit *int
	var cursor *string
	query := r.URL.Query()
	if query.Has("page") {
		WriteError(w, http.StatusBadRequest, "use cursor instead of page")
		return pagination.Params{}, false
	}
	if query.Has("limit") {
		value, err := strconv.Atoi(query.Get("limit"))
		if err != nil || !pagination.ValidLimit(value) {
			WriteError(w, http.StatusBadRequest, "invalid limit")
			return pagination.Params{}, false
		}
		limit = &value
	}
	if tokens, present := query["cursor"]; present {
		if len(tokens) != 1 {
			WriteError(w, http.StatusBadRequest, "invalid cursor")
			return pagination.Params{}, false
		}
		cursor = &tokens[0]
	}
	params, err := pagination.NewParams(limit, cursor)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return pagination.Params{}, false
	}
	return params, true
}

func WritePage[T any](w http.ResponseWriter, items []T, limit int, total int64, position func(T) pagination.Cursor) {
	page, err := pagination.NewPage(items, limit, total, position)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	WriteOK(w, http.StatusOK, page)
}
