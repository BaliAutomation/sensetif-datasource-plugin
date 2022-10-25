package handler

import (
    "github.com/Sensetif/sensetif-datasource/pkg/client"
    "github.com/Sensetif/sensetif-datasource/pkg/model"
    "github.com/grafana/grafana-plugin-sdk-go/backend"
    "net/http"
    "strconv"
)

func ImportLink2WebFvc1(orgId int64, _ []string, body []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
    clients.Pulsar.Send(model.ConfigurationTopic, "importLink2WebFvc1:1:"+strconv.FormatInt(orgId, 10), body)
    return &backend.CallResourceResponse{
        Status: http.StatusAccepted,
    }, nil
}

func ImportTtnv3App(orgId int64, _ []string, body []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
    clients.Pulsar.Send(model.ConfigurationTopic, "importTtnv3App:1:"+strconv.FormatInt(orgId, 10), body)
    return &backend.CallResourceResponse{
        Status: http.StatusAccepted,
    }, nil
}
