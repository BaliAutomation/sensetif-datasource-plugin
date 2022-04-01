package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type streamHandler struct {
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
	walker := rand.Float64() * 100

	labelFrame := data.NewFrame("labeled",
		data.NewField("labels", nil, make([]string, 2)),
		data.NewField("Time", nil, make([]time.Time, 2)),
		data.NewField("Value", nil, make([]float64, 2)),
	)

	rand.Seed(time.Now().UnixNano())

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.DefaultLogger.Debug("-- ctx error --")
			return ctx.Err()

		case t := <-ticker.C:
			log.DefaultLogger.Info(fmt.Sprintf("-- sending new frame --"))

			secA := t.Second() / 3
			secB := t.Second() / 7

			labelFrame.Fields[0].Set(0, fmt.Sprintf("s=A,s=p%d,x=X", secA))
			labelFrame.Fields[1].Set(0, t)
			labelFrame.Fields[2].Set(0, walker)

			labelFrame.Fields[0].Set(1, fmt.Sprintf("s=B,s=p%d,x=X", secB))
			labelFrame.Fields[1].Set(1, t)
			labelFrame.Fields[2].Set(1, walker+10)

			log.DefaultLogger.Info("Frame: %v", *labelFrame)
			// Send frame to stream including both frame schema and data frame parts.
			err := sender.SendFrame(labelFrame, data.IncludeAll)
			if err != nil {
				log.DefaultLogger.Warn(fmt.Sprintf("err sending frame:: %v", err))
				return err
			}
		}
	}
}
