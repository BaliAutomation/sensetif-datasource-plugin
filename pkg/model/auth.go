package model

type AuthenticationType int

const (
	None AuthenticationType = iota
	Basic
	QueryParam
)
