package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/BaliAutomation/sensetif-datasource/pkg/client"
	"github.com/BaliAutomation/sensetif-datasource/pkg/model"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type streamHandler struct {
	pulsar      *client.PulsarClient
	consumers   map[int64]*pulsar.Consumer
	subscribers map[int64][]*chan *pulsar.ConsumerMessage
}

func (h *streamHandler) SubscribeStream(_ context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	// Called once for each new Organization?? Or is once per Browser?? Or once per Browser tab??
	log.DefaultLogger.Info("SubscribeStream: " + req.Path + " from " + strconv.FormatInt(req.PluginContext.OrgID, 10) + ": " + req.PluginContext.User.Login)
	if req.Path != "_notifications" {
		return &backend.SubscribeStreamResponse{
			Status: backend.SubscribeStreamStatusNotFound,
		}, nil
	}
	log.DefaultLogger.Info("SubscribeStream: succeed")
	return &backend.SubscribeStreamResponse{
		Status: backend.SubscribeStreamStatusOK,
	}, nil
}

func (h *streamHandler) PublishStream(_ context.Context, req *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	log.DefaultLogger.Info("SubscribeStream: " + req.Path + " from " + strconv.FormatInt(req.PluginContext.OrgID, 10) + ":" + req.PluginContext.User.Login)
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}

func (h *streamHandler) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
	log.DefaultLogger.Info("RunStream from " + strconv.FormatInt(req.PluginContext.OrgID, 10) + ":" + req.PluginContext.User.Login)
	orgId := req.PluginContext.OrgID

	// It is Ok to send one value in each frame, since there shouldn't be too many arriving, as that indicates misconfigured
	// system and it lies in people's own interest to fix those. However, this could be revisited in future and sending
	// batches of errors.
	labelFrame := data.NewFrame("error",
		NewField("Time", nil, make([]int64, 1)),
		NewFilterableField("Severity", nil, make([]string, 1)),
		NewFilterableField("Source", nil, make([]string, 1)),
		NewFilterableField("Key", nil, make([]string, 1)),
		NewFilterableField("Value", nil, make([]string, 1)),
		NewFilterableField("Message", nil, make([]string, 1)),
		NewFilterableField("ExceptionMessage", nil, make([]string, 1)),
		NewFilterableField("ExceptionStackTrace", nil, make([]string, 1)),
	)

	reader := h.pulsar.CreateReader(model.NotificationTopics + strconv.FormatInt(orgId, 10))
	defer reader.Close()
	log.DefaultLogger.Info("Created Pulsar Reader.")

	for {
		// The provided Context is capturing the connection back to the browser, so when it is
		// Done(), then we should just exit the for loop.
		msg, err := reader.Next(ctx)
		if msg == nil {
			log.DefaultLogger.Info("Grafana sender: DONE")
			return ctx.Err()
		}
		if err != nil {
			log.DefaultLogger.Error(fmt.Sprintf("Couldn't get the message via reader.Next(): %+v", err))
			continue
		}
		notification := Notification{}
		err = json.Unmarshal(msg.Payload(), &notification)
		if err != nil {
			log.DefaultLogger.Error(fmt.Sprintf("Could not unmarshall json: %v", err))
			continue
		}

		labelFrame.SetRow(0,
			notification.Time,
			notification.Severity,
			notification.Source,
			notification.Key,
			string(notification.Value),
			notification.Message,
			notification.Exception.Message,
			notification.Exception.StackTrace,
		)

		log.DefaultLogger.Info("Sending notification to " + req.PluginContext.User.Login)
		err = sender.SendFrame(labelFrame, data.IncludeAll)
		if err != nil {
			log.DefaultLogger.Error(fmt.Sprintf("Couldn't send frame: %v", err))
			return err
		}
	}
}

func NewField(name string, labels data.Labels, values interface{}) *data.Field {
	return data.NewField(name, labels, values)
}

func NewFilterableField(name string, labels data.Labels, values interface{}) *data.Field {
	out := NewField(name, labels, values)

	out.SetConfig(
		&data.FieldConfig{
			Custom: map[string]interface{}{"filterable": true},
		})

	return out
}

type ExceptionDto struct {
	Message    string `json:"message"`
	StackTrace string `json:"stacktrace"`
}

type Notification struct {
	Time      int64        `json:"time"`
	Severity  string       `json:"severity"`
	Source    string       `json:"source"`
	Key       string       `json:"key"`
	Value     []byte       `json:"value"`
	Message   string       `json:"message"`
	Exception ExceptionDto `json:"exception"`
}
