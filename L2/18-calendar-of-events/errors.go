package main

import "errors"

var (
	ErrInvalidID          = errors.New("invalid id")
	ErrInvalidUserID      = errors.New("invalid user_id")
	ErrInvalidDate        = errors.New("invalid date")
	ErrInvalidDescription = errors.New("invalid description")
	ErrInvalidHTTPMethod  = errors.New("invalid http method")
	ErrInvalidContentType = errors.New("invalid content type")
)
