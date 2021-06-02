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
	clusterConfig *gocql.ClusterConfig
	session       *gocql.Session
	err           error
	ctx           context.Context
}

func (cass *CassandraClient) initializeCassandra(hosts []string) {
	log.DefaultLogger.Info("Initialize Cassandra client: " + hosts[0])
	cass.clusterConfig = gocql.NewCluster()
	cass.clusterConfig.Hosts = hosts
	cass.clusterConfig.Port = 9042
	cass.clusterConfig.HostFilter = gocql.HostFilterFunc(func(host *gocql.HostInfo) bool {
		log.DefaultLogger.Info("Filter: " + host.ConnectAddress().String() + ":" + strconv.Itoa(host.Port()) + " --> " + host.String())
		return true
	})
	cass.clusterConfig.Keyspace = "ks_sensetif"
	cass.reinitialize()
}

func (cass *CassandraClient) reinitialize() {
	log.DefaultLogger.Info("Re-initialize Cassandra session: " + fmt.Sprintf("%+v", cass.session))
	if cass.session != nil {
		cass.session.Close()
	}
	cass.session, cass.err = cass.clusterConfig.CreateSession()
	if cass.err != nil {
		log.DefaultLogger.Error("Unable to create Cassandra session: " + fmt.Sprintf("%+v", cass.err))
	}
	log.DefaultLogger.Info("Cassandra session: " + fmt.Sprintf("%+v", cass.session))
}

func (cass *CassandraClient) queryTimeseries(org int64, sensor SensorRef, from time.Time, to time.Time) []TsPair {
	log.DefaultLogger.Info("queryTimeseries:  " + strconv.FormatInt(org, 10) + "/" + sensor.project + "/" + sensor.subsystem + "/" + sensor.datapoint + "   " + from.Format(time.RFC3339) + "->" + to.Format(time.RFC3339))

	log.DefaultLogger.Info("1" + fmt.Sprintf("%+v", cass))
	cass.ctx = context.Background()
	log.DefaultLogger.Info("3")

	log.DefaultLogger.Info("Making a Cassandra SCAN query  " + fmt.Sprintf("%+v", cass))
	var result []TsPair
	startYearMonth := from.Year()*12 + int(from.Month())
	endYearMonth := to.Year()*12 + int(to.Month())
	for yearmonth := startYearMonth; yearmonth <= endYearMonth; yearmonth++ {
		var readValue []TsPair
		err := cass.session.Query(TsQuery, org, sensor.project, sensor.subsystem, yearmonth, sensor.datapoint, from, to).WithContext(cass.ctx).Scan(&readValue)

		log.DefaultLogger.Info("Query returned", len(readValue), err)
		if err != nil {
			log.DefaultLogger.Error("Unable to query timeseries", err)
			return nil
		} else {
			result = append(result, readValue...)
		}
	}
	log.DefaultLogger.Info("Found: %d datapoints", len(result))
	return result
}

func (cass *CassandraClient) shutdown() {
	log.DefaultLogger.Info("Shutdown Cassandra client")
	cass.session.Close()
}

const TimeseriesTablename = "timeseries"

const TsQuery = "SELECT * FROM " + TimeseriesTablename +
	" WHERE" +
	" orgId = ?" +
	" AND" +
	" project = ?" +
	" AND" +
	" subsystem = ?" +
	" AND" +
	" yearmonth = ?" +
	" AND" +
	" datapoint = ?" +
	" AND " +
	"ts >= ?" +
	" AND " +
	" ts <= ?" +
	";"
