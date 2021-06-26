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

func ListDatapoints(cmd *model.Command, cassandra client.Cassandra) (*backend.CallResourceResponse, error) {
	values, missingParams := getParams(cmd.Params, "project", "subsystem")
	if len(missingParams) > 0 {
		return nil, fmt.Errorf("%w: missing params: \"%v\"", model.ErrBadRequest, missingParams)
	}

	datapoints := cassandra.FindAllDatapoints(cmd.OrgID, values[0], values[1])
	rawJson, err := json.Marshal(datapoints)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", model.ErrUnprocessableEntity, err.Error())
	}

	return &backend.CallResourceResponse{
		Status: http.StatusOK,
		Body:   rawJson,
	}, nil
}

func GetDatapoint(cmd *model.Command, cassandra client.Cassandra) (*backend.CallResourceResponse, error) {
	values, missingParams := getParams(cmd.Params, "project", "subsystem", "datapoint")
	if len(missingParams) > 0 {
		return nil, fmt.Errorf("%w: missing params: \"%v\"", model.ErrBadRequest, missingParams)
	}

	datapoint := cassandra.GetDatapoint(cmd.OrgID, values[0], values[1], values[2])
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

func UpdateDatapoint(cmd *model.Command, cassandra client.Cassandra, kafka client.Kafka) (*backend.CallResourceResponse, error) {
	kafka.Send(model.ConfigurationTopic, "updateDatapoint:1:"+strconv.FormatInt(cmd.OrgID, 10), cmd.Payload)
	if util.IsDevelopmentMode() {
		var datapoint model.DatapointSettings
		if err := json.Unmarshal(cmd.Payload, &datapoint); err != nil {
			log.DefaultLogger.Error(fmt.Sprintf("Could not unmarshal datapoint; err: %v", err))
			return nil, fmt.Errorf("%w: invalid datapoint json", model.ErrBadRequest)
		}
		if err := cassandra.UpsertDatapoint(cmd.OrgID, &datapoint); err != nil {
			return nil, err
		}
	}
	return &backend.CallResourceResponse{
		Status: http.StatusAccepted,
	}, nil
}

func DeleteDatapoint(cmd *model.Command, kafka client.Kafka) (*backend.CallResourceResponse, error) {
	kafka.Send(model.ConfigurationTopic, "deleteDatapoint:1:"+strconv.FormatInt(cmd.OrgID, 10), cmd.Payload)
	return &backend.CallResourceResponse{
		Status: http.StatusAccepted,
	}, nil
}

func RenameDatapoint(cmd *model.Command, kafka client.Kafka) (*backend.CallResourceResponse, error) {
	kafka.Send(model.ConfigurationTopic, "renameDatapoint:1:"+strconv.FormatInt(cmd.OrgID, 10), cmd.Payload)
	return &backend.CallResourceResponse{
		Status: http.StatusAccepted,
	}, nil
}
