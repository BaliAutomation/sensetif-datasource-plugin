package main

import (
	"context"
	"fmt"
	"time"

	"github.com/BaliAutomation/sensetif-datasource/pkg/model"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type pulsarReceiver interface {
	Subscribe(options pulsar.ConsumerOptions) (pulsar.Consumer, error)
}

type streamHandler struct {
	receiver pulsarReceiver
}

// SubscribeStream called when a user tries to subscribe to a plugin/datasource
// managed channel path – thus plugin can check subscribe permissions and communicate
// options with Grafana Core. As soon as first subscriber joins channel RunStream
// will be called.

func (h *streamHandler) SubscribeStream(ctx context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	// status := backend.SubscribeStreamStatusPermissionDenied
	// if req.Path == "stream" {
	// 	// Allow subscribing only on expected path.
	// 	status = backend.SubscribeStreamStatusOK
	// }
	// return &backend.SubscribeStreamResponse{
	// 	Status: status,
	// }, nil
	return &backend.SubscribeStreamResponse{
		Status: backend.SubscribeStreamStatusOK,
	}, nil
}

// PublishStream called when a user tries to publish to a plugin/datasource
// managed channel path. Here plugin can check publish permissions and
// modify publication data if required.
func (h *streamHandler) PublishStream(ctx context.Context, req *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}

// RunStream will be initiated by Grafana to consume a stream. RunStream will be
// called once for the first client successfully subscribed to a channel path.
// When Grafana detects that there are no longer any subscribers inside a channel,
// the call will be terminated until next active subscriber appears. Call termination
// can happen with a delay.
func (h *streamHandler) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
	log.DefaultLogger.Info("RunStream")
	consumer, err := h.receiver.Subscribe(pulsar.ConsumerOptions{
		SubscriptionName: "test-sub-1",
		Topic:            model.Namespace + "/" + model.MessagingTopic,
	})
	if err != nil {
		log.DefaultLogger.Error(fmt.Sprintf("failed to create pulsar consumer err: %v", err))
		return err
	}
	log.DefaultLogger.Info("created consumer")

	msgChannel := consumer.Chan()

	labelFrame := data.NewFrame("labeled",
		data.NewField("labels", nil, make([]string, 2)),
		data.NewField("Time", nil, make([]time.Time, 2)),
		data.NewField("Value", nil, make([]string, 2)),
	)

	for {
		select {
		case ctxErr := <-ctx.Done():
			log.DefaultLogger.Warn(fmt.Sprintf("Ctx error: %v", ctxErr))
			return ctx.Err()

		case msg := <-msgChannel:
			log.DefaultLogger.Info("Received msg: %s", msg.Payload())

			labelFrame.Fields[0].Set(0, fmt.Sprintf("s=A,s=p%d,x=X", 1))
			labelFrame.Fields[1].Set(0, time.Now())
			labelFrame.Fields[2].Set(0, string(msg.Payload()))

			err := sender.SendFrame(labelFrame, data.IncludeAll)
			if err != nil {
				log.DefaultLogger.Warn(fmt.Sprintf("Couldn't send frame: %v", err))
				return err
			}
			log.DefaultLogger.Info("Sent frame: %v")
		}
	}
}
