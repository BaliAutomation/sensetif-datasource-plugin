package handler

import (
    "encoding/json"
    "fmt"
    "github.com/Sensetif/sensetif-datasource/pkg/client"
    "github.com/Sensetif/sensetif-datasource/pkg/model"
    "github.com/grafana/grafana-plugin-sdk-go/backend"
    "github.com/grafana/grafana-plugin-sdk-go/backend/log"
    "net/http"
    "strconv"
    "time"
)

type TsDatapoint struct {
    Organization int64     `json:"organization"`
    Project      string    `json:"project"`
    Subsystem    string    `json:"subsystem"`
    Name         string    `json:"name"`
    Timestamp    time.Time `json:"timestamp"`
    Value        float64   `json:"value"`
}

func UpdateTimeseries(orgId int64, params []string, body []byte, clients *client.Clients) (*backend.CallResourceResponse, error) {
    key := "2:" + strconv.FormatInt(orgId, 10) + ":" + params[1] + "/" + params[2] + "/" + params[3]
    log.DefaultLogger.Info("Timeseries update of: " + key)
    tspairs := []model.TsPair{}
    err := json.Unmarshal(body, &tspairs)
    log.DefaultLogger.Info("Timeseries: " + strconv.FormatInt(int64(len(tspairs)), 10))
    if err != nil {
        log.DefaultLogger.Error("Invalid format: " + err.Error())
        return &backend.CallResourceResponse{
            Status: http.StatusBadRequest,
        }, nil
    }
    for _, tspair := range tspairs {
        message := TsDatapoint{
            Organization: orgId,
            Project:      params[1],
            Subsystem:    params[2],
            Name:         params[3],
            Timestamp:    tspair.TS,
            Value:        tspair.Value,
        }
        msgjson, err2 := json.Marshal(message)
        if err2 == nil {
            clients.Pulsar.Send(model.TimeseriesTopic, key, msgjson)
            log.DefaultLogger.Info(fmt.Sprintf("Update sent for: %d:%s/%s/%s = %f", orgId, params[1], params[2], params[3], tspair.Value))
        }
    }
    return &backend.CallResourceResponse{
        Status: http.StatusAccepted,
    }, nil
}
