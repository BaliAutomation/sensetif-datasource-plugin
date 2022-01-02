package client

import (
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/price"
	"github.com/stripe/stripe-go/v72/product"
	"strconv"
)

type Stripe interface {
	IsCurrentPlan(orgId int64, planId string)
}

type StripeClient struct {
	PlansPerOrg      map[int64]string
	Products         []stripe.Product
	Prices           []stripe.Price
	authorizationKey string
}

func (s *StripeClient) InitializeStripe(authorization string) {
	log.DefaultLogger.Info("Initializing Stripe products.")
	s.authorizationKey = authorization
	s.PlansPerOrg = make(map[int64]string)
	s.Products = s.LoadProductsFromStripe()
	s.Prices = s.LoadPricesFromStripe()
}

func (s *StripeClient) GetStripeKey() string {
	return s.authorizationKey
}

func (s *StripeClient) LoadPricesFromStripe() []stripe.Price {
	stripe.Key = s.GetStripeKey()
	include := true
	recurring := "recurring"
	params := &stripe.PriceListParams{
		Active: &include,
		Type:   &recurring,
	}
	i := price.List(params)
	result := []stripe.Price{}
	for i.Next() {
		p := *i.Price()
		log.DefaultLogger.Info("Found price: " + p.ID)
		if p.Active {
			result = append(result, p)
		}
	}
	return result
}

func (s *StripeClient) LoadProductsFromStripe() []stripe.Product {
	stripe.Key = s.GetStripeKey()
	params := &stripe.ProductListParams{}
	i := product.List(params)
	result := []stripe.Product{}
	for i.Next() {
		p := *i.Product()
		log.DefaultLogger.Info("Found product: " + p.ID)
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
		stripe.Key = s.GetStripeKey()
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
