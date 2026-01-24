package domain

import "errors"

var (
	ErrNotFound         = errors.New("record not found")
	ErrInvalidParameter = errors.New("invalid parameter")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
)
