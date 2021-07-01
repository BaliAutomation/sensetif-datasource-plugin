package model

import (
	"errors"
)

var (
	ErrServerError         = errors.New("internal error")
	ErrUnprocessableEntity = errors.New("wrong payload format")
	ErrBadRequest          = errors.New("bad request")
	ErrNotFound            = errors.New("not found")
)
