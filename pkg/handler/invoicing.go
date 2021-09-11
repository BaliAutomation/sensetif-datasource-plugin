package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/BaliAutomation/sensetif-datasource/pkg/client"
	"github.com/BaliAutomation/sensetif-datasource/pkg/model"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

//goland:noinspection GoUnusedParameter
func ListInvoices(orgId int64, params []string, body []byte, kafka *client.KafkaClient, cassandra *client.CassandraClient) (*backend.CallResourceResponse, error) {
	log.DefaultLogger.Info("ListInvoices()")
	invoices := cassandra.FindAllInvoices(orgId)
	rawJson, err := json.Marshal(invoices)
	if err != nil {
		log.DefaultLogger.Error("Unable to marshal json")
		return nil, fmt.Errorf("%w: %s", model.ErrUnprocessableEntity, err.Error())
	}
	return &backend.CallResourceResponse{
		Status:  http.StatusOK,
		Headers: make(map[string][]string),
		Body:    rawJson,
	}, nil
}

//goland:noinspection GoUnusedParameter
func ListPayments(orgId int64, params []string, body []byte, kafka *client.KafkaClient, cassandra *client.CassandraClient) (*backend.CallResourceResponse, error) {
	log.DefaultLogger.Info("ListPayments()")
	payments := cassandra.FindAllPayments(orgId)
	rawJson, err := json.Marshal(payments)
	if err != nil {
		log.DefaultLogger.Error("Unable to marshal json")
		return nil, fmt.Errorf("%w: %s", model.ErrUnprocessableEntity, err.Error())
	}
	return &backend.CallResourceResponse{
		Status:  http.StatusOK,
		Headers: make(map[string][]string),
		Body:    rawJson,
	}, nil
}

//goland:noinspection GoUnusedParameter
// TODO: No idea how this looks like yet, but this is to be called AFTER confirmed payment through gateway.
func MakePayment(orgId int64, params []string, body []byte, kafka *client.KafkaClient, cassandra *client.CassandraClient) (*backend.CallResourceResponse, error) {
	log.DefaultLogger.Info("MakePayment()")
	empty := make([]byte, 0)
	return &backend.CallResourceResponse{
		Status:  http.StatusOK,
		Headers: make(map[string][]string),
		Body:    empty,
	}, nil
}
