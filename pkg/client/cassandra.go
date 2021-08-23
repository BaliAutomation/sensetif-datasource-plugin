package client

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/BaliAutomation/sensetif-datasource/pkg/model"
	"github.com/gocql/gocql"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

type Cassandra interface {
	QueryTimeseries(org int64, sensor model.SensorRef, from time.Time, to time.Time, maxValue int) []model.TsPair
	GetProject(orgId int64, name string) model.ProjectSettings

	UpsertProject(orgId int64, project model.ProjectSettings) error
	UpsertSubsystem(orgId int64, project model.SubsystemSettings) error
	UpsertDatapoint(orgId int64, datapoint model.Datapoint) error

	FindAllProjects(org int64) []model.ProjectSettings
	GetSubsystem(org int64, projectName string, subsystem string) model.SubsystemSettings
	FindAllSubsystems(org int64, projectName string) []model.SubsystemSettings
	GetDatapoint(org int64, projectName string, subsystemName string, datapoint string) model.Datapoint
	FindAllDatapoints(org int64, projectName string, subsystemName string) []model.Datapoint

	Shutdown()
	Reinitialize()
	Err() error
	IsHealthy() bool
}

type CassandraClient struct {
	clusterConfig *gocql.ClusterConfig
	session       *gocql.Session
	err           error
	ctx           context.Context
}

func (cass CassandraClient) InitializeCassandra(hosts []string) {
	log.DefaultLogger.Info("Initialize Cassandra client: " + hosts[0])
	cass.clusterConfig = gocql.NewCluster()
	cass.clusterConfig.Hosts = hosts
	cass.clusterConfig.Port = 9042
	cass.clusterConfig.HostFilter = gocql.HostFilterFunc(func(host *gocql.HostInfo) bool {
		log.DefaultLogger.Info("Filter: " + host.ConnectAddress().String() + ":" + strconv.Itoa(host.Port()) + " --> " + host.String())
		return true
	})
	cass.clusterConfig.Keyspace = "ks_sensetif"
	cass.Reinitialize()
}

func (cass CassandraClient) IsHealthy() bool {
	return !cass.session.Closed()
}

func (cass CassandraClient) Reinitialize() {
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

func (cass CassandraClient) QueryTimeseries(org int64, sensor model.SensorRef, from time.Time, to time.Time, maxValues int) []model.TsPair {
	log.DefaultLogger.Info("queryTimeseries:  " + strconv.FormatInt(org, 10) + "/" + sensor.Project + "/" + sensor.Subsystem + "/" + sensor.Datapoint + "   " + from.Format(time.RFC3339) + "->" + to.Format(time.RFC3339))
	var result []model.TsPair
	startYearMonth := from.Year()*12 + int(from.Month()) - 1
	endYearMonth := to.Year()*12 + int(to.Month()) - 1
	log.DefaultLogger.Info(fmt.Sprintf("yearMonths:  start=%d, end=%d", startYearMonth, endYearMonth))

	for yearmonth := startYearMonth; yearmonth <= endYearMonth; yearmonth++ {
		queryText := fmt.Sprintf(tsQuery, cass.clusterConfig.Keyspace, timeseriesTablename)
		query := cass.session.Query(queryText, org, sensor.Project, sensor.Subsystem, yearmonth, sensor.Datapoint, from, to)
		query.Idempotent(true)
		query.Consistency(gocql.One)
		scanner := query.Iter().Scanner()
		for scanner.Next() {
			var rowValue model.TsPair
			err := scanner.Scan(&rowValue.Value, &rowValue.TS)
			if err != nil {
				log.DefaultLogger.Error("Internal Error? Failed to read record", err)
			}
			result = append(result, rowValue)
		}
	}
	return reduceSize(maxValues, result)
}

func reduceSize(maxValues int, result []model.TsPair) []model.TsPair {
	resultLength := len(result)
	if resultLength > maxValues && resultLength > 0 && maxValues > 0 {
		log.DefaultLogger.Info(fmt.Sprintf("Reducing datapoints from %d to %d", resultLength, maxValues))
		// Grafana has a MaxDatapoints expectations that we need to deal with
		var factor int
		factor = resultLength/maxValues + 1
		newSize := resultLength / factor
		var downsized = make([]model.TsPair, newSize, newSize)
		resultIndex := resultLength - 1
		for i := newSize - 1; i >= 0; i = i - 1 {
			downsized[i] = result[resultIndex]
			// TODO; Should we have some type of function for this reduction?? Average, Min, Max?
			resultIndex = resultIndex - factor
		}
		log.DefaultLogger.Info(fmt.Sprintf("Reduced to %d", len(downsized)))
		return downsized
	}
	log.DefaultLogger.Info(fmt.Sprintf("Returning %d datapoints", len(result)))
	return result
}

func (cass CassandraClient) GetProject(orgId int64, name string) model.ProjectSettings {
	log.DefaultLogger.Info("getProject:  " + strconv.FormatInt(orgId, 10) + "/" + name)
	scanner := cass.session.
		Query(fmt.Sprintf(projectQuery, cass.clusterConfig.Keyspace, projectsTablename), orgId, name).
		Iter().
		Scanner()
	for scanner.Next() {
		var rowValue model.ProjectSettings
		err := scanner.Scan(&rowValue.Name, &rowValue.Title, &rowValue.City, &rowValue.Country, &rowValue.Timezone, &rowValue.Geolocation)
		if err != nil {
			log.DefaultLogger.Error("Internal Error? Failed to read record", err)
		}
		return rowValue
	}
	return model.ProjectSettings{}
}

func (cass CassandraClient) UpsertProject(orgId int64, project model.ProjectSettings) error {
	log.DefaultLogger.Info("addProject:  " + strconv.FormatInt(orgId, 10) + "/" + project.Name)
	return cass.session.Query(projectsInsert, orgId, project.Name, project.Title, project.City, project.Country, project.Timezone, project.Geolocation).Exec()
}

func (cass CassandraClient) UpsertSubsystem(orgId int64, subsystem model.SubsystemSettings) error {
	log.DefaultLogger.Info("addSubsystem:  " + strconv.FormatInt(orgId, 10) + "/" + subsystem.Name)
	return cass.session.Query(subsystemInsert, orgId, subsystem.Name, subsystem.Title, subsystem.Locallocation, time.Now(), subsystem.Project).Exec()
}

func (cass CassandraClient) UpsertDatapoint(orgId int64, datapoint model.Datapoint) error {
	log.DefaultLogger.Info("addDatapoint:  " + strconv.FormatInt(orgId, 10) + "/" + (datapoint).Name())
	sourceType := (datapoint).SourceType()
	switch sourceType {
	case model.Web:
		d := (datapoint).(model.WebDocument)
		datasource := make(map[string]string)
		datasource["url"] = d.URL
		datasource["authtype"] = strconv.Itoa(int(d.AuthenticationType))
		datasource["auth"] = d.Auth
		datasource["docformat"] = strconv.Itoa(int(d.Format))
		datasource["valueexpression"] = d.ValueExpression
		datasource["timestamptype"] = strconv.Itoa(int(d.TimestampType))
		datasource["timeexpression"] = d.TimestampExpression
		// (orgid,project,subsystem,name,ts,pollinterval,datasourcetype,datasource,unit,timetolive,scaling,k,m)
		return cass.session.Query(datapointInsert, orgId, d.Project, d.Subsystem, d.Name, time.Now(), d.Interval, sourceType, datasource, d.Unit, d.TimeToLive, d.Scaling, d.K, d.M).Exec()
	case model.Ttnv3:
		d := (datapoint).(model.Ttnv3Document)
		datasource := make(map[string]string)
		datasource["zone"] = d.Zone
		datasource["application"] = d.Application
		datasource["device"] = d.Device
		datasource["pointname"] = d.PointName
		datasource["authorizationKey"] = d.AuthorizationKey
		// (orgid,project,subsystem,name,ts,pollinterval,datasourcetype,datasource,unit,timetolive,scaling,k,m)
		return cass.session.Query(datapointInsert, orgId, d.Project, d.Subsystem, d.Name, time.Now(), d.Interval, sourceType, datasource, d.Unit, d.TimeToLive, d.Scaling, d.K, d.M).Exec()
	}
	return nil
}

func (cass CassandraClient) FindAllProjects(org int64) []model.ProjectSettings {
	log.DefaultLogger.Info("findAllProjects:  " + strconv.FormatInt(org, 10))
	result := make([]model.ProjectSettings, 0)
	scanner := cass.session.
		Query(fmt.Sprintf(projectsQuery, cass.clusterConfig.Keyspace, projectsTablename), org).
		Iter().
		Scanner()
	for scanner.Next() {
		var rowValue model.ProjectSettings
		err := scanner.Scan(&rowValue.Name, &rowValue.Title, &rowValue.City, &rowValue.Country, &rowValue.Timezone, &rowValue.Geolocation)
		if err != nil {
			log.DefaultLogger.Error("Internal Error? Failed to read record", err)
		}
		result = append(result, rowValue)
	}
	log.DefaultLogger.Info(fmt.Sprintf("Found: %d projects", len(result)))
	return result
}

func (cass CassandraClient) GetSubsystem(org int64, projectName string, subsystem string) model.SubsystemSettings {
	log.DefaultLogger.Info("getSubsystem:  " + strconv.FormatInt(org, 10) + "/" + projectName + "/" + subsystem)
	scanner := cass.session.
		Query(fmt.Sprintf(subsystemQuery, cass.clusterConfig.Keyspace, subsystemsTablename), org, projectName, subsystem).
		Iter().
		Scanner()
	for scanner.Next() {
		var rowValue model.SubsystemSettings
		rowValue.Project = projectName
		err := scanner.Scan(&rowValue.Name, &rowValue.Title, &rowValue.Locallocation)
		if err != nil {
			log.DefaultLogger.Error("Internal Error? Failed to read record", err)
		}
		return rowValue
	}
	return model.SubsystemSettings{}
}

func (cass CassandraClient) FindAllSubsystems(org int64, projectName string) []model.SubsystemSettings {
	log.DefaultLogger.Info("findAllSubsystems:  " + strconv.FormatInt(org, 10) + "/" + projectName)
	result := make([]model.SubsystemSettings, 0)
	scanner := cass.session.
		Query(fmt.Sprintf(subsystemsQuery, cass.clusterConfig.Keyspace, subsystemsTablename), org, projectName).
		Iter().
		Scanner()
	for scanner.Next() {
		var rowValue model.SubsystemSettings
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

func (cass CassandraClient) GetDatapoint(org int64, projectName string, subsystemName string, datapoint string) model.Datapoint {
	log.DefaultLogger.Info("getDatapoint:  " + strconv.FormatInt(org, 10) + "/" + projectName + "/" + datapoint)
	scanner := cass.session.
		Query(fmt.Sprintf(datapointQuery, cass.clusterConfig.Keyspace, datapointsTablename), org, projectName, subsystemName, datapoint).
		Iter().
		Scanner()
	for scanner.Next() {
		return cass.deserializeRow(scanner)
	}
	return nil
}

func (cass CassandraClient) deserializeRow(scanner gocql.Scanner) model.Datapoint {
	var sourceType model.SourceType
	err := scanner.Scan(&sourceType)
	if err == nil {

		switch sourceType {
		case model.Web:
			var r model.WebDocument
			var properties = make(map[string]string)
			// (orgid,project,subsystem,name,ts,pollinterval,datasourcetype,datasource,unit,timetolive,scaling,k,m)
			err = scanner.Scan(&r.Project_, &r.Subsystem_, &r.Name_, &r.Interval_, &r.SourceType_, &properties, &r.Unit_, &r.TimeToLive, &r.Scaling, &r.K, &r.M)
			if err == nil {
				r.URL = properties["url"]

				var value, err = strconv.Atoi(properties["authtype"])
				if err != nil {
					break
				}
				r.AuthenticationType = model.AuthenticationType(value)
				r.Auth = properties["auth"]
				value, err = strconv.Atoi(properties["docformat"])
				if err != nil {
					break
				}
				r.Format = model.OriginDocumentFormat(value)
				r.ValueExpression = properties["valueexpression"]
				value, err = strconv.Atoi(properties["timestamptype"])
				if err != nil {
					break
				}
				r.TimestampType = model.TimestampType(value)
				r.TimestampExpression = properties["timeexpression"]
			}
			return r
		case model.Ttnv3:
			var t model.Ttnv3Document
			var properties = make(map[string]string)
			// (orgid,project,subsystem,name,ts,pollinterval,datasourcetype,datasource,unit,timetolive,scaling,k,m)
			err = scanner.Scan(&t.Project_, &t.Subsystem_, &t.Name_, &t.Interval_, &t.SourceType_, &properties, &t.Unit_, &t.TimeToLive, &t.Scaling, &t.K, &t.M)
			if err == nil {
				t.Zone = properties["zone"]
				t.Application = properties["application"]
				t.Device = properties["device"]
				t.PointName = properties["pointname"]
				t.AuthorizationKey = properties["authorizationKey"]
			}
			return t
		}
	}
	if err != nil {
		log.DefaultLogger.Error("Internal Error? Failed to read record", err)
	}
	return nil
}

func (cass CassandraClient) FindAllDatapoints(org int64, projectName string, subsystemName string) []model.Datapoint {
	log.DefaultLogger.Info("findAllDatapoints:  " + strconv.FormatInt(org, 10) + "/" + projectName + "/" + subsystemName)
	result := make([]model.Datapoint, 0)
	query := fmt.Sprintf(datapointsQuery, cass.clusterConfig.Keyspace, datapointsTablename)
	scanner := cass.session.
		Query(query, org, projectName, subsystemName).
		Iter().
		Scanner()
	for scanner.Next() {
		datapoint := cass.deserializeRow(scanner)
		result = append(result, datapoint)
	}
	log.DefaultLogger.Info(fmt.Sprintf("Found: %d datapoints", len(result)))
	return result
}

func (cass CassandraClient) Shutdown() {
	log.DefaultLogger.Info("Shutdown Cassandra client")
	cass.session.Close()
}

func (cass CassandraClient) Err() error {
	return cass.err
}

const projectsTablename = "projects"

const projectsInsert = "INSERT into projects (orgid,name,title,city,country,timezone,geolocation) values (?, ?, ?, ?, ?, ?, ?);"
const subsystemInsert = "INSERT into subsystems (orgid,name,title,location,ts,project) values (?, ?, ?, ?, ?, ?);"

const datapointInsert = "INSERT INTO %s.%s (orgid,project,subsystem,name,ts,pollinterval,datasourcetype,datasource,unit,timetolive,scaling,k,m)" +
	" VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?);"

const projectQuery = "SELECT name,title,city,country,timezone,geolocation FROM %s.%s WHERE orgid = ? AND name = ?;"

const projectsQuery = "SELECT name,title,city,country,timezone,geolocation FROM %s.%s WHERE orgid = ?;"

const subsystemsTablename = "subsystems"

const subsystemQuery = "SELECT name,title,location FROM %s.%s WHERE orgid = ? AND project = ? AND name = ?;"

const subsystemsQuery = "SELECT name,title,location FROM %s.%s WHERE orgid = ? AND project = ?;"

const datapointsTablename = "datapoints"

const datapointQuery = "SELECT orgid,project,subsystem,name,ts,pollinterval,datasourcetype,datasource,unit,timetolive,scaling,k,m FROM %s.%s WHERE orgid = ? AND project = ? AND subsystem = ? AND name = ?;"

const datapointsQuery = "SELECT orgid,project,subsystem,name,ts,pollinterval,datasourcetype,datasource,unit,timetolive,scaling,k,m FROM %s.%s WHERE orgid = ? AND project = ? AND subsystem = ?;"

const timeseriesTablename = "timeseries"

const tsQuery = "SELECT value,ts FROM %s.%s" +
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
