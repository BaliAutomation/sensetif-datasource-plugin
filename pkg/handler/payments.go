package handler

import (
	"github.com/BaliAutomation/sensetif-datasource/pkg/client"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/stripe/stripe-go/v72"
	"net/http"
)

type PayamentInfo struct {
	Customer stripe.Customer `json:"customer"`
	Price    stripe.Price    `json:"price"`
	Amount   int64           `json:"amount"`
	Currency stripe.Currency `json:"currency"`
}

//goland:noinspection GoUnusedParameter
func ListPayments(orgId int64, parameters []string, body []byte, kafka *client.KafkaClient, cassandra *client.CassandraClient) (*backend.CallResourceResponse, error) {
	log.DefaultLogger.Info("ListPayments()")

	return &backend.CallResourceResponse{
		Status:  http.StatusOK,
		Headers: make(map[string][]string),
		Body:    nil,
	}, nil
}
