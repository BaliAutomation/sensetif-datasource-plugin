package main

import (
	"context"
	"fmt"
	"github.com/gocql/gocql"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"strconv"
	"time"
)

type CassandraClient struct {
	cluster *gocql.ClusterConfig
	session *gocql.Session
	err     error
	ctx     context.Context
}

func (cass *CassandraClient) initializeCassandra(hosts []string) {
	log.DefaultLogger.Info("Initialize Cassandra client")
	cass.cluster = gocql.NewCluster(hosts[0])
	cass.cluster.Hosts = hosts
	cass.cluster.Keyspace = "ks_sensetif"
	cass.reinitialize()
}

func (cass *CassandraClient) reinitialize() {
	log.DefaultLogger.Info("Re-initialize Cassandra client: " + fmt.Sprintf("%+v", cass, cass.session))
	if cass.session != nil {
		cass.session.Close()
	}
	cass.session, cass.err = cass.cluster.CreateSession()
	if cass.err != nil {
		log.DefaultLogger.Error("Unable to create Cassandra session: " + fmt.Sprintf("%+v", cass.err))
	}
}

func (cass *CassandraClient) queryTimeseries(org int64, sensor SensorRef, from time.Time, to time.Time) []TsPair {
	log.DefaultLogger.Info("queryTimeseries:  " + strconv.FormatInt(org, 10) + "/" + sensor.project + "/" + sensor.subsystem + "/" + sensor.datapoint + "   " + from.Format(time.RFC3339) + "->" + to.Format(time.RFC3339))
	var readValue []TsPair

	log.DefaultLogger.Info("1" + fmt.Sprintf("%+v", cass))
	cass.ctx = context.Background()
	log.DefaultLogger.Info("3")

	log.DefaultLogger.Info("Making a Cassandra SCAN query  " + fmt.Sprintf("%+v", cass))
	err := cass.session.Query(TS_QUERY, org, sensor.project, sensor.subsystem, sensor.datapoint, from, to).WithContext(cass.ctx).Scan(&readValue)
	log.DefaultLogger.Info("Query returned", len(readValue), err)
	if err != nil {
		log.DefaultLogger.Error("Unable to query timeseries", err)
		return nil
	}
	log.DefaultLogger.Info("Found: %d datapoints", len(readValue))
	return readValue
}

func (cass *CassandraClient) shutdown() {
	log.DefaultLogger.Info("Shutdown Cassandra client")
	cass.session.Close()
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
	" datapoint = ?" +
	" AND " +
	"ts >= ?" +
	" AND " +
	" ts <= ?" +
	";"
