package httputil

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type BadRequest string

func (e BadRequest) Error() string { return string(e) }
func DecodeJSON[T any](r *http.Request) (T, error) {
	var input *T
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil || input == nil {
		var zero T
		return zero, BadRequest("invalid JSON body")
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		var zero T
		return zero, BadRequest("invalid JSON body")
	}
	return *input, nil
}

func PositiveID(raw, name string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, BadRequest("invalid " + name)
	}
	return id, nil
}

func PathID(r *http.Request, name string) (int64, error) {
	return PositiveID(r.PathValue(name), name)
}

// ScalarQuery rejects malformed URL encoding and repeated scalar parameters.
func ScalarQuery(r *http.Request, fields ...string) (url.Values, error) {
	q, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return nil, BadRequest("invalid query parameters")
	}
	for _, field := range fields {
		if len(q[field]) > 1 {
			return nil, BadRequest("repeated " + field)
		}
	}
	return q, nil
}
