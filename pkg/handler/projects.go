package handler

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"

    "github.com/Sensetif/sensetif-datasource/pkg/client"
    "github.com/Sensetif/sensetif-datasource/pkg/model"
    "github.com/grafana/grafana-plugin-sdk-go/backend"
    "github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

//goland:noinspection GoUnusedParameter
func ListProjects(orgId int64, params []string, body []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
    log.DefaultLogger.Info("ListProjects()")
    projects, err := clients.Cassandra.FindAllProjects(orgId)
    if err != nil {
        log.DefaultLogger.Error("Unable to read project.")
        return nil, fmt.Errorf("%w: %s", model.ErrUnprocessableEntity, err.Error())
    }
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
    project, err := clients.Cassandra.GetProject(orgId, params[1])
    if err != nil {
        return nil, fmt.Errorf("%w: %s", model.ErrUnprocessableEntity, err.Error())
    }
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
    key := "2:" + strconv.FormatInt(orgId, 10) + ":updateProject"
    log.DefaultLogger.Info(fmt.Sprintf("%+v", *clients.Pulsar))
    clients.Pulsar.Send(model.ConfigurationTopic, key, body)
    return &backend.CallResourceResponse{
        Status: http.StatusAccepted,
    }, nil
}

//goland:noinspection GoUnusedParameter
func DeleteProject(orgId int64, params []string, body []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
    log.DefaultLogger.Info("DeleteProject()")
    if len(params) < 2 {
        return nil, fmt.Errorf("%w: missing params: \"%v\"", model.ErrBadRequest, params)
    }
    key := "2:" + strconv.FormatInt(orgId, 10) + ":deleteProject"
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
    key := "2:" + strconv.FormatInt(orgId, 10) + ":renameProject"
    clients.Pulsar.Send(model.ConfigurationTopic, key, body)
    return &backend.CallResourceResponse{
        Status: http.StatusAccepted,
    }, nil
}
