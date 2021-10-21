package client

import (
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/price"
	"github.com/stripe/stripe-go/v72/product"
	"os"
	"strconv"
)

type Stripe interface {
	IsCurrentPlan(orgId int64, planId string)
}

type StripeClient struct {
	PlansPerOrg map[int64]string
	Products    []stripe.Product
	Prices      []stripe.Price
}

func (s *StripeClient) InitializeStripe(authorization string) {
	s.PlansPerOrg = make(map[int64]string)
	s.Products = LoadProductsFromStripe()
	s.Prices = LoadPricesFromStripe()

}

func GetStripeKey() string {
	if key, ok := os.LookupEnv("STRIPE_KEY"); ok {
		return key
	}
	// If not set in environment, return the key for the Strip Test Mode.
	return "sk_test_51JZvsFBil9jp3I2LySc7piIiEpXUlDdcxpXdVERSLL10nv2AUM1dfoCjSAZIMJ2XlC8zK1tkxJw85F2KlkBh9mxE00Vne8Kp5Z"
}

func LoadPricesFromStripe() []stripe.Price {
	stripe.Key = GetStripeKey()
	params := &stripe.PriceListParams{}
	i := price.List(params)
	result := []stripe.Price{}
	for i.Next() {
		p := *i.Price()
		if p.Active {
			result = append(result, p)
		}
	}
	return result
}

func LoadProductsFromStripe() []stripe.Product {
	stripe.Key = GetStripeKey()
	params := &stripe.ProductListParams{}
	i := product.List(params)
	result := []stripe.Product{}
	for i.Next() {
		p := *i.Product()
		if p.Active {
			result = append(result, p)
		}
	}
	return result
}

func (s *StripeClient) IsSelected(orgId int64, id string, stripeCustomer string) bool {
	if stripeCustomer == "" {
		return id == "prod_KFtaaxi4gvLTTL"
	}
	planId, exists := s.PlansPerOrg[orgId]
	if !exists {
		stripe.Key = GetStripeKey()
		params := &stripe.CustomerParams{}
		cust, err := customer.Get(stripeCustomer, params)
		if err != nil {
			log.DefaultLogger.Error("Unable to GET the customer from Stripe: " + err.Error())
			return id == "prod_KFtaaxi4gvLTTL"
		}
		subscriptions := cust.Subscriptions.Data
		i := 0
		for _, sub := range subscriptions {
			s.PlansPerOrg[orgId] = sub.Plan.ID
			i++
		}
		if i != 1 {
			log.DefaultLogger.Warn("Organization had incorrect number of Subscriptions: " + strconv.Itoa(i))
		}
	}
	return planId == id
}
