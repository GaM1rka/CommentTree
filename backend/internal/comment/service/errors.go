package service

import "errors"

var (
	ErrNotFound   = errors.New("comment not found")
	ErrValidation = errors.New("validation error")
)
