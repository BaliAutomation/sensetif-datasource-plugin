package model

type SubsystemSettings struct {
	Project       string `json:"project"`
	Name          string `json:"name"`          // validate regexp [a-z][A-Za-z0-9_]*
	Title         string `json:"title"`         // allow all characters
	Locallocation string `json:"locallocation"` // allow all characters
}
