package pagination

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

type Metadata struct {
	Limit      int     `json:"limit"`
	Total      int64   `json:"total"`
	NextCursor *string `json:"next_cursor"`
}

type Response[T any] struct {
	Items      []T      `json:"items"`
	Pagination Metadata `json:"pagination"`
}

type Cursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        int64     `json:"id"`
}

func NewPage[T any](items []T, limit int, total int64, position func(T) Cursor) (Response[T], error) {
	if !ValidLimit(limit) {
		return Response[T]{}, errors.New("invalid limit")
	}
	var nextCursor *string
	if len(items) > limit {
		items = items[:limit]
		data, err := json.Marshal(position(items[limit-1]))
		if err != nil {
			return Response[T]{}, err
		}
		token := base64.RawURLEncoding.EncodeToString(data)
		nextCursor = &token
	}
	if items == nil {
		items = []T{}
	}
	return Response[T]{
		Items: items,
		Pagination: Metadata{
			Limit: limit, Total: total, NextCursor: nextCursor,
		},
	}, nil
}
