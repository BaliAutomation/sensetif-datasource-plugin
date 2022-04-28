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
func ListProjects(orgId int64, params []string, body []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
	log.DefaultLogger.Info("ListProjects()")
	projects := clients.Cassandra.FindAllProjects(orgId)
	rawJson, err := json.Marshal(projects)
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
func GetProject(orgId int64, params []string, body []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
	log.DefaultLogger.Info("GetProject()")
	project := clients.Cassandra.GetProject(orgId, params[1])
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
func UpdateProject(orgId int64, params []string, body []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
	log.DefaultLogger.Info("UpdateProject()")
	key := "updateProject:1:" + strconv.FormatInt(orgId, 10)
	log.DefaultLogger.Info(fmt.Sprintf("%+v", *clients.Pulsar))
	clients.Pulsar.Send(model.ConfigurationTopic, key, body)
	rawJson, _ := json.Marshal("ok")
	return &backend.CallResourceResponse{
		Status: http.StatusOK,
		Body:   rawJson,
	}, nil
}

//goland:noinspection GoUnusedParameter
func DeleteProject(orgId int64, params []string, body []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
	log.DefaultLogger.Info("DeleteProject()")
	if len(params) < 2 {
		return nil, fmt.Errorf("%w: missing params: \"%v\"", model.ErrBadRequest, params)
	}
	key := "deleteProject:1:" + strconv.FormatInt(orgId, 10)
	data, err := json.Marshal(map[string]string{
		"project": params[1],
	})
	if err == nil {
		clients.Pulsar.Send(model.ConfigurationTopic, key, data)
		return &backend.CallResourceResponse{
			Status: http.StatusAccepted,
		}, nil
	}
	return &backend.CallResourceResponse{
		Status: http.StatusBadRequest,
	}, nil
}

//goland:noinspection GoUnusedParameter
func RenameProject(orgId int64, params []string, body []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
	log.DefaultLogger.Info("RenameProject()")
	clients.Pulsar.Send(model.ConfigurationTopic, "renameProject:1:"+strconv.FormatInt(orgId, 10), body)
	return &backend.CallResourceResponse{
		Status: http.StatusAccepted,
	}, nil
}
