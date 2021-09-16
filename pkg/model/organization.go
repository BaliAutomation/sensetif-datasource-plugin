package model

type Address struct {
	Address1 string `json:"address1"`
	Address2 string `json:"address2"`
	City     string `json:"city"`
	ZipCode  string `json:"zipcode"`
	State    string `json:"state"`
	Country  string `json:"country"`
}

type OrganizationSettings struct {
	Name           string  `json:"name"`
	Address        Address `json:"address"`
	Email          string  `json:"email"`
	StripeCustomer string  `json:"stripecustomer"`
	CurrentPlan    string  `json:"currentplan"`
}
