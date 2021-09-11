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
func ListPlans(orgId int64, params []string, body []byte, kafka *client.KafkaClient, cassandra *client.CassandraClient) (*backend.CallResourceResponse, error) {
	log.DefaultLogger.Info("ListPlans()")
	plans := cassandra.FindAllPlans(orgId)
	rawJson, err := json.Marshal(plans)
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
