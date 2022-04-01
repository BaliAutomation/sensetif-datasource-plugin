package model

import "github.com/stripe/stripe-go/v72"

type PlanSettings struct {
	Product     stripe.Product `json:"product"`
	Prices      []stripe.Price `json:"prices"`
	Selected    bool           `json:"selected"`
	Expired     bool           `json:"expiring"`
	GracePeriod bool           `json:"gracePeriod"`
}

type PlanLimits struct {
	MaxStorage      string   `json:"maxStorage"`
	MaxDatapoints   uint64   `json:"maxDatapoints"`
	MinPollInterval string   `json:"minPollInterval"`
	Permissions     []string `json:"permissions"`
}
