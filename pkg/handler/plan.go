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
	"strconv"
)

var (
	Products = LoadProductsFromStripe()
	Prices   = LoadPricesFromStripe()
)

type PlanPricing struct {
	Price string `json:"price"`
}

type PaymentInfo struct {
	Customer stripe.Customer `json:"customer"`
	Price    stripe.Price    `json:"price"`
	Amount   int64           `json:"amount"`
	Currency stripe.Currency `json:"currency"`
}

type SessionProxy struct {
	Id string `json:"id"`
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
	successUrl := "https://sensetif.net/a/sensetif-app?tab=succeeded?session_id={CHECKOUT_SESSION_ID}"
	cancelUrl := "https://sensetif.net/a/sensetif-app?tab=cancelled?session_id={CHECKOUT_SESSION_ID}"
	params := &stripe.CheckoutSessionParams{
		SuccessURL: &successUrl,
		CancelURL:  &cancelUrl,
		PaymentMethodTypes: stripe.StringSlice([]string{
			//"alipay",
			"card",
			//"ideal",
			//"fpx",
			"bacs_debit",
			//"bancontact",
			"giropay",
			"p24",
			"eps",
			"sofort",
			"sepa_debit",
			//"grabpay",
			//"afterpay_clearpay",
			//"acss_debit",
			//"wechat_pay",
			//"boleto",
			//"oxxo",
		}),
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
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
		log.DefaultLogger.Info("Redirect browser to: " + sess.URL)
		return &backend.CallResourceResponse{
			Status: http.StatusOK,
			Body:   []byte(sess.URL),
		}, nil
	}
}

func CheckOutSuccess(orgId int64, parameters []string, body []byte, kafka *client.KafkaClient, cassandra *client.CassandraClient) (*backend.CallResourceResponse, error) {
	log.DefaultLogger.Info("CheckOutSuccess()")

	var sessionProxy SessionProxy
	err := json.Unmarshal(body, &sessionProxy)
	if err != nil {
		return nil, err
	}

	params := &stripe.CheckoutSessionParams{}
	stripeSession, err := session.Get(sessionProxy.Id, params)
	if err != nil {
		log.DefaultLogger.Error("Unable to GET checkout session after Success")
		return &backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte(fmt.Sprintf("%+v", err)),
		}, nil
	}
	log.DefaultLogger.Info(fmt.Sprintf("Payment Success: %+v", stripeSession))
	if stripeSession.PaymentStatus == stripe.CheckoutSessionPaymentStatusPaid {
		var paymentInfo = PaymentInfo{
			Customer: stripe.Customer{},
			Amount:   stripeSession.AmountTotal,
			Currency: stripeSession.Currency,
		}
		bytes, err := json.Marshal(paymentInfo)
		if err == nil {
			kafka.Send(model.PaymentsTopic, strconv.FormatInt(orgId, 10), bytes)
		} else {
			kafka.Send(model.ErrorsTopic, model.GlobalKey, []byte(fmt.Sprintf("%+v", err)))
		}
	}
	return &backend.CallResourceResponse{
		Status: http.StatusOK,
		Body:   nil,
	}, nil
}

func CheckOutCancelled(orgId int64, parameters []string, body []byte, kafka *client.KafkaClient, cassandra *client.CassandraClient) (*backend.CallResourceResponse, error) {
	log.DefaultLogger.Info("CheckOutCancelled()")

	var sessionProxy SessionProxy
	err := json.Unmarshal(body, &sessionProxy)
	if err != nil {
		return nil, err
	}

	params := &stripe.CheckoutSessionParams{}
	stripeSession, err := session.Get(sessionProxy.Id, params)
	if err != nil {
		log.DefaultLogger.Error("Unable to GET checkout session after CANCEL")
		return &backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte(fmt.Sprintf("%+v", err)),
		}, nil
	}
	log.DefaultLogger.Info(fmt.Sprintf("Payment Success: %+v", stripeSession))
	if stripeSession.PaymentStatus == stripe.CheckoutSessionPaymentStatusPaid {
		var paymentInfo = PaymentInfo{
			Customer: stripe.Customer{},
			Amount:   stripeSession.AmountTotal,
			Currency: stripeSession.Currency,
		}
		bytes, err := json.Marshal(paymentInfo)
		if err == nil {
			kafka.Send(model.PaymentsTopic, strconv.FormatInt(orgId, 10), bytes)
		} else {
			kafka.Send(model.ErrorsTopic, model.GlobalKey, []byte(fmt.Sprintf("%+v", err)))
		}
	}
	return &backend.CallResourceResponse{
		Status: http.StatusOK,
		Body:   nil,
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
