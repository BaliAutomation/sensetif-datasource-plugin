package streaming

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/BaliAutomation/sensetif-datasource/pkg/model"
    "github.com/grafana/grafana-plugin-sdk-go/backend"
    "github.com/grafana/grafana-plugin-sdk-go/backend/log"
    "github.com/grafana/grafana-plugin-sdk-go/data"
    "strconv"
    "time"
)

func (h *StreamHandler) RunNotificationsStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender, orgId int64) error {
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
    oneHourAgo := time.Now().Add(60 * time.Minute)
    seekError := reader.SeekByTime(oneHourAgo)
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
        var notification = Notification{}
        err = json.Unmarshal(msg.Payload(), &notification)
        if err != nil {
            log.DefaultLogger.Error(fmt.Sprintf("Could not unmarshall json: %v", err))
            continue
        }
        labelFrame.Fields[0].Set(0, notification.Time)
        labelFrame.Fields[1].Set(0, notification.Severity)
        labelFrame.Fields[2].Set(0, notification.Source)
        labelFrame.Fields[3].Set(0, notification.Key)
        labelFrame.Fields[4].Set(0, string(notification.Value))
        labelFrame.Fields[5].Set(0, notification.Message)
        labelFrame.Fields[6].Set(0, notification.Exception.Message)
        labelFrame.Fields[7].Set(0, notification.Exception.StackTrace)
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
