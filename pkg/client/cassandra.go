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
	GetProject(orgId int64, name string) *model.ProjectSettings

	UpsertProject(orgId int64, project *model.ProjectSettings) error
	UpsertSubsystem(orgId int64, project *model.SubsystemSettings) error
	UpsertDatapoint(orgId int64, datapoint *model.DatapointSettings) error

	FindAllProjects(org int64) []model.ProjectSettings
	GetSubsystem(org int64, projectName string, subsystem string) *model.SubsystemSettings
	FindAllSubsystems(org int64, projectName string) []model.SubsystemSettings
	GetDatapoint(org int64, projectName string, subsystemName string, datapoint string) *model.DatapointSettings
	FindAllDatapoints(org int64, projectName string, subsystemName string) []model.DatapointSettings

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

func (cass *CassandraClient) InitializeCassandra(hosts []string) {
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

func (cass *CassandraClient) IsHealthy() bool {
	return !cass.session.Closed()
}

func (cass *CassandraClient) Reinitialize() {
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

func (cass *CassandraClient) QueryTimeseries(org int64, sensor model.SensorRef, from time.Time, to time.Time, maxValues int) []model.TsPair {
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
		query.Release()
	}
	resultLength := len(result)
	log.DefaultLogger.Info(fmt.Sprintf("Max: %d, found: %d datapoints", maxValues, resultLength))
	result = reduceSize(resultLength, maxValues, result)
	log.DefaultLogger.Info(fmt.Sprintf("Returning %d datapoints", len(result)))
	return result
}

func reduceSize(resultLength int, maxValues int, result []model.TsPair) []model.TsPair {
	if resultLength > maxValues && resultLength > 0 && maxValues > 0 {
		log.DefaultLogger.Info(fmt.Sprintf("Reducing datapoints from %d to %d", resultLength, maxValues))
		// Grafana has a MaxDatapoints expectations that we need to deal with
		var factor int
		factor = resultLength/maxValues + 1
		newSize := resultLength/factor + 1
		resultIndex := resultLength - 1
		var downsized = make([]model.TsPair, newSize, newSize)
		for i := newSize - 1; i >= 0; i = i - 1 {
			downsized[i] = result[resultIndex]
			// TODO; Should we have some type of function for this reduction?? Average, Min, Max?
			resultIndex = resultIndex - factor
		}
		return downsized
	}
	return result
}

func (cass *CassandraClient) GetProject(orgId int64, name string) *model.ProjectSettings {
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
		return &rowValue
	}
	return nil
}

func (cass *CassandraClient) UpsertProject(orgId int64, project *model.ProjectSettings) error {
	log.DefaultLogger.Info("addProject:  " + strconv.FormatInt(orgId, 10) + "/" + project.Name)
	return cass.session.Query(projectsInsert, orgId, project.Name, project.Title, project.City, project.Country, project.Timezone, project.Geolocation).Exec()
}

func (cass *CassandraClient) UpsertSubsystem(orgId int64, subsystem *model.SubsystemSettings) error {
	log.DefaultLogger.Info("addSubsystem:  " + strconv.FormatInt(orgId, 10) + "/" + subsystem.Name)
	return cass.session.Query(subsystemInsert, orgId, subsystem.Name, subsystem.Title, subsystem.Locallocation, time.Now(), subsystem.Project).Exec()
}

func (cass *CassandraClient) UpsertDatapoint(orgId int64, d *model.DatapointSettings) error {
	log.DefaultLogger.Info("addDatapoint:  " + strconv.FormatInt(orgId, 10) + "/" + d.Name)
	return cass.session.Query(datapointInsert, orgId, d.Project, d.Subsystem, d.Name, time.Now(), d.Interval, d.URL, d.Format,
		d.AuthenticationType, d.Auth, d.ValueExpression, d.Unit, d.TimestampExpression, d.TimestampType, d.TimeToLive, d.Scaling, d.K, d.M).Exec()
}

func (cass *CassandraClient) FindAllProjects(org int64) []model.ProjectSettings {
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

func (cass *CassandraClient) GetSubsystem(org int64, projectName string, subsystem string) *model.SubsystemSettings {
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
		return &rowValue
	}
	return nil
}

func (cass *CassandraClient) FindAllSubsystems(org int64, projectName string) []model.SubsystemSettings {
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

func (cass *CassandraClient) GetDatapoint(org int64, projectName string, subsystemName string, datapoint string) *model.DatapointSettings {
	log.DefaultLogger.Info("getDatapoint:  " + strconv.FormatInt(org, 10) + "/" + projectName + "/" + datapoint)
	scanner := cass.session.
		Query(fmt.Sprintf(datapointQuery, cass.clusterConfig.Keyspace, datapointsTablename), org, projectName, subsystemName, datapoint).
		Iter().
		Scanner()
	for scanner.Next() {
		var r model.DatapointSettings
		r.Project = projectName
		r.Project = subsystemName
		err := scanner.Scan(&r.Name, &r.Interval, &r.URL, &r.Format, &r.AuthenticationType, &r.Auth,
			&r.ValueExpression, &r.Unit, &r.TimestampExpression, &r.TimestampType, &r.TimeToLive, &r.Scaling, &r.K, &r.M)
		if err != nil {
			log.DefaultLogger.Error("Internal Error? Failed to read record", err)
		}
		return &r
	}
	return nil
}

func (cass *CassandraClient) FindAllDatapoints(org int64, projectName string, subsystemName string) []model.DatapointSettings {
	log.DefaultLogger.Info("findAllDatapoints:  " + strconv.FormatInt(org, 10) + "/" + projectName + "/" + subsystemName)
	result := make([]model.DatapointSettings, 0)
	query := fmt.Sprintf(datapointsQuery, cass.clusterConfig.Keyspace, datapointsTablename)
	scanner := cass.session.
		Query(query, org, projectName, subsystemName).
		Iter().
		Scanner()
	for scanner.Next() {
		var r model.DatapointSettings
		r.Project = projectName
		r.Project = subsystemName
		err := scanner.Scan(&r.Name, &r.Interval, &r.URL, &r.Format, &r.AuthenticationType, &r.Auth,
			&r.ValueExpression, &r.Unit, &r.TimestampExpression, &r.TimestampType, &r.TimeToLive, &r.Scaling, &r.K, &r.M)
		if err != nil {
			log.DefaultLogger.Error("Internal Error? Failed to read record: " + err.Error())
		}
		result = append(result, r)
	}
	log.DefaultLogger.Info(fmt.Sprintf("Found: %d datapoints", len(result)))
	return result
}

func (cass *CassandraClient) Shutdown() {
	log.DefaultLogger.Info("Shutdown Cassandra client")
	cass.session.Close()
}

func (cass *CassandraClient) Err() error {
	return cass.err
}

const projectsTablename = "projects"

const projectsInsert = "INSERT into projects (orgid,name,title,city,country,timezone,geolocation) values (?, ?, ?, ?, ?, ?, ?);"
const subsystemInsert = "INSERT into subsystems (orgid,name,title,location,ts,project) values (?, ?, ?, ?, ?, ?);"

const datapointInsert = "INSERT INTO datapoints (orgid,project,subsystem,name,ts,pollinterval,url,docformat,authtype,auth,valueexpression,unit,timeexpression,timestamptype,timetolive,scaling,k,m)" +
	"VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?);"

const projectQuery = "SELECT name,title,city,country,timezone,geolocation FROM %s.%s WHERE orgid = ? AND name = ?;"

const projectsQuery = "SELECT name,title,city,country,timezone,geolocation FROM %s.%s WHERE orgid = ?;"

const subsystemsTablename = "subsystems"

const subsystemQuery = "SELECT name,title,location FROM %s.%s WHERE orgid = ? AND project = ? AND name = ?;"

const subsystemsQuery = "SELECT name,title,location FROM %s.%s WHERE orgid = ? AND project = ?;"

const datapointsTablename = "datapoints"

const datapointQuery = "SELECT name,pollinterval,url,docformat,authtype,auth,valueexpression,unit,timeexpression,timestamptype,timetolive,scaling,k,m FROM %s.%s WHERE orgid = ? AND project = ? AND subsystem = ? AND name = ?;"

const datapointsQuery = "SELECT name,pollinterval,url,docformat,authtype,auth,valueexpression,unit,timeexpression,timestamptype,timetolive,scaling,k,m FROM %s.%s WHERE orgid = ? AND project = ? AND subsystem = ?;"

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
