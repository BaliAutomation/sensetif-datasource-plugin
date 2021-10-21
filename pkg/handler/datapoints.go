package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/BaliAutomation/sensetif-datasource/pkg/client"
	"github.com/BaliAutomation/sensetif-datasource/pkg/model"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

//goland:noinspection GoUnusedParameter
func ListDatapoints(orgId int64, params []string, body []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
	if len(params) < 3 {
		return nil, fmt.Errorf("%w: missing params: \"%v\"", model.ErrBadRequest, params)
	}

	datapoints := clients.Cassandra.FindAllDatapoints(orgId, params[1], params[2])
	if len(datapoints) == 0 {
		setupPlanIfNeeded(orgId)
	}
	rawJson, err := json.Marshal(datapoints)
	if err != nil {
		log.DefaultLogger.Error("Unable to marshal json")
		return nil, fmt.Errorf("%w: %s", model.ErrUnprocessableEntity, err.Error())
	}
	return &backend.CallResourceResponse{
		Status: http.StatusOK,
		Body:   rawJson,
	}, nil
}

func setupPlanIfNeeded(orgid int64) {

}

//goland:noinspection GoUnusedParameter
func GetDatapoint(orgId int64, params []string, body []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
	if len(params) < 4 {
		return nil, fmt.Errorf("%w: missing params: \"%v\"", model.ErrBadRequest, params)
	}

	datapoint := clients.Cassandra.GetDatapoint(orgId, params[1], params[2], params[3])
	bytes, err := json.Marshal(datapoint)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", model.ErrUnprocessableEntity, err.Error())
	}
	return &backend.CallResourceResponse{
		Status:  http.StatusOK,
		Headers: make(map[string][]string),
		Body:    bytes,
	}, nil
}

//goland:noinspection GoUnusedParameter
func UpdateDatapoint(orgId int64, params []string, body []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
	clients.Kafka.Send(model.ConfigurationTopic, "updateDatapoint:1:"+strconv.FormatInt(orgId, 10), body)
	return &backend.CallResourceResponse{
		Status: http.StatusAccepted,
	}, nil
}

//goland:noinspection GoUnusedParameter
func DeleteDatapoint(orgId int64, params []string, body []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
	if len(params) < 4 {
		return nil, fmt.Errorf("%w: missing params: \"%v\"", model.ErrBadRequest, params)
	}
	key := "deleteDatapoint:1:" + strconv.FormatInt(orgId, 10)
	data, err := json.Marshal(map[string]string{
		"project":   params[1],
		"subsystem": params[2],
		"datapoint": params[3],
	})
	if err == nil {
		clients.Kafka.Send(model.ConfigurationTopic, key, data)
		return &backend.CallResourceResponse{
			Status: http.StatusAccepted,
		}, nil
	}
	return &backend.CallResourceResponse{
		Status: http.StatusBadRequest,
	}, nil
}

//goland:noinspection GoUnusedParameter
func RenameDatapoint(orgId int64, params []string, body []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
	clients.Kafka.Send(model.ConfigurationTopic, "renameDatapoint:1:"+strconv.FormatInt(orgId, 10), body)
	return &backend.CallResourceResponse{
		Status: http.StatusAccepted,
	}, nil
}
