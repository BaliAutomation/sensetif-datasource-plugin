package model

type AuthenticationType int

const (
	None        AuthenticationType = 0
	Basic       AuthenticationType = 1
	BearerToken AuthenticationType = 2
)
