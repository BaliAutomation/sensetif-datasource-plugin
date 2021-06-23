package handler

import (
	"encoding/json"
	"fmt"
	"github.com/BaliAutomation/sensetif-datasource/pkg/util"
	"net/http"
	"strconv"

	"github.com/BaliAutomation/sensetif-datasource/pkg/client"
	"github.com/BaliAutomation/sensetif-datasource/pkg/model"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

func ListSubsystems(cmd *model.Command, cassandra client.Cassandra) (*backend.CallResourceResponse, error) {
	values, missingParams := getParams(cmd.Params, "project")
	if len(missingParams) > 0 {
		return nil, fmt.Errorf("%w: missing params: \"%v\"", model.ErrBadRequest, missingParams)
	}

	subsystems := cassandra.FindAllSubsystems(cmd.OrgID, values[0])
	rawJson, err := json.Marshal(subsystems)
	if err != nil {
		log.DefaultLogger.Error("Unable to marshal json")
		return nil, fmt.Errorf("%w: %s", model.ErrUnprocessableEntity, err.Error())
	}

	return &backend.CallResourceResponse{
		Status: http.StatusOK,
		Body:   rawJson,
	}, nil
}

func GetSubsystem(cmd *model.Command, cassandra client.Cassandra) (*backend.CallResourceResponse, error) {
	values, missingParams := getParams(cmd.Params, "project", "subsystem")
	if len(missingParams) > 0 {
		return nil, fmt.Errorf("%w: missing params: \"%v\"", model.ErrBadRequest, missingParams)
	}

	subsystem := cassandra.GetSubsystem(cmd.OrgID, values[0], values[1])
	bytes, err := json.Marshal(subsystem)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", model.ErrUnprocessableEntity, err.Error())
	}
	return &backend.CallResourceResponse{
		Status:  http.StatusOK,
		Headers: make(map[string][]string),
		Body:    bytes,
	}, nil

}

func UpdateSubsystem(cmd *model.Command, kafka client.Kafka) (*backend.CallResourceResponse, error) {
	kafka.Send(model.ConfigurationTopic, "updateSubsystem:1:"+strconv.FormatInt(cmd.OrgID, 10), cmd.Payload)
	if util.IsDevelopmentMode() {
		// TODO: direct update to cassandra
	}
	return &backend.CallResourceResponse{
		Status: http.StatusAccepted,
	}, nil
}

func DeleteSubsystem(cmd *model.Command, kafka client.Kafka) (*backend.CallResourceResponse, error) {
	kafka.Send(model.ConfigurationTopic, "deleteSubsystem:1:"+strconv.FormatInt(cmd.OrgID, 10), cmd.Payload)
	return &backend.CallResourceResponse{
		Status: http.StatusAccepted,
	}, nil
}
