package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BaliAutomation/sensetif-datasource/pkg/client"
	"github.com/BaliAutomation/sensetif-datasource/pkg/model"
	"github.com/apache/pulsar-client-go/pulsar"
	"strconv"
	"time"

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
	log.DefaultLogger.Info("SubscribeStream: " + req.Path)
	if req.Path != "_notifications" {
		return &backend.SubscribeStreamResponse{
			Status: backend.SubscribeStreamStatusNotFound,
		}, nil
	}
	return &backend.SubscribeStreamResponse{
		Status: backend.SubscribeStreamStatusOK,
	}, nil
}

func (h *streamHandler) PublishStream(_ context.Context, _ *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}

func (h *streamHandler) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
	log.DefaultLogger.Info("RunStream")
	orgId := req.PluginContext.OrgID

	// It is Ok to send one value in each frame, since there shouldn't be too many arriving, as that indicates misconfigured
	// system and it lies in people's own interest to fix those. However, this could be revisited in future and sending
	// batches of errors.
	labelFrame := data.NewFrame("error",
		data.NewField("Time", nil, make([]time.Time, 1)),
		data.NewField("Source", nil, make([]time.Time, 1)),
		data.NewField("Key", nil, make([]time.Time, 1)),
		data.NewField("Value", nil, make([]string, 1)),
		data.NewField("Message", nil, make([]string, 1)),
		data.NewField("ExceptionMessage", nil, make([]string, 1)),
		data.NewField("ExceptionStackTrace", nil, make([]string, 1)),
	)
	reader := h.pulsar.CreateReader(model.NotificationTopics + strconv.FormatInt(orgId, 10))
	defer reader.Close()

	for {
		msg, err := reader.Next(context.Background())
		log.DefaultLogger.Info("Received msg.")
		if err == nil {
			log.DefaultLogger.Info(fmt.Sprintf("    Message: %s", msg.Payload()))
			var notification = Notification{}
			err = json.Unmarshal(msg.Payload(), &notification)
			if err == nil {
				// Work for later; Refactor so that the serialization below is happening in the Pulsar message receiver
				// Go-routine to reduce work needed if more than one client is connected.
				labelFrame.Fields[0].Set(0, notification.Time)
				labelFrame.Fields[1].Set(0, notification.Source)
				labelFrame.Fields[2].Set(0, notification.Key)
				labelFrame.Fields[3].Set(0, string(notification.Value))
				labelFrame.Fields[4].Set(0, notification.Message)
				labelFrame.Fields[5].Set(0, notification.Exception.Message)
				labelFrame.Fields[6].Set(0, notification.Exception.StackTrace)
				err = sender.SendFrame(labelFrame, data.IncludeAll)
				if err != nil {
					log.DefaultLogger.Error(fmt.Sprintf("Couldn't send frame: %v", err))
					return err
				}
			} else {
				log.DefaultLogger.Error(fmt.Sprintf("Could not unmarshall json: %v", err))
			}
		} else {
			log.DefaultLogger.Error(fmt.Sprintf("Couldn't get the message via reader.Next(): %+v", err))
		}

		select {
		case <-ctx.Done():
			log.DefaultLogger.Info("Grafana sender: DONE")
			return ctx.Err()
		}
	}
}

type ExceptionDto struct {
	Message    string `json:"message"`
	StackTrace string `json:"stacktrace"`
}

type Notification struct {
	Time      time.Time    `json:"time"`
	Source    string       `json:"source"`
	Key       string       `json:"key"`
	Value     []byte       `json:"value"`
	Message   string       `json:"message"`
	Exception ExceptionDto `json:"exception"`
}
