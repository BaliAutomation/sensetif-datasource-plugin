package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/BaliAutomation/sensetif-datasource/pkg/client"
	"github.com/BaliAutomation/sensetif-datasource/pkg/model"
	"github.com/BaliAutomation/sensetif-datasource/pkg/util"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

//goland:noinspection GoUnusedParameter
func ListProjects(orgId int64, params []string, body []byte, kafka client.Kafka, cassandra client.Cassandra) (*backend.CallResourceResponse, error) {
	projects := cassandra.FindAllProjects(orgId)
	rawJson, err := json.Marshal(projects)
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
func GetProject(orgId int64, params []string, body []byte, kafka client.Kafka, cassandra client.Cassandra) (*backend.CallResourceResponse, error) {
	project := cassandra.GetProject(orgId, params[1])
	bytes, err := json.Marshal(project)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", model.ErrUnprocessableEntity, err.Error())
	}
	return &backend.CallResourceResponse{
		Status: http.StatusOK,
		Body:   bytes,
	}, nil
}

//goland:noinspection GoUnusedParameter
func UpdateProject(orgId int64, params []string, body []byte, kafka client.Kafka, cassandra client.Cassandra) (*backend.CallResourceResponse, error) {
	kafka.Send(model.ConfigurationTopic, "updateProject:1:"+strconv.FormatInt(orgId, 10), body)

	if util.IsDevelopmentMode() {
		var project model.ProjectSettings
		if err := json.Unmarshal(body, &project); err != nil {
			log.DefaultLogger.Error(fmt.Sprintf("Could not unmarshal project; err: %v", err))
			return nil, fmt.Errorf("%w: invalid project json", model.ErrBadRequest)
		}
		if err := cassandra.UpsertProject(orgId, project); err != nil {
			return nil, err
		}
	}
	return &backend.CallResourceResponse{
		Status: http.StatusAccepted,
	}, nil
}

//goland:noinspection GoUnusedParameter
func DeleteProject(orgId int64, params []string, body []byte, kafka client.Kafka, cassandra client.Cassandra) (*backend.CallResourceResponse, error) {
	if len(params) < 2 {
		return nil, fmt.Errorf("%w: missing params: \"%v\"", model.ErrBadRequest, params)
	}
	key := "deleteProject:1:" + strconv.FormatInt(orgId, 10)
	data, err := json.Marshal(map[string]string{
		"project": params[1],
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
func RenameProject(orgId int64, params []string, body []byte, kafka client.Kafka, cassandra client.Cassandra) (*backend.CallResourceResponse, error) {
	kafka.Send(model.ConfigurationTopic, "renameProject:1:"+strconv.FormatInt(orgId, 10), body)
	return &backend.CallResourceResponse{
		Status: http.StatusAccepted,
	}, nil
}
