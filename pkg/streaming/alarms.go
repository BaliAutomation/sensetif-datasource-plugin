package streaming

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/Sensetif/sensetif-datasource/pkg/model"
    "github.com/grafana/grafana-plugin-sdk-go/backend"
    "github.com/grafana/grafana-plugin-sdk-go/backend/log"
    "github.com/grafana/grafana-plugin-sdk-go/data"
    "strconv"
)

//func (h *StreamHandler) RunAlarmsStatusStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender, orgId int64) error {
func (h *StreamHandler) RunAlarmsStatusStream(_ context.Context, _ *backend.StreamSender, _ int64) error {
    return fmt.Errorf("Not implemented yet!")
}

const TOPIC_COMMANDS = "alarmcommands"

//type alarmCommand struct {
//    command string
//    args    map[string]string
//}

func (h *StreamHandler) RunAlarmCommandsStream(_ context.Context, req *backend.PublishStreamRequest, orgId int64, user string) (*backend.PublishStreamResponse, error) {
    key := "2:" + strconv.FormatInt(orgId, 10) + ":" + user
    h.pulsar.Send(TOPIC_COMMANDS, key, req.Data)
    return &backend.PublishStreamResponse{
        Status: backend.PublishStreamStatusOK,
    }, nil

}

func (h *StreamHandler) RunAlarmsHistoryStream(ctx context.Context, sender *backend.StreamSender, orgId int64) error {
    labelFrame := data.NewFrame("error",
        data.NewField("Time", nil, make([]int64, 1)),
        data.NewField("Class", nil, make([]string, 1)),
        data.NewField("Category", nil, make([]string, 1)),
        data.NewField("Project", nil, make([]string, 1)),
        data.NewField("Subsystem", nil, make([]string, 1)),
        data.NewField("Name", nil, make([]string, 1)),
        data.NewField("Description", nil, make([]string, 1)),
        data.NewField("", nil, make([]string, 1)),
    )
    reader := h.pulsar.CreateReader(model.NotificationTopics+strconv.FormatInt(orgId, 10), false)
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
        labelFrame.Fields[0].Set(0, notification.Time)
        labelFrame.Fields[1].Set(0, notification.Severity)
        labelFrame.Fields[2].Set(0, notification.Source)
        labelFrame.Fields[3].Set(0, notification.Key)
        labelFrame.Fields[4].Set(0, string(notification.Value))
        labelFrame.Fields[5].Set(0, notification.Message)
        labelFrame.Fields[6].Set(0, notification.Exception.Message)
        labelFrame.Fields[7].Set(0, notification.Exception.StackTrace)
        log.DefaultLogger.Info("Sending notification.")
        err = sender.SendFrame(labelFrame, data.IncludeAll)
        if err != nil {
            log.DefaultLogger.Error(fmt.Sprintf("Couldn't send frame: %v", err))
            return err
        }
    }
}

//func (h *StreamHandler) FindAlarmStates(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender, orgId int64) error {
// It is Ok to send one value in each frame, since there shouldn't be too many arriving, as that indicates misconfigured
// system and it lies in people's own interest to fix those. However, this could be revisited in future and sending
// batches of errors.
//labelFrame := data.NewFrame("error",
//    data.NewField("Time", nil, make([]int64, 1)),
//    data.NewField("Class", nil, make([]string, 1)),
//    data.NewField("Category", nil, make([]string, 1)),
//    data.NewField("Project", nil, make([]string, 1)),
//    data.NewField("Subsystem", nil, make([]string, 1)),
//    data.NewField("Name", nil, make([]string, 1)),
//    data.NewField("Description", nil, make([]string, 1)),
//    data.NewField("", nil, make([]string, 1)),
//)
//}

//func FormatAlarmStates(queryName string, alarmStates []model.TsPair) *data.Frame {
//    return &data.Frame{}
//}
