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
func ListSubsystems(orgId int64, params []string, body []byte, kafka client.Kafka, cassandra client.Cassandra) (*backend.CallResourceResponse, error) {
	if len(params) < 2 {
		return nil, fmt.Errorf("%w: missing params: \"%v\"", model.ErrBadRequest, params)
	}

	subsystems := cassandra.FindAllSubsystems(orgId, params[1])
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

//goland:noinspection GoUnusedParameter
func GetSubsystem(orgId int64, params []string, body []byte, kafka client.Kafka, cassandra client.Cassandra) (*backend.CallResourceResponse, error) {

	if len(params) < 3 {
		return nil, fmt.Errorf("%w: missing params: \"%v\"", model.ErrBadRequest, params)
	}

	subsystem := cassandra.GetSubsystem(orgId, params[1], params[2])
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

//goland:noinspection GoUnusedParameter
func UpdateSubsystem(orgId int64, params []string, body []byte, kafka client.Kafka, cassandra client.Cassandra) (*backend.CallResourceResponse, error) {
	kafka.Send(model.ConfigurationTopic, "updateSubsystem:1:"+strconv.FormatInt(orgId, 10), body)
	if util.IsDevelopmentMode() {
		var subsystem model.SubsystemSettings
		if err := json.Unmarshal(body, &subsystem); err != nil {
			log.DefaultLogger.Error(fmt.Sprintf("Could not unmarshal subsystem; err: %v", err))
			return nil, fmt.Errorf("%w: invalid subsystem json", model.ErrBadRequest)
		}
		if err := cassandra.UpsertSubsystem(orgId, subsystem); err != nil {
			return nil, err
		}
	}
	return &backend.CallResourceResponse{
		Status: http.StatusAccepted,
	}, nil
}

//goland:noinspection GoUnusedParameter
func DeleteSubsystem(orgId int64, params []string, body []byte, kafka client.Kafka, cassandra client.Cassandra) (*backend.CallResourceResponse, error) {
	if len(params) < 3 {
		return nil, fmt.Errorf("%w: missing params: \"%v\"", model.ErrBadRequest, params)
	}
	key := "deleteSubsystem:1:" + strconv.FormatInt(orgId, 10)
	data, err := json.Marshal(map[string]string{
		"project":   params[1],
		"subsystem": params[2],
	})
	if err == nil {
		kafka.Send(model.ConfigurationTopic, key, data)
		return &backend.CallResourceResponse{
			Status: http.StatusAccepted,
		}, nil
	}
	return &backend.CallResourceResponse{
		Status: http.StatusBadRequest,
	}, nil
}

//goland:noinspection GoUnusedParameter
func RenameSubsystem(orgId int64, params []string, body []byte, kafka client.Kafka, cassandra client.Cassandra) (*backend.CallResourceResponse, error) {
	if len(params) < 3 {
		return nil, fmt.Errorf("%w: missing params: \"%v\"", model.ErrBadRequest, params)
	}
	key := "renameSubsystem:1:" + strconv.FormatInt(orgId, 10)
	kafka.Send(model.ConfigurationTopic, key, body)
	return &backend.CallResourceResponse{
		Status: http.StatusAccepted,
	}, nil
}
