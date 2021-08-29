package model

type AuthenticationType string

const (
	None        AuthenticationType = "none"
	Basic       AuthenticationType = "basic"
	BearerToken AuthenticationType = "bearerToken"
)
