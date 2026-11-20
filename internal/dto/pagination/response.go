package pagination

type Metadata struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

type Response[T any] struct {
	Items      []T      `json:"items"`
	Pagination Metadata `json:"pagination"`
}
