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
func ListDatapoints(orgId int64, params []string, body []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
    if len(params) < 3 {
        return nil, fmt.Errorf("%w: missing params: \"%v\"", model.ErrBadRequest, params)
    }
    datapoints, err := clients.Cassandra.FindAllDatapoints(orgId, params[1], params[2])
    if err != nil {
        log.DefaultLogger.Error("Unable read datapoint.")
        return nil, fmt.Errorf("%w: %s", model.ErrUnprocessableEntity, err.Error())
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

func GetDatapoint(orgId int64, params []string, _ []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
    if len(params) < 4 {
        return nil, fmt.Errorf("%w: missing params: \"%v\"", model.ErrBadRequest, params)
    }
    datapoint, err := clients.Cassandra.GetDatapoint(orgId, params[1], params[2], params[3])
    if err != nil {
        return nil, fmt.Errorf("%w: %s", model.ErrUnprocessableEntity, err.Error())
    }
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
    key := "2:" + strconv.FormatInt(orgId, 10) + ":updateDatapoint"
    clients.Pulsar.Send(model.ConfigurationTopic, key, body)
    return &backend.CallResourceResponse{
        Status: http.StatusAccepted,
    }, nil
}

func DeleteDatapoint(orgId int64, params []string, _ []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
    if len(params) < 4 {
        return nil, fmt.Errorf("%w: missing params: \"%v\"", model.ErrBadRequest, params)
    }
    datapoint := model.DatapointIdentifier{
        OrgId:     orgId,
        Project:   params[1],
        Subsystem: params[2],
        Datapoint: params[3],
    }
    bytes, err := json.Marshal(datapoint)
    if err == nil {
        key := "2:" + strconv.FormatInt(orgId, 10) + ":deleteDatapoint"
        clients.Pulsar.Send(model.ConfigurationTopic, key, bytes)
        return &backend.CallResourceResponse{
            Status: http.StatusAccepted,
        }, nil
    }
    return nil, err
}

//goland:noinspection GoUnusedParameter
func RenameDatapoint(orgId int64, params []string, body []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
    key := "2:" + strconv.FormatInt(orgId, 10) + ":renameDatapoint"
    clients.Pulsar.Send(model.ConfigurationTopic, key, body)
    return &backend.CallResourceResponse{
        Status: http.StatusAccepted,
    }, nil
}
