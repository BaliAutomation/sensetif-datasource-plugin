package model

type AuthenticationType int

const (
	None AuthenticationType = iota
	Basic
	QueryParam
)

type Auth struct{}

type BasicAuth struct {
	Auth
	User     string `json:"user"`
	Password string `json:"password"`
}

type QueryParamAuth struct {
	Auth
	Key   string `json:"key"`
	Value string `json:"value"`
}
