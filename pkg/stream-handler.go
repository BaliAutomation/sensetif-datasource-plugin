package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BaliAutomation/sensetif-datasource/pkg/client"
	"net"
	"strconv"
	"time"

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
	if req.Path == "_errors" {
		orgId := req.PluginContext.OrgID
		organization := strconv.FormatInt(orgId, 10)
		topic := model.ErrorsTopic + organization
		ip := getIpAddress()
		subscriptionName := "grafana-errors-" + organization + "-" + ip.String()
		if h.consumers[orgId] != nil {
			// We have already established the Pulsar read channel, and can simply return with an OK.
			return &backend.SubscribeStreamResponse{
				Status: backend.SubscribeStreamStatusOK,
			}, nil
		}
		consumer, err := h.pulsar.Subscribe(pulsar.ConsumerOptions{
			SubscriptionName: subscriptionName,
			Type:             pulsar.Exclusive,
			Topic:            model.ErrorNamespace + "/" + topic,
		})

		if err != nil {
			log.DefaultLogger.Error(fmt.Sprintf("failed to create pulsar consumer err: %v", err))
			return &backend.SubscribeStreamResponse{
				Status: backend.SubscribeStreamStatusNotFound,
			}, err
		}
		h.consumers[orgId] = consumer

		// Create the Go Routine that reads the Pulsar messages and publish to local Go channel per browser client
		pulsarChannel := (*consumer).Chan()
		subscribers := h.subscribers[orgId]
		if subscribers == nil {
			subscribers = []*chan *pulsar.ConsumerMessage{}
			h.subscribers[orgId] = subscribers
		}

		// We are creating a Go Routing that will read the Pulsar Consumer and feed the messages to
		// all registered Go Channels. Each such Channel is serving one websocket client, i.e. a Error Reporting
		// Tab in Grafana. I.e. We have a local fan-out of incoming Pulsar messages.
		go func() {
			select {
			case msg := <-pulsarChannel:
				var ch *chan *pulsar.ConsumerMessage
				for _, ch = range subscribers {
					*ch <- &msg
				}
			}
		}()

		return &backend.SubscribeStreamResponse{
			Status: backend.SubscribeStreamStatusOK,
		}, nil
	}
	return &backend.SubscribeStreamResponse{
		Status: backend.SubscribeStreamStatusNotFound,
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
	channel := make(chan *pulsar.ConsumerMessage)
	h.subscribers[orgId] = append(h.subscribers[orgId], &channel)

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

	for {
		select {
		case <-ctx.Done():
			index := find(h.subscribers[orgId], &channel)
			if index >= 0 {
				h.subscribers[orgId] = remove(h.subscribers[orgId], index)
			}
			close(channel)
			return ctx.Err()

		case msg := <-channel:
			log.DefaultLogger.Info("Received msg: %s", msg.Payload())
			var notification = Notification{}
			err := json.Unmarshal(msg.Payload(), &notification)
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
		}
	}
}

func getIpAddress() net.IP {
	conn, _ := net.Dial("udp", "10.20.0.1:80")
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

func remove(array []*chan *pulsar.ConsumerMessage, index int) []*chan *pulsar.ConsumerMessage {
	array[index] = array[len(array)-1]
	return array[:len(array)-1]
}

func find(array []*chan *pulsar.ConsumerMessage, channel *chan *pulsar.ConsumerMessage) int {
	for index, ch := range array {
		if ch == channel {
			return index
		}
	}
	return -1
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
