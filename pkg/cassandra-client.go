package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gocql/gocql"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
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
	cass.ctx = context.Background()
	log.DefaultLogger.Info("Cassandra session: " + fmt.Sprintf("%+v", cass.session))
}

func (cass *CassandraClient) queryTimeseries(org int64, sensor SensorRef, from time.Time, to time.Time) []TsPair {
	log.DefaultLogger.Info("queryTimeseries:  " + strconv.FormatInt(org, 10) + "/" + sensor.project + "/" + sensor.subsystem + "/" + sensor.datapoint + "   " + from.Format(time.RFC3339) + "->" + to.Format(time.RFC3339))
	var result []TsPair
	startYearMonth := from.Year()*12 + int(from.Month())
	endYearMonth := to.Year()*12 + int(to.Month())
	for yearmonth := startYearMonth; yearmonth <= endYearMonth; yearmonth++ {
		scanner := cass.session.
			Query(fmt.Sprintf(tsQuery, cass.clusterConfig.Keyspace, timeseriesTablename), org, sensor.project, sensor.subsystem, yearmonth, sensor.datapoint, from, to).
			Iter().
			Scanner()
		for scanner.Next() {
			var rowValue TsPair
			err := scanner.Scan(&rowValue.value, &rowValue.ts)
			if err != nil {
				log.DefaultLogger.Error("Internal Error? Failed to read record", err)
			}
			result = append(result, rowValue)
		}
	}
	log.DefaultLogger.Info(fmt.Sprintf("Found: %d datapoints", len(result)))
	return result
}

func (cass *CassandraClient) getProject(orgId int64, name string) *ProjectSettings {
	log.DefaultLogger.Info("getProject:  " + strconv.FormatInt(orgId, 10) + "/" + name)
	scanner := cass.session.
		Query(fmt.Sprintf(projectQuery, cass.clusterConfig.Keyspace, projectsTablename), orgId, name).
		Iter().
		Scanner()
	for scanner.Next() {
		var rowValue ProjectSettings
		err := scanner.Scan(&rowValue.Name, &rowValue.Title, &rowValue.City, &rowValue.Country, &rowValue.Timezone, &rowValue.Geolocation)
		if err != nil {
			log.DefaultLogger.Error("Internal Error? Failed to read record", err)
		}
		return &rowValue
	}
	return nil
}

func (cass *CassandraClient) findAllProjects(org int64) []ProjectSettings {
	log.DefaultLogger.Info("findAllProjects:  " + strconv.FormatInt(org, 10))
	var result []ProjectSettings
	scanner := cass.session.
		Query(fmt.Sprintf(projectsQuery, cass.clusterConfig.Keyspace, projectsTablename), org).
		Iter().
		Scanner()
	for scanner.Next() {
		var rowValue ProjectSettings
		err := scanner.Scan(&rowValue.Name, &rowValue.Title, &rowValue.City, &rowValue.Country, &rowValue.Timezone, &rowValue.Geolocation)
		if err != nil {
			log.DefaultLogger.Error("Internal Error? Failed to read record", err)
		}
		result = append(result, rowValue)
	}
	log.DefaultLogger.Info(fmt.Sprintf("Found: %d projects", len(result)))
	return result
}

func (cass *CassandraClient) getSubsystem(org int64, projectName string, subsystem string) *SubsystemSettings {
	log.DefaultLogger.Info("getSubsystem:  " + strconv.FormatInt(org, 10) + "/" + projectName + "/" + subsystem)
	scanner := cass.session.
		Query(fmt.Sprintf(subsystemQuery, cass.clusterConfig.Keyspace, subsystemsTablename), org, projectName, subsystem).
		Iter().
		Scanner()
	for scanner.Next() {
		var rowValue SubsystemSettings
		rowValue.Project = projectName
		err := scanner.Scan(&rowValue.Name, &rowValue.Title, &rowValue.Locallocation)
		if err != nil {
			log.DefaultLogger.Error("Internal Error? Failed to read record", err)
		}
		return &rowValue
	}
	return nil
}

func (cass *CassandraClient) findAllSubsystems(org int64, projectName string) []SubsystemSettings {
	log.DefaultLogger.Info("findAllSubsystems:  " + strconv.FormatInt(org, 10) + "/" + projectName)
	var result []SubsystemSettings
	scanner := cass.session.
		Query(fmt.Sprintf(subsystemsQuery, cass.clusterConfig.Keyspace, subsystemsTablename), org, projectName).
		Iter().
		Scanner()
	for scanner.Next() {
		var rowValue SubsystemSettings
		rowValue.Project = projectName
		err := scanner.Scan(&rowValue.Name, &rowValue.Title, &rowValue.Locallocation)
		if err != nil {
			log.DefaultLogger.Error("Internal Error? Failed to read record", err)
		}
		result = append(result, rowValue)
	}
	log.DefaultLogger.Info(fmt.Sprintf("Found: %d subsystems", len(result)))
	return result
}

func (cass *CassandraClient) getDatapoint(org int64, projectName string, subsystemName string, datapoint string) *DatapointSettings {
	log.DefaultLogger.Info("getDatapoint:  " + strconv.FormatInt(org, 10) + "/" + projectName + "/" + datapoint)
	scanner := cass.session.
		Query(fmt.Sprintf(datapointsQuery, cass.clusterConfig.Keyspace, datapointsTablename), org, projectName, subsystemName, datapoint).
		Iter().
		Scanner()
	for scanner.Next() {
		var r DatapointSettings
		r.Project = projectName
		r.Project = subsystemName
		err := scanner.Scan(&r.Name, &r.Interval, &r.URL, &r.Format, &r.AuthenticationType, &r.Credentials,
			&r.ValueExpression, &r.Unit, &r.TimestampExpression, &r.TimestampType, &r.TimeToLive, &r.Scaling, &r.K, &r.M)
		if err != nil {
			log.DefaultLogger.Error("Internal Error? Failed to read record", err)
		}
		return &r
	}
	return nil
}

func (cass *CassandraClient) findAllDatapoints(org int64, projectName string, subsystemName string) []DatapointSettings {
	log.DefaultLogger.Info("findAllSubsystems:  " + strconv.FormatInt(org, 10) + "/" + projectName)
	var result []DatapointSettings
	scanner := cass.session.
		Query(fmt.Sprintf(datapointsQuery, cass.clusterConfig.Keyspace, datapointsTablename), org, projectName, subsystemName).
		Iter().
		Scanner()
	for scanner.Next() {
		var r DatapointSettings
		r.Project = projectName
		r.Project = subsystemName
		err := scanner.Scan(&r.Name, &r.Interval, &r.URL, &r.Format, &r.AuthenticationType, &r.Credentials,
			&r.ValueExpression, &r.Unit, &r.TimestampExpression, &r.TimestampType, &r.TimeToLive, &r.Scaling, &r.K, &r.M)
		if err != nil {
			log.DefaultLogger.Error("Internal Error? Failed to read record", err)
		}
		result = append(result, r)
	}
	log.DefaultLogger.Info(fmt.Sprintf("Found: %d subsystems", len(result)))
	return result
}

func (cass *CassandraClient) shutdown() {
	log.DefaultLogger.Info("Shutdown Cassandra client")
	cass.session.Close()
}

const projectsTablename = "projects"

const projectQuery = "SELECT name,title,city,country,timezone,geolocation FROM %s.%s WHERE orgid = ? AND name = ?;"

const projectsQuery = "SELECT name,title,city,country,timezone,geolocation FROM %s.%s WHERE orgid = ?;"

const subsystemsTablename = "subsystems"

const subsystemQuery = "SELECT name,title,location FROM %s.%s WHERE orgid = ? AND project = ? AND name = ?;"

const subsystemsQuery = "SELECT name,title,location FROM %s.%s WHERE orgid = ? AND project = ?;"

const datapointsTablename = "datapoints"

const datapointQuery = "SELECT name,pollinterval,url,docformat,authtype,credentials,valueexpresssion,unit,timeexpression,timestamptype,timetolive,scalingfunction,k,m FROM %s.%s WHERE orgid = ? AND project = ? AND subsystem = ? AND name = ?;"

const datapointsQuery = "SELECT name,pollinterval,url,docformat,authtype,credentials,valueexpresssion,unit,timeexpression,timestamptype,timetolive,scalingfunction,k,m FROM %s.%s WHERE orgid = ? AND project = ? AND subsystem = ?;"

const timeseriesTablename = "timeseries"

const tsQuery = "SELECT value,ts FROM " + timeseriesTablename +
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
