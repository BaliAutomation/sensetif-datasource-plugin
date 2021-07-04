package main

import (
	"context"
	JSON "encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/BaliAutomation/sensetif-datasource/pkg/client"
	"github.com/BaliAutomation/sensetif-datasource/pkg/model"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

func (sds *SensetifDatasource) initializeInstance() {
	im := datasource.NewInstanceManager(sds.newDataSourceInstance)
	sds.im = im
}

type SensetifDatasource struct {
	im              instancemgmt.InstanceManager
	hosts           []string
	cassandraClient client.Cassandra
}

func (sds *SensetifDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Info(fmt.Sprintf("QueryData, ctx=%+v, req=%+v", ctx, req))
	orgId := req.PluginContext.OrgID
	response := backend.NewQueryDataResponse()
	for _, q := range req.Queries {
		res := sds.query(ctx, orgId, q)
		response.Responses[q.RefID] = res
	}
	return response, nil
}

type queryModel struct {
	Format     string `json:"format"`
	Parameters string `json:"parameters"`
}

func (sds *SensetifDatasource) query(ctx context.Context, orgId int64, query backend.DataQuery) backend.DataResponse {
	log.DefaultLogger.Info("query()")
	response := backend.DataResponse{}
	var qm queryModel
	response.Error = JSON.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		return response
	}

	if qm.Format == "" {
		log.DefaultLogger.Warn("format is empty. defaulting to time series")
		qm.Format = "timeseries"
	}

	log.DefaultLogger.Info("format is " + qm.Format)
	switch qm.Format {
	case "timeseries":
		return sds.executeTimeseriesQuery(qm.Parameters, orgId, query)
	}
	response.Error = fmt.Errorf("unknown Format: %s", qm.Format)
	return response
}

func (sds *SensetifDatasource) executeTimeseriesQuery(parameters string, orgId int64, query backend.DataQuery) backend.DataResponse {
	from := query.TimeRange.From
	to := query.TimeRange.To

	response := backend.DataResponse{}
	var model_ model.SensorRef
	response.Error = JSON.Unmarshal(query.JSON, &model_)
	if response.Error != nil {
		return response
	}

	log.DefaultLogger.Info("executeTimeseriesQuery(" + parameters + "," + strconv.FormatInt(orgId, 10) + ")")
	log.DefaultLogger.Info(fmt.Sprintf("Model: %+v", model_))
	timeseries := sds.cassandraClient.QueryTimeseries(orgId, model_, from, to)

	times := []time.Time{}
	values := []float64{}
	for _, t := range timeseries {
		times = append(times, t.TS)
		values = append(values, t.Value)
	}

	frame := data.NewFrame("response")
	frame.Fields = append(frame.Fields, data.NewField("time", nil, times))
	frame.Fields = append(frame.Fields, data.NewField(model_.Label, nil, values))
	response.Frames = append(response.Frames, frame)
	return response
}

func (sds *SensetifDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	log.DefaultLogger.Info("Check Health")
	var status = backend.HealthStatusOk
	var message = "Data source is working"
	// TODO; Make sure Cassandra is operational
	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}

type instanceSettings struct {
	cassandraClient client.Cassandra
}

func (sds *SensetifDatasource) newDataSourceInstance(setting backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	log.DefaultLogger.Info("newDataSourceInstance():\n\t" + fmt.Sprintf("Raw JSON;\n\t\t%s", string(setting.JSONData)))
	settings := &instanceSettings{
		cassandraClient: sds.cassandraClient,
	}
	settings.cassandraClient.Reinitialize()
	return settings, settings.cassandraClient.Err()
}

func (s *instanceSettings) Dispose() {
	s.cassandraClient.Shutdown()
}
