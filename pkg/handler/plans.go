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
	"net/http"
	"os"
	"strconv"
)

type PlanPricing struct {
	Price string `json:"price"`
}

type SubscriptionInfo struct {
	Customer int64           `json:"customer"`
	Price    stripe.Price    `json:"price"`
	Amount   int64           `json:"amount"`
	Currency stripe.Currency `json:"currency"`
}

type SessionProxy struct {
	Id string `json:"id"`
}

//goland:noinspection GoUnusedParameter
func ListPlans(orgId int64, parameters []string, body []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
	log.DefaultLogger.Info("ListPlans()")

	productPrices := map[string][]stripe.Price{}
	for _, prize := range clients.Stripe.Prices {
		productPrices[prize.Product.ID] = append(productPrices[prize.Product.ID], prize)
	}
	p, _ := json.Marshal(productPrices)
	log.DefaultLogger.Info(fmt.Sprintf("Plans: %s", p))

	organization := clients.Cassandra.GetOrganization(orgId)

	result := []*model.PlanSettings{}
	for n, prod := range clients.Stripe.Products {
		p, _ := json.Marshal(prod)
		log.DefaultLogger.Info(fmt.Sprintf("Product: %s - %s", n, p))
		result = append(result, &model.PlanSettings{
			Product:     prod,
			Prices:      productPrices[prod.ID],
			Selected:    clients.Stripe.IsSelected(orgId, prod.ID, organization.StripeCustomer),
			Expired:     false,
			GracePeriod: true,
		})
	}

	plansInJson, _ := json.Marshal(result)

	return &backend.CallResourceResponse{
		Status:  http.StatusOK,
		Headers: make(map[string][]string),
		Body:    plansInJson,
	}, nil
}

func CheckOut(orgId int64, parameters []string, body []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
	log.DefaultLogger.Info("CheckOut()")
	log.DefaultLogger.Info(fmt.Sprintf("Parameters: %+v", parameters))
	log.DefaultLogger.Info(fmt.Sprintf("Body: %s", string(body)))

	var pricing PlanPricing
	err := json.Unmarshal(body, &pricing)
	if err != nil {
		return nil, err
	}
	stripe.Key = GetStripeKey()
	successUrl := "https://sensetif.net/a/sensetif-app?tab=succeeded&session_id={CHECKOUT_SESSION_ID}"
	cancelUrl := "https://sensetif.net/a/sensetif-app?tab=cancelled&session_id={CHECKOUT_SESSION_ID}"
	params := &stripe.CheckoutSessionParams{
		SuccessURL: &successUrl,
		CancelURL:  &cancelUrl,

		PaymentMethodTypes: stripe.StringSlice([]string{
			//"alipay",
			"card",
			//"ideal",
			//"fpx",
			//"bacs_debit",
			//"bancontact",
			//"giropay",
			//"p24",
			//"eps",
			//"sofort",
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

func CheckOutSuccess(orgId int64, parameters []string, body []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
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
		var paymentInfo = SubscriptionInfo{
			Customer: orgId,
			Amount:   stripeSession.AmountTotal,
			Currency: stripeSession.Currency,
		}
		bytes, err := json.Marshal(paymentInfo)
		if err == nil {
			clients.Kafka.Send(model.PaymentsTopic, strconv.FormatInt(orgId, 10), bytes)
		} else {
			clients.Kafka.Send(model.ErrorsTopic, model.GlobalKey, []byte(fmt.Sprintf("%+v", err)))
		}
	}
	return &backend.CallResourceResponse{
		Status: http.StatusOK,
		Body:   nil,
	}, nil
}

func CheckOutCancelled(orgId int64, parameters []string, body []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
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
		var paymentInfo = SubscriptionInfo{
			Customer: orgId,
			Amount:   stripeSession.AmountTotal,
			Currency: stripeSession.Currency,
		}
		bytes, err := json.Marshal(paymentInfo)
		if err == nil {
			clients.Kafka.Send(model.PaymentsTopic, strconv.FormatInt(orgId, 10), bytes)
		} else {
			clients.Kafka.Send(model.ErrorsTopic, model.GlobalKey, []byte(fmt.Sprintf("%+v", err)))
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
