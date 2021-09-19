package model

type OrganizationSettings struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	StripeCustomer string `json:"stripecustomer"`
	CurrentPlan    string `json:"currentplan"`
}
