package streaming

import (
    "context"
    "fmt"
    "github.com/BaliAutomation/sensetif-datasource/pkg/model"
    "github.com/apache/pulsar-client-go/pulsar"
    "github.com/grafana/grafana-plugin-sdk-go/backend"
    "github.com/grafana/grafana-plugin-sdk-go/backend/log"
    "strconv"
    "time"
)

func (h *StreamHandler) SubscribeNotificationsStream(_ context.Context, _ *backend.SubscribeStreamRequest, orgId int64) (*backend.SubscribeStreamResponse, error) {

    reader := h.pulsar.CreateReader(model.NotificationTopics + strconv.FormatInt(orgId, 10), true)
    defer reader.Close()

    hourAgo := time.Now().Add(-60 * time.Minute)
    log.DefaultLogger.Info("Subscribing. Send last 60 minutes of messages " + hourAgo.String())
    seekError := reader.SeekByTime(hourAgo)
    if seekError != nil {
        log.DefaultLogger.Error(fmt.Sprintf("Unable to seek one hour back: %+v", seekError))
    }
    var err error
    var msg pulsar.Message
    var messages [][]byte
    var result []byte
    var count int32 = 0
    for reader.HasNext() {
        msg, err = reader.Next(context.Background())
        messages = append(messages, msg.Payload())
        count++
    }
    //messages = reverse(messages)
    result = append(result, '[')
    notFirst := false
    for i := 0; i < len(messages); i++ {
        if notFirst {
            result = append(result, ',')
        } else {
            notFirst = true
        }
        result = append(result, messages[i]...)
    }
    result = append(result, ']')
    initialData, err := backend.NewInitialData(result)
    if err == nil {
        log.DefaultLogger.Error(fmt.Sprintf("Sending %d messages", count))
        return &backend.SubscribeStreamResponse{
            Status:      backend.SubscribeStreamStatusOK,
            InitialData: initialData,
        }, nil
    }
    log.DefaultLogger.Error(fmt.Sprintf("Error in creating InitialData to be sent to client; %+v", err), err)
    log.DefaultLogger.Error("Error in: \n" + string(result))
    return &backend.SubscribeStreamResponse{
        Status:      backend.SubscribeStreamStatusPermissionDenied,
        InitialData: initialData,
    }, nil
}

func (h *StreamHandler) RunNotificationsStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender, orgId int64) error {
    // It is Ok to send one value in each frame, since there shouldn't be too many arriving, as that indicates misconfigured
    // system and it lies in people's own interest to fix those. However, this could be revisited in future and sending
    // batches of errors.
    reader := h.pulsar.CreateReader(model.NotificationTopics+strconv.FormatInt(orgId, 10), false)
    defer reader.Close()
    //log.DefaultLogger.Info("Created Pulsar Reader.")
    //minuteAgo := time.Now().Add(-1 * time.Minute)
    //seekError := reader.SeekByTime(minuteAgo)
    //if seekError != nil {
    //    log.DefaultLogger.Error(fmt.Sprintf("Unable to seek one minute back: %+v", seekError))
    //}
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

//func reverse(src [][]byte) [][]byte {
//    length := len(src)
//    nlen := length - 1
//    dest := make([][]byte, length)
//    for i := 0; i < length; i++ {
//        dest[i] = src[nlen-i]
//    }
//    return dest
//}
