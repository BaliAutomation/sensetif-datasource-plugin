package main

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// newDatasource returns datasource.ServeOpts.
func (td *SensetifDatasource) newDatasource(hosts []string) datasource.ServeOpts {
	im := datasource.NewInstanceManager(td.newDataSourceInstance)
	ds := &SensetifDatasource{
		im:    im,
		hosts: hosts,
	}

	return datasource.ServeOpts{
		QueryDataHandler:   ds,
		CheckHealthHandler: ds,
	}
}

type SensetifDatasource struct {
	im              instancemgmt.InstanceManager
	hosts           []string
	cassandraClient *CassandraClient
}

func (td *SensetifDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Info("QueryData", "request", req)
	orgId := req.PluginContext.OrgID
	response := backend.NewQueryDataResponse()
	for _, q := range req.Queries {
		res := td.query(ctx, orgId, q)
		response.Responses[q.RefID] = res
	}
	return response, nil
}

type queryModel struct {
	Format    string `json:"format"`
	Project   string `json:"project"`
	Subsystem string `json:"subsystem"`
	Datapoint string `json:"datapoint"`
}

func (td *SensetifDatasource) query(ctx context.Context, orgId int64, query backend.DataQuery) backend.DataResponse {
	response := backend.DataResponse{}
	var qm queryModel
	response.Error = json.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		return response
	}

	if qm.Format == "" {
		log.DefaultLogger.Warn("format is empty. defaulting to time series")
		qm.Format = "timeseries"
	}
	log.DefaultLogger.Info("format is " + qm.Format)

	from := query.TimeRange.From
	to := query.TimeRange.To

	datapoint := SensorRef{
		project:   qm.Project,
		subsystem: qm.Subsystem,
		sensor:    qm.Datapoint,
	}
	timeseries := td.cassandraClient.queryTimeseries(orgId, datapoint, from, to)

	times := []time.Time{}
	values := []float64{}
	for _, t := range timeseries {
		times = append(times, t.ts)
		values = append(values, t.value)
	}

	frame := data.NewFrame("response")
	frame.Fields = append(frame.Fields, data.NewField("time", nil, times))
	frame.Fields = append(frame.Fields, data.NewField("values", nil, values))
	response.Frames = append(response.Frames, frame)
	return response
}

func (td *SensetifDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	var status = backend.HealthStatusOk
	var message = "Data source is working"

	if rand.Int()%2 == 0 {
		status = backend.HealthStatusError
		message = "randomized error"
	}

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}

type instanceSettings struct {
	cassandraClient *CassandraClient
}

func (td *SensetifDatasource) newDataSourceInstance(setting backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	settings := &instanceSettings{
		cassandraClient: &CassandraClient{},
	}
	settings.cassandraClient.initializeCassandra(td.hosts)
	td.cassandraClient = settings.cassandraClient
	return settings, nil
}

func (s *instanceSettings) Dispose() {
	s.cassandraClient.shutdown()
}
