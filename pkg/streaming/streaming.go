package streaming

import (
    "fmt"
    "github.com/grafana/grafana-plugin-sdk-go/backend"
    "github.com/grafana/grafana-plugin-sdk-go/backend/log"
    "github.com/sensetif/sensetif-datasource/pkg/client"
    "golang.org/x/net/context"
    "strconv"
)

type StreamHandler struct {
    pulsar    *client.PulsarClient
    cassandra *client.CassandraClient
}

func CreateStreamHandler(pulsarClient *client.PulsarClient) StreamHandler {
    return StreamHandler{
        pulsar: pulsarClient,
    }
}
func (h *StreamHandler) SubscribeStream(_ context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
    // Called once for each new Organization?? Or is once per Browser?? Or once per Browser tab??
    log.DefaultLogger.Info("SubscribeStream: " + req.Path + " from " + strconv.FormatInt(req.PluginContext.OrgID, 10) + ": " + req.PluginContext.User.Login)
    if req.Path == "_notifications" {
        return &backend.SubscribeStreamResponse{
            Status: backend.SubscribeStreamStatusOK,
        }, nil
    }
    if req.Path == "_alarms/status" {
        return &backend.SubscribeStreamResponse{
            Status: backend.SubscribeStreamStatusOK,
        }, nil
    }
    if req.Path == "_alarms/history" {
        return &backend.SubscribeStreamResponse{
            Status: backend.SubscribeStreamStatusOK,
        }, nil
    }
    log.DefaultLogger.Error(fmt.Sprintf("SubscribeStream requested unknown resource type: %s", req.Path))
    return &backend.SubscribeStreamResponse{
        Status: backend.SubscribeStreamStatusNotFound,
    }, nil
}

func (h *StreamHandler) PublishStream(_ context.Context, req *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
    log.DefaultLogger.Info("SubscribeStream: " + req.Path + " from " + strconv.FormatInt(req.PluginContext.OrgID, 10) + ":" + req.PluginContext.User.Login)
    return &backend.PublishStreamResponse{
        Status: backend.PublishStreamStatusPermissionDenied,
    }, nil
}

func (h *StreamHandler) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
    log.DefaultLogger.Info("RunStream from " + strconv.FormatInt(req.PluginContext.OrgID, 10) + ":" + req.PluginContext.User.Login)
    orgId := req.PluginContext.OrgID
    if req.Path == "_notifications" {
        return h.RunNotificationsStream(ctx, req, sender, orgId)
    }
    if req.Path == "_alarms/status" {
        return h.RunAlarmsStatusStream(ctx, req, sender, orgId)
    }
    if req.Path == "_alarms/history" {
        return h.RunAlarmsHistoryStream(ctx, req, sender, orgId)
    }
    return fmt.Errorf("Unknown request.")
}
