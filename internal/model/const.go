package model

import "errors"

var (
	ErrNotFound = errors.New("url not found")
	ErrDeleted  = errors.New("url deleted")
)
