package streaming

import (
    "context"
    "fmt"
    "github.com/BaliAutomation/sensetif-datasource/pkg/model"
    "github.com/grafana/grafana-plugin-sdk-go/backend"
    "github.com/grafana/grafana-plugin-sdk-go/backend/log"
    "strconv"
    "time"
)

func (h *StreamHandler) RunNotificationsStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender, orgId int64) error {
    // It is Ok to send one value in each frame, since there shouldn't be too many arriving, as that indicates misconfigured
    // system and it lies in people's own interest to fix those. However, this could be revisited in future and sending
    // batches of errors.
    reader := h.pulsar.CreateReader(model.NotificationTopics + strconv.FormatInt(orgId, 10))
    defer reader.Close()
    thirty_minutes_ago := time.Now().Add(-30 * time.Minute)
    log.DefaultLogger.Info("Created Pulsar Reader. Start reading from " + thirty_minutes_ago.String())
    seekError := reader.SeekByTime(thirty_minutes_ago)
    if seekError != nil {
        log.DefaultLogger.Error(fmt.Sprintf("Unable to seek one hour back: %+v", seekError))
    }
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

        log.DefaultLogger.Info("Sending notification to " + req.PluginContext.User.Login)
        err = sender.SendJSON(msg.Payload())
        if err != nil {
            log.DefaultLogger.Error(fmt.Sprintf("Couldn't send frame: %v", err))
            return err
        }
    }
}
