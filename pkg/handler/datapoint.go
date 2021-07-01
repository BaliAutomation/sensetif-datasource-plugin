package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/BaliAutomation/sensetif-datasource/pkg/util"

	"github.com/BaliAutomation/sensetif-datasource/pkg/client"
	"github.com/BaliAutomation/sensetif-datasource/pkg/model"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

//goland:noinspection GoUnusedParameter
func ListDatapoints(orgId int64, params []string, body []byte, kafka client.Kafka, cassandra client.Cassandra) (*backend.CallResourceResponse, error) {
	if len(params) < 3 {
		return nil, fmt.Errorf("%w: missing params: \"%v\"", model.ErrBadRequest, params)
	}

	datapoints := cassandra.FindAllDatapoints(orgId, params[1], params[2])
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

//goland:noinspection GoUnusedParameter
func GetDatapoint(orgId int64, params []string, body []byte, kafka client.Kafka, cassandra client.Cassandra) (*backend.CallResourceResponse, error) {
	if len(params) < 4 {
		return nil, fmt.Errorf("%w: missing params: \"%v\"", model.ErrBadRequest, params)
	}

	datapoint := cassandra.GetDatapoint(orgId, params[1], params[2], params[3])
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
func UpdateDatapoint(orgId int64, params []string, body []byte, kafka client.Kafka, cassandra client.Cassandra) (*backend.CallResourceResponse, error) {
	kafka.Send(model.ConfigurationTopic, "updateDatapoint:1:"+strconv.FormatInt(orgId, 10), body)
	if util.IsDevelopmentMode() {
		var datapoint model.DatapointSettings
		if err := json.Unmarshal(body, &datapoint); err != nil {
			log.DefaultLogger.Error(fmt.Sprintf("Could not unmarshal datapoint. error: %v", err))
			return nil, fmt.Errorf("%w: invalid subsystem json", model.ErrBadRequest)
		}
		if err := cassandra.UpsertDatapoint(orgId, &datapoint); err != nil {
			return nil, err
		}
	}
	return &backend.CallResourceResponse{
		Status: http.StatusAccepted,
	}, nil
}

//goland:noinspection GoUnusedParameter
func DeleteDatapoint(orgId int64, params []string, body []byte, kafka client.Kafka, cassandra client.Cassandra) (*backend.CallResourceResponse, error) {
	kafka.Send(model.ConfigurationTopic, "deleteDatapoint:1:"+strconv.FormatInt(orgId, 10), body)
	return &backend.CallResourceResponse{
		Status: http.StatusAccepted,
	}, nil
}

//goland:noinspection GoUnusedParameter
func RenameDatapoint(orgId int64, params []string, body []byte, kafka client.Kafka, cassandra client.Cassandra) (*backend.CallResourceResponse, error) {
	kafka.Send(model.ConfigurationTopic, "renameDatapoint:1:"+strconv.FormatInt(orgId, 10), body)
	return &backend.CallResourceResponse{
		Status: http.StatusAccepted,
	}, nil
}
