package client

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Sensetif/sensetif-datasource/pkg/model"
	"github.com/gocql/gocql"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

type Cassandra interface {
	QueryTimeseries(org int64, sensor model.SensorRef, from time.Time, to time.Time, maxValue int) []model.TsPair
	FindAllProjects(org int64) []model.ProjectSettings
	FindAllSubsystems(org int64, projectName string) []model.SubsystemSettings
	FindAllDatapoints(org int64, projectName string, subsystemName string) []model.DatapointSettings
	GetOrganization(orgId int64) model.OrganizationSettings
	GetProject(orgId int64, name string) model.ProjectSettings
	GetSubsystem(org int64, projectName string, subsystem string) model.SubsystemSettings
	GetDatapoint(org int64, projectName string, subsystemName string, datapoint string) model.DatapointSettings

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
	cass.clusterConfig.Keyspace = "ks_sensetif"
	cass.clusterConfig.Hosts = hosts
	cass.clusterConfig.Port = 9042
	cass.clusterConfig.HostFilter = gocql.HostFilterFunc(func(host *gocql.HostInfo) bool {
		log.DefaultLogger.Info("Filter: " + host.ConnectAddress().String() + ":" + strconv.Itoa(host.Port()) + " --> " + host.String())
		return true
	})
	cass.Reinitialize()
}

func (cass *CassandraClient) IsHealthy() bool {
	return !cass.session.Closed()
}

func (cass *CassandraClient) Reinitialize() {
	log.DefaultLogger.Info("Re-initialize Cassandra session: " + fmt.Sprintf("%+v", cass.session) + "," + fmt.Sprintf("%+v", cass.clusterConfig))
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
	//log.DefaultLogger.Info("queryTimeseries:  " + strconv.FormatInt(org, 10) + "/" + sensor.Project + "/" + sensor.Subsystem + "/" + sensor.Datapoint + "   " + from.Format(time.RFC3339) + "->" + to.Format(time.RFC3339))
	var result []model.TsPair
	startYearMonth := from.Year()*12 + int(from.Month()) - 1
	endYearMonth := to.Year()*12 + int(to.Month()) - 1
	//log.DefaultLogger.Info(fmt.Sprintf("yearMonths:  start=%d, end=%d", startYearMonth, endYearMonth))

	for yearmonth := endYearMonth; yearmonth >= startYearMonth; yearmonth-- {
		scanner := cass.createQuery(timeseriesTablename, tsQuery, org, sensor.Project, sensor.Subsystem, yearmonth, sensor.Datapoint, from, to)
		for scanner.Next() {
			var rowValue model.TsPair
			err := scanner.Scan(&rowValue.Value, &rowValue.TS)
			if err != nil {
				log.DefaultLogger.Error("Internal Error? Failed to read record", err)
			}
			p := []model.TsPair{rowValue}
			result = append(p, result...)
		}
	}
	return reduceSize(maxValues, result)
}

func (cass *CassandraClient) GetCurrentLimits(orgId int64) model.PlanLimits {

	//log.DefaultLogger.Info("GetCurrentLimits for " + strconv.FormatInt(orgId, 10))
	scanner := cass.createQuery(planlimitsTablename, planlimitsQuery, orgId)
	var limits model.PlanLimits
	limits.MaxStorage = "b"
	limits.MaxDatapoints = 50
	limits.MinPollInterval = "one_hour"
	for scanner.Next() {
		err := scanner.Scan(&limits.MaxDatapoints, &limits.MaxStorage, &limits.MinPollInterval)
		if err != nil {
			log.DefaultLogger.Error("Internal Error? Failed to read record", err)
		}
	}
	return limits
}

func (cass *CassandraClient) GetOrganization(orgId int64) model.OrganizationSettings {
	log.DefaultLogger.Info("getOrganization:  " + strconv.FormatInt(orgId, 10))
	//SELECT name,email,stripecustomer,currentplan,address1,address2,zipcode,city,state,country FROM %s.%s WHERE orgid = ? AND DELETED = '1970-01-01 0:00:00+0000';"
	scanner := cass.createQuery(organizationsTablename, organizationQuery, orgId)
	for scanner.Next() {
		var org model.OrganizationSettings
		err := scanner.Scan(&org.Name, &org.Email, &org.StripeCustomer, &org.CurrentPlan)
		if err != nil {
			log.DefaultLogger.Error("Internal Error? Failed to read record", err)
		}
		return org
	}
	return model.OrganizationSettings{}
}

func (cass *CassandraClient) GetProject(orgId int64, name string) model.ProjectSettings {
	log.DefaultLogger.Info("getProject:  " + strconv.FormatInt(orgId, 10) + "/" + name)
	scanner := cass.createQuery(projectsTablename, projectQuery, orgId, name)
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

func (cass *CassandraClient) FindAllProjects(org int64) []model.ProjectSettings {
	log.DefaultLogger.Info("findAllProjects:  " + strconv.FormatInt(org, 10))
	result := make([]model.ProjectSettings, 0)
	scanner := cass.createQuery(projectsTablename, projectsQuery, org)
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

func (cass *CassandraClient) GetSubsystem(org int64, projectName string, subsystem string) model.SubsystemSettings {
	log.DefaultLogger.Info("getSubsystem:  " + strconv.FormatInt(org, 10) + "/" + projectName + "/" + subsystem)
	scanner := cass.createQuery(subsystemsTablename, subsystemQuery, org, projectName, subsystem)
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

func (cass *CassandraClient) FindAllSubsystems(org int64, projectName string) []model.SubsystemSettings {
	log.DefaultLogger.Info("findAllSubsystems:  " + strconv.FormatInt(org, 10) + "/" + projectName)
	result := make([]model.SubsystemSettings, 0)
	scanner := cass.createQuery(subsystemsTablename, subsystemsQuery, org, projectName)
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

func (cass *CassandraClient) GetDatapoint(org int64, projectName string, subsystemName string, datapoint string) model.DatapointSettings {
	log.DefaultLogger.Info("getDatapoint:  " + strconv.FormatInt(org, 10) + "/" + projectName + "/" + datapoint)
	scanner := cass.createQuery(datapointsTablename, datapointQuery, org, projectName, subsystemName, datapoint)
	for scanner.Next() {
		return cass.deserializeDatapointRow(scanner)
	}
	return model.DatapointSettings{}
}

func (cass *CassandraClient) FindAllDatapoints(org int64, projectName string, subsystemName string) []model.DatapointSettings {
	log.DefaultLogger.Info("findAllDatapoints:  " + strconv.FormatInt(org, 10) + "/" + projectName + "/" + subsystemName)
	result := make([]model.DatapointSettings, 0)
	scanner := cass.createQuery(datapointsTablename, datapointsQuery, org, projectName, subsystemName)
	for scanner.Next() {
		datapoint := cass.deserializeDatapointRow(scanner)
		result = append(result, datapoint)
	}
	log.DefaultLogger.Info(fmt.Sprintf("Found: %d datapoints", len(result)))
	return result
}

func (cass *CassandraClient) SelectAllInJournal(org int64, journaltype string, journalname string) (model.Journal, error) {
	log.DefaultLogger.Info("SelectAllInJournal:  " + strconv.FormatInt(org, 10) + "/" + journaltype + "/" + journalname)
	result := model.Journal{
		Type: journaltype,
		Name: journalname,
	}
	scanner := cass.createQuery(journalTablename, journalSelectAllQuery, org, journaltype, journalname)
	for scanner.Next() {
		entry := model.JournalEntry{}
		err := scanner.Scan(&entry.Value, &entry.Added)
		if err != nil {
			log.DefaultLogger.Error(fmt.Sprintf("Unable to read Cassandra row(s) for %s (%s)", journalname, journaltype))
			return model.Journal{}, err
		}
		result.Entries = append(result.Entries, entry)
	}
	return result, nil
}

func (cass *CassandraClient) SelectRangeInJournal(org int64, journaltype string, journalname string, from time.Time, to time.Time) (model.Journal, error) {
	log.DefaultLogger.Info("SelectAllInJournal:  " + strconv.FormatInt(org, 10) + "/" + journaltype + "/" + journalname)
	result := model.Journal{
		Type: journaltype,
		Name: journalname,
	}
	scanner := cass.createQuery(journalTablename, journalSelectRangeQuery, org, journaltype, journalname, from, to)
	for scanner.Next() {
		entry := model.JournalEntry{}
		err := scanner.Scan(&entry.Value, &entry.Added)
		if err != nil {
			log.DefaultLogger.Error(fmt.Sprintf("Unable to read Cassandra row(s) for %s (%s)", journalname, journaltype))
			return model.Journal{}, err
		}
		result.Entries = append(result.Entries, entry)
	}
	return result, nil
}

func (cass *CassandraClient) Shutdown() {
	log.DefaultLogger.Info("Shutdown Cassandra client")
	cass.session.Close()
}

func (cass *CassandraClient) Err() error {
	return cass.err
}

func (cass *CassandraClient) createQuery(tableName string, query string, args ...interface{}) gocql.Scanner {
	t := fmt.Sprintf(query, cass.clusterConfig.Keyspace, tableName)
	q := cass.session.Query(t).Consistency(gocql.One).Idempotent(true).Bind(args...)
	//	log.DefaultLogger.Info("query:  " + q.String())
	return q.Iter().Scanner()
}

func (cass *CassandraClient) deserializeDatapointRow(scanner gocql.Scanner) model.DatapointSettings {
	var r model.DatapointSettings
	// project,subsystem,name,pollinterval,datasourcetype,timetolive,proc,ttnv3,web
	var ttnv3 model.Ttnv3Datasource
	var web model.WebDatasource
	var mqtt model.MqttDatasource
	err := scanner.Scan(&r.Project, &r.Subsystem, &r.Name, &r.Interval, &r.SourceType, &r.TimeToLive, &r.Proc, &ttnv3, &web, &mqtt)
	if err == nil {
		switch r.SourceType {
		case model.Web:
			r.Datasource = web
		case model.Ttnv3:
			r.Datasource = ttnv3
		case model.Mqtt:
			r.Datasource = mqtt
		}
	}
	if err != nil {
		log.DefaultLogger.Error(fmt.Sprintf("Internal Error? Failed to read record: %s, %+v", err.Error(), err))
	}
	return r
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
		//log.DefaultLogger.Info(fmt.Sprintf("Reduced to %d", len(downsized)))
		return downsized
	}
	//log.DefaultLogger.Info(fmt.Sprintf("Returning %d datapoints", len(result)))
	return result
}

const projectsTablename = "projects"

const projectQuery = "SELECT name,title,city,country,timezone,geolocation FROM %s.%s WHERE orgid = ? AND name = ? AND DELETED = '1970-01-01 0:00:00+0000' ALLOW FILTERING;"

const organizationsTablename = "organizations"

const organizationQuery = "SELECT name,email,stripecustomer,currentplan,address1,address2,zipcode,city,state,country FROM %s.%s WHERE orgid = ? AND DELETED = '1970-01-01 0:00:00+0000'  ALLOW FILTERING;"

const projectsQuery = "SELECT name,title,city,country,timezone,geolocation FROM %s.%s WHERE orgid = ? AND DELETED = '1970-01-01 0:00:00+0000' ALLOW FILTERING;"

const subsystemsTablename = "subsystems"

const subsystemQuery = "SELECT name,title,location FROM %s.%s WHERE orgid = ? AND project = ? AND name = ? AND DELETED = '1970-01-01 0:00:00+0000' ALLOW FILTERING;"

const subsystemsQuery = "SELECT name,title,location FROM %s.%s WHERE orgid = ? AND project = ? AND DELETED = '1970-01-01 0:00:00+0000' ALLOW FILTERING;"

const datapointsTablename = "datapoints"

const datapointQuery = "SELECT project,subsystem,name,pollinterval,datasourcetype,timetolive,proc,ttnv3,web,mqtt FROM %s.%s WHERE orgid = ? AND project = ? AND subsystem = ? AND name = ? AND DELETED = '1970-01-01 0:00:00+0000' ALLOW FILTERING;"

const datapointsQuery = "SELECT project,subsystem,name,pollinterval,datasourcetype,timetolive,proc,ttnv3,web,mqtt FROM %s.%s WHERE orgid = ? AND project = ? AND subsystem = ? AND DELETED = '1970-01-01 0:00:00+0000' ALLOW FILTERING;"

const planlimitsQuery = "SELECT orgid,created,maxdatapoints,maxstorage,minpollinterval FROM  %s.%s WHERE orgid = 5 AND deleted = '1970-01-01 00:00:00.000000+0000' ALLOW FILTERING;"

const planlimitsTablename = "planlimits"

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

const journalTablename = "journals"
const journalSelectAllQuery = "SELECT value,ts FROM %s.%s WHERE orgid = ? AND type = ? AND name = ?;"
const journalSelectRangeQuery = "SELECT value,ts FROM %s.%s WHERE orgid = ? AND type = ? AND name = ? AND ts >= ? AND  ts <= ? ;"
