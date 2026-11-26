package delivery

import "errors"

var (
	ErrInvalidRole    = errors.New("invalid role")
	ErrInvalidRequest = errors.New("invalid request")
)
