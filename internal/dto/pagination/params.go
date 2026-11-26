package pagination

import (
	"encoding/base64"
	"encoding/json"
	"errors"
)

type Params struct {
	Cursor *Cursor
	Limit  int
}

func (params Params) WithLookahead() Params {
	params.Limit++
	return params
}

func ValidLimit(limit int) bool {
	return limit >= 1 && limit <= 100
}

func NewParams(limit *int, token *string) (Params, error) {
	params := Params{Limit: 12}
	if limit != nil {
		if !ValidLimit(*limit) {
			return Params{}, errors.New("invalid limit")
		}
		params.Limit = *limit
	}
	if token != nil {
		if len(*token) > 256 {
			return Params{}, errors.New("invalid cursor")
		}
		data, err := base64.RawURLEncoding.DecodeString(*token)
		var cursor Cursor
		if err != nil || json.Unmarshal(data, &cursor) != nil || cursor.CreatedAt.IsZero() || cursor.ID <= 0 {
			return Params{}, errors.New("invalid cursor")
		}
		params.Cursor = &cursor
	}
	return params, nil
}
