package handler

import (
	"encoding/json"
	"fmt"
	"github.com/BaliAutomation/sensetif-datasource/pkg/client"
	"github.com/BaliAutomation/sensetif-datasource/pkg/model"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/price"
	"github.com/stripe/stripe-go/v72/product"
	"net/http"
)

var (
	Products = LoadProductsFromStripe()
	Prices   = LoadPricesFromStripe()
)

//goland:noinspection GoUnusedParameter
func ListPlans(orgId int64, parameters []string, body []byte, kafka *client.KafkaClient, cassandra *client.CassandraClient) (*backend.CallResourceResponse, error) {
	log.DefaultLogger.Info("ListPlans()")
	pricesJson, err := json.Marshal(Prices)
	productsJson, err := json.Marshal(Products)
	rawJson := "{\"products\":" + string(productsJson) + ",\"prices\":" + string(pricesJson) + "}"
	if err != nil {
		log.DefaultLogger.Error("Unable to marshal json")
		return nil, fmt.Errorf("%w: %s", model.ErrUnprocessableEntity, err.Error())
	}
	return &backend.CallResourceResponse{
		Status:  http.StatusOK,
		Headers: make(map[string][]string),
		Body:    []byte(rawJson),
	}, nil
}

func GetStripeKey() string {
	// TODO: fetch from /etc/ something
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
