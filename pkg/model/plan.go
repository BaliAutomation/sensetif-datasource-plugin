package model

import "github.com/stripe/stripe-go/v72"

//
//import "time"
//
//
//type PlanLimits struct {
//	MaxDatapoints    int32        `json:"maxdatapoints"`
//	MaxStorage       TimeToLive   `json:"maxstorage"`
//	MinPollInterval  PollInterval `json:"minpollinterval"`
//}
//
//type PlanSettings struct {
//	Name        string     `json:"name"`
//	Title       string     `json:"title"`
//	SubTitle    string     `json:"subtitle"`
//	Description string     `json:"description"`
//	Private     bool       `json:"private"`
//	Active      bool       `json:"active"`
//	Limits      PlanLimits `json:"limits"`
//	Start       time.Time  `json:"start"`
//	End         time.Time  `json:"end"`
//	Price       int64      `json:"price"`
//	Currency    string     `json:"currency"`
//}

type PlanSettings struct {
	Product stripe.Product `json:"product"`
	Prices  []stripe.Price `json:"prices"`
}
