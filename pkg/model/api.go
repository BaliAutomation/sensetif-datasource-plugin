package model

import (
	"encoding/json"
	"errors"
)

type Command struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`

	Params  map[string]string `json:"params"`
	Payload json.RawMessage   `json:"payload"`
	OrgID   int64             `json:"-"`
}

var (
	ErrServerError         = errors.New("internal error")
	ErrUnprocessableEntity = errors.New("wrong payload format")
	ErrBadRequest          = errors.New("bad request")
	ErrNotFound            = errors.New("not found")
)
