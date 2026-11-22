package pagination

import "time"

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
