package handler

import (
    "encoding/json"
    "fmt"
    "net/http"

    "github.com/Sensetif/sensetif-datasource/pkg/client"
    "github.com/Sensetif/sensetif-datasource/pkg/model"
    "github.com/grafana/grafana-plugin-sdk-go/backend"
    "github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

//goland:noinspection GoUnusedParameter
func GetOrganization(orgId int64, params []string, body []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
    log.DefaultLogger.Info("GetOrganization")
    organization, err := clients.Cassandra.GetOrganization(orgId)
    if err != nil {
        log.DefaultLogger.Error("Unable to read organization")
        return nil, fmt.Errorf("%w: %s", model.ErrUnprocessableEntity, err.Error())
    }
    rawJson, err := json.Marshal(organization)
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
