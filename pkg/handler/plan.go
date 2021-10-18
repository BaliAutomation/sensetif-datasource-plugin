package handler

import (
	"encoding/json"
	"fmt"
	"github.com/BaliAutomation/sensetif-datasource/pkg/client"
	"github.com/BaliAutomation/sensetif-datasource/pkg/model"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/checkout/session"
	"github.com/stripe/stripe-go/v72/price"
	"github.com/stripe/stripe-go/v72/product"
	"net/http"
	"os"
)

var (
	Products = LoadProductsFromStripe()
	Prices   = LoadPricesFromStripe()
)

type PlanPricing struct {
	Price string `json:"price"`
}

//goland:noinspection GoUnusedParameter
func ListPlans(orgId int64, parameters []string, body []byte, kafka *client.KafkaClient, cassandra *client.CassandraClient) (*backend.CallResourceResponse, error) {
	log.DefaultLogger.Info("ListPlans()")

	productPrices := map[string][]stripe.Price{}
	for _, price := range Prices {
		productPrices[price.Product.ID] = append(productPrices[price.Product.ID], price)
	}

	result := []*model.PlanSettings{}
	for _, product := range Products {
		result = append(result, &model.PlanSettings{
			Product: product,
			Prices:  productPrices[product.ID],
		})
	}

	resultJSON, _ := json.Marshal(result)

	return &backend.CallResourceResponse{
		Status:  http.StatusOK,
		Headers: make(map[string][]string),
		Body:    resultJSON,
	}, nil
}

func CheckOut(orgId int64, parameters []string, body []byte, kafka *client.KafkaClient, cassandra *client.CassandraClient) (*backend.CallResourceResponse, error) {
	log.DefaultLogger.Info("CheckOut()")
	log.DefaultLogger.Info(fmt.Sprintf("Parameters: %+v", parameters))
	log.DefaultLogger.Info(fmt.Sprintf("Body: %s", string(body)))
	var pricing PlanPricing
	err := json.Unmarshal(body, &pricing)
	if err != nil {
		return nil, err
	}
	stripe.Key = GetStripeKey()
	successUrl := "http://localhost:3000/api/plugins/sensetif-datasource/resources/_plans/success?session_id={CHECKOUT_SESSION_ID}"
	cancelUrl := "http://localhost:3000/api/plugins/sensetif-datasource/resources/_plans/canceled"
	params := &stripe.CheckoutSessionParams{
		SuccessURL: &successUrl,
		CancelURL:  &cancelUrl,
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			&stripe.CheckoutSessionLineItemParams{
				Price:    stripe.String(pricing.Price),
				Quantity: stripe.Int64(1),
			},
		},
	}
	log.DefaultLogger.Info("Calling Stripe")
	sess, err := session.New(params)
	if err != nil {
		log.DefaultLogger.Error(fmt.Sprintf("Strip error: %+v", err), err)
		return &backend.CallResourceResponse{
			Status:  http.StatusBadRequest,
			Headers: make(map[string][]string),
			Body:    []byte("{\"message\": \"Unable to establish Stripe session. Please try again later.\"}"),
		}, nil
	} else {
		log.DefaultLogger.Info(fmt.Sprintf("Session: %+v", sess))
		headers := make(map[string][]string)
		headers["Location"] = []string{sess.URL}
		log.DefaultLogger.Info("Redirect browser to: " + sess.URL)
		return &backend.CallResourceResponse{
			Status:  http.StatusSeeOther,
			Headers: headers,
			Body:    nil,
		}, nil
	}
}
func CheckOutSuccess(orgId int64, parameters []string, body []byte, kafka *client.KafkaClient, cassandra *client.CassandraClient) (*backend.CallResourceResponse, error) {
	log.DefaultLogger.Info("CheckOutSuccess()")
	// TODO: We have received Payment, redirect to THANK YOU page.
	headers := make(map[string][]string)
	headers["Location"] = []string{"/a/sensetif-app?tab=plans"}
	return &backend.CallResourceResponse{
		Status:  http.StatusSeeOther,
		Headers: headers,
		Body:    nil,
	}, nil
}

func CheckOutCancelled(orgId int64, parameters []string, body []byte, kafka *client.KafkaClient, cassandra *client.CassandraClient) (*backend.CallResourceResponse, error) {
	log.DefaultLogger.Info("CheckOutCancelled()")
	// Checkout was cancelled, just go back to Plans page.
	headers := make(map[string][]string)
	headers["Location"] = []string{"/a/sensetif-app?tab=plans"}
	return &backend.CallResourceResponse{
		Status:  http.StatusSeeOther,
		Headers: headers,
		Body:    nil,
	}, nil
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
