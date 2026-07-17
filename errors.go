package web

import "errors"

var (
	// ErrResponseTooLarge is returned when the response body exceeds the maximum allowed size.
	ErrResponseTooLarge = errors.New("response body exceeds maximum allowed size")
)
