package httputil

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/pagination"
)

func ReadPagination(w http.ResponseWriter, r *http.Request) (pagination.Params, bool) {
	params := pagination.Params{Limit: 12}
	query := r.URL.Query()
	if query.Has("page") {
		WriteError(w, http.StatusBadRequest, "use cursor instead of page")
		return pagination.Params{}, false
	}
	if query.Has("limit") {
		value, err := strconv.Atoi(query.Get("limit"))
		if err != nil || value < 1 || value > 100 {
			WriteError(w, http.StatusBadRequest, "invalid limit")
			return pagination.Params{}, false
		}
		params.Limit = value
	}
	if tokens, present := query["cursor"]; present {
		if len(tokens) != 1 || len(tokens[0]) > 256 {
			WriteError(w, http.StatusBadRequest, "invalid cursor")
			return pagination.Params{}, false
		}
		data, err := base64.RawURLEncoding.DecodeString(tokens[0])
		cursor := new(pagination.Cursor)
		if err != nil || json.Unmarshal(data, cursor) != nil || cursor.CreatedAt.IsZero() || cursor.ID <= 0 {
			WriteError(w, http.StatusBadRequest, "invalid cursor")
			return pagination.Params{}, false
		}
		params.Cursor = cursor
	}
	return params, true
}

func WritePage[T any](w http.ResponseWriter, items []T, limit int, total int64, position func(T) pagination.Cursor) {
	var nextCursor *string
	if len(items) > limit {
		items = items[:limit]
		data, err := json.Marshal(position(items[limit-1]))
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "internal error")
			return
		}
		token := base64.RawURLEncoding.EncodeToString(data)
		nextCursor = &token
	}
	if items == nil {
		items = []T{}
	}
	WriteOK(w, http.StatusOK, pagination.Response[T]{
		Items: items,
		Pagination: pagination.Metadata{
			Limit: limit, Total: total, NextCursor: nextCursor,
		},
	})
}
