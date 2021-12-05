package handler

import (
	"encoding/json"
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	"net/http"
	"strconv"

	"github.com/BaliAutomation/sensetif-datasource/pkg/client"
	"github.com/BaliAutomation/sensetif-datasource/pkg/model"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

var updateDatapointSchema = pulsar.NewBytesSchema(nil)
var deleteDatapointSchema = pulsar.NewBytesSchema(nil)
var renameDatapointSchema = pulsar.NewBytesSchema(nil)

//goland:noinspection GoUnusedParameter
func ListDatapoints(orgId int64, params []string, body []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
	if len(params) < 3 {
		return nil, fmt.Errorf("%w: missing params: \"%v\"", model.ErrBadRequest, params)
	}

	datapoints := clients.Cassandra.FindAllDatapoints(orgId, params[1], params[2])
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

func GetDatapoint(orgId int64, params []string, _ []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
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

func UpdateDatapoint(orgId int64, _ []string, body []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
	clients.Pulsar.Send(model.ConfigurationTopic, updateDatapointSchema, "updateDatapoint:1:"+strconv.FormatInt(orgId, 10), body)
	return &backend.CallResourceResponse{
		Status: http.StatusAccepted,
	}, nil
}

func DeleteDatapoint(orgId int64, params []string, _ []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
	if len(params) < 4 {
		return nil, fmt.Errorf("%w: missing params: \"%v\"", model.ErrBadRequest, params)
	}
	key := "deleteDatapoint:1:" + strconv.FormatInt(orgId, 10)
	datapoint, err := GetDatapoint(orgId, params, nil, clients)
	if err != nil {
		return datapoint, err
	}
	body, err := json.Marshal(datapoint)
	if err == nil {
		clients.Pulsar.Send(model.ConfigurationTopic, deleteDatapointSchema, key, body)
		return &backend.CallResourceResponse{
			Status: http.StatusAccepted,
		}, nil
	}
	return nil, err
}

//goland:noinspection GoUnusedParameter
func RenameDatapoint(orgId int64, params []string, body []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
	clients.Pulsar.Send(model.ConfigurationTopic, renameDatapointSchema, "renameDatapoint:1:"+strconv.FormatInt(orgId, 10), body)
	return &backend.CallResourceResponse{
		Status: http.StatusAccepted,
	}, nil
}
