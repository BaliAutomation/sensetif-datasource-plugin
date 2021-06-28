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

func ListProjects(cmd *model.Command, cassandra client.Cassandra) (*backend.CallResourceResponse, error) {
	projects := cassandra.FindAllProjects(cmd.OrgID)
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

func GetProject(cmd *model.Command, cassandra client.Cassandra) (*backend.CallResourceResponse, error) {
	values, missingParams := getParams(cmd.Params, "project")
	if len(missingParams) > 0 {
		return nil, fmt.Errorf("%w: missing params: \"%v\"", model.ErrBadRequest, missingParams)
	}

	project := cassandra.GetProject(cmd.OrgID, values[0])
	bytes, err := json.Marshal(project)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", model.ErrUnprocessableEntity, err.Error())
	}

	return &backend.CallResourceResponse{
		Status: http.StatusOK,
		Body:   bytes,
	}, nil
}

func UpdateProject(cmd *model.Command, cassandra client.Cassandra, kafka client.Kafka) (*backend.CallResourceResponse, error) {
	kafka.Send(model.ConfigurationTopic, "updateProject:1:"+strconv.FormatInt(cmd.OrgID, 10), cmd.Payload)

	if util.IsDevelopmentMode() {
		var project model.ProjectSettings
		if err := json.Unmarshal(cmd.Payload, &project); err != nil {
			log.DefaultLogger.Error(fmt.Sprintf("Could not unmarshal project; err: %v", err))
			return nil, fmt.Errorf("%w: invalid project json", model.ErrBadRequest)
		}
		if err := cassandra.UpsertProject(cmd.OrgID, &project); err != nil {
			return nil, err
		}
	}
	return &backend.CallResourceResponse{
		Status: http.StatusAccepted,
	}, nil
}

func DeleteProject(cmd *model.Command, kafka client.Kafka) (*backend.CallResourceResponse, error) {
	kafka.Send(model.ConfigurationTopic, "deleteProject:1:"+strconv.FormatInt(cmd.OrgID, 10), cmd.Payload)
	return &backend.CallResourceResponse{
		Status: http.StatusAccepted,
	}, nil
}

func RenameProject(cmd *model.Command, kafka client.Kafka) (*backend.CallResourceResponse, error) {
	kafka.Send(model.ConfigurationTopic, "renameProject:1:"+strconv.FormatInt(cmd.OrgID, 10), cmd.Payload)
	return &backend.CallResourceResponse{
		Status: http.StatusAccepted,
	}, nil
}
