package main

import (
	"context"
	"github.com/gocql/gocql"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"time"
)

type CassandraClient struct {
	cluster *gocql.ClusterConfig
	session *gocql.Session
	err     error
	ctx     context.Context
}

func (td *CassandraClient) initializeCassandra(hosts []string) {
	td.cluster = gocql.NewCluster()
	td.cluster.Hosts = hosts
	td.cluster.Keyspace = "ks_sensetif"
	td.session, td.err = td.cluster.CreateSession()
}

func (td *CassandraClient) queryTimeseries(org int64, sensor SensorRef, from time.Time, to time.Time) []TsPair {
	var readValue []TsPair
	td.ctx = context.Background()
	err := td.session.Query(TS_QUERY, org, sensor.project, sensor.subsystem, sensor.sensor, from, to).WithContext(td.ctx).Scan(&readValue)
	if err != nil {
		log.DefaultLogger.Error("Unable to query timeseries", err)
		return nil
	}
	return readValue
}

func (td *CassandraClient) shutdown() {

}

const TIMESERIES_TABLENAME = "timeseries"

const TS_QUERY = "SELECT * FROM " + TIMESERIES_TABLENAME +
	" WHERE" +
	" organization = ?" +
	" AND" +
	" project = ?" +
	" AND" +
	" subsystem = ?" +
	" AND" +
	" sensor = ?" +
	" AND " +
	"ts >= ?" +
	" AND " +
	" ts <= ?" +
	";"
