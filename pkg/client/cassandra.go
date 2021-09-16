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
	log.DefaultLogger.Info("queryTimeseries:  " + strconv.FormatInt(org, 10) + "/" + sensor.Project + "/" + sensor.Subsystem + "/" + sensor.Datapoint + "   " + from.Format(time.RFC3339) + "->" + to.Format(time.RFC3339))
	var result []model.TsPair
	startYearMonth := from.Year()*12 + int(from.Month()) - 1
	endYearMonth := to.Year()*12 + int(to.Month()) - 1
	log.DefaultLogger.Info(fmt.Sprintf("yearMonths:  start=%d, end=%d", startYearMonth, endYearMonth))

	for yearmonth := startYearMonth; yearmonth <= endYearMonth; yearmonth++ {
		scanner := cass.createQuery(timeseriesTablename, tsQuery, org, sensor.Project, sensor.Subsystem, yearmonth, sensor.Datapoint, from, to)
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

func (cass *CassandraClient) GetOrganization(orgId int64) model.OrganizationSettings {
	log.DefaultLogger.Info("getOrganization:  " + strconv.FormatInt(orgId, 10))
	//SELECT name,email,stripecustomer,currentplan,address1,address2,zipcode,city,state,country FROM %s.%s WHERE orgid = ? AND DELETED = '1970-01-01 0:00:00+0:00';"
	scanner := cass.createQuery(organizationsTablename, organizationQuery, orgId)
	for scanner.Next() {
		var org model.OrganizationSettings
		err := scanner.Scan(&org.Name, &org.Email, &org.StripeCustomer, &org.CurrentPlan,
			&org.Address.Address1, &org.Address.Address2, &org.Address.ZipCode, &org.Address.City, &org.Address.State, &org.Address.Country)
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
		return cass.deserializeRow(scanner)
	}
	return model.DatapointSettings{}
}

func (cass *CassandraClient) FindAllDatapoints(org int64, projectName string, subsystemName string) []model.DatapointSettings {
	log.DefaultLogger.Info("findAllDatapoints:  " + strconv.FormatInt(org, 10) + "/" + projectName + "/" + subsystemName)
	result := make([]model.DatapointSettings, 0)
	scanner := cass.createQuery(datapointsTablename, datapointsQuery, org, projectName, subsystemName)
	for scanner.Next() {
		datapoint := cass.deserializeRow(scanner)
		result = append(result, datapoint)
	}
	log.DefaultLogger.Info(fmt.Sprintf("Found: %d datapoints", len(result)))
	return result
}

func (cass *CassandraClient) FindAllPlans(orgid int64) []model.PlanSettings {
	log.DefaultLogger.Info("findAllPlans:  " + strconv.FormatInt(orgid, 10))
	result := make([]model.PlanSettings, 0)
	scanner := cass.createQuery(plansTablename, plansQuery)
	for scanner.Next() {
		var plan model.PlanSettings
		err := scanner.Scan(&plan.Name, &plan.Title, &plan.Description, &plan.Price, &plan.StripePrice, &plan.Currency,
			&plan.Active, &plan.Private,
			&plan.Start, &plan.End,
			&plan.Limits.MaxProjects, &plan.Limits.MaxCollaborators, &plan.Limits.MaxDatapoints,
			&plan.Limits.MaxStorage, &plan.Limits.MinPollInterval)
		if err != nil {
			log.DefaultLogger.Error(fmt.Sprintf("Internal Error? Failed to read record: %s", err.Error()))
		} else {
			result = append(result, plan)
		}
	}
	log.DefaultLogger.Info(fmt.Sprintf("Found: %d plans", len(result)))
	return result
}
func (cass *CassandraClient) FindAllPayments(orgid int64) []model.Payment {
	log.DefaultLogger.Info("findAllPayments:  " + strconv.FormatInt(orgid, 10))
	result := make([]model.Payment, 0)
	scanner := cass.createQuery(paymentsTablename, paymentsQuery)
	for scanner.Next() {
		var p model.Payment
		err := scanner.Scan(&p.InvoiceDate, &p.PaymentDate, &p.Amount, &p.Currency)
		if err != nil {
			log.DefaultLogger.Error("Internal Error? Failed to read record", err)
		} else {
			result = append(result, p)
		}
	}
	log.DefaultLogger.Info(fmt.Sprintf("Found: %d payments", len(result)))
	return result
}

func (cass *CassandraClient) FindAllInvoices(orgid int64) []model.Invoice {
	log.DefaultLogger.Info("findAllInvoices:  " + strconv.FormatInt(orgid, 10))
	result := make([]model.Invoice, 0)
	scanner := cass.createQuery(invoicesTablename, invoicesQuery)
	for scanner.Next() {
		var inv model.Invoice
		err := scanner.Scan(&inv.InvoiceDate, &inv.PlanTitle, &inv.PlanDescription, &inv.Stats, &inv.Amount, &inv.Currency)
		if err != nil {
			log.DefaultLogger.Error("Internal Error? Failed to read record", err)
		} else {
			result = append(result, inv)
		}
	}
	log.DefaultLogger.Info(fmt.Sprintf("Found: %d invoices", len(result)))
	return result
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
	log.DefaultLogger.Info("query:  " + q.String())
	return q.Iter().Scanner()
}

func (cass *CassandraClient) deserializeRow(scanner gocql.Scanner) model.DatapointSettings {
	var r model.DatapointSettings
	// project,subsystem,name,pollinterval,datasourcetype,timetolive,proc,ttnv3,web
	var ttnv3 model.Ttnv3Datasource
	var web model.WebDatasource
	err := scanner.Scan(&r.Project, &r.Subsystem, &r.Name, &r.Interval, &r.SourceType, &r.TimeToLive, &r.Proc, &ttnv3, &web)
	if err == nil {
		switch r.SourceType {
		case model.Web:
			r.Datasource = web
		case model.Ttnv3:
			r.Datasource = ttnv3
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
		log.DefaultLogger.Info(fmt.Sprintf("Reduced to %d", len(downsized)))
		return downsized
	}
	log.DefaultLogger.Info(fmt.Sprintf("Returning %d datapoints", len(result)))
	return result
}

const projectsTablename = "projects"

const projectQuery = "SELECT name,title,city,country,timezone,geolocation FROM %s.%s WHERE orgid = ? AND name = ?;"

const organizationsTablename = "organizations"

const organizationQuery = "SELECT name,email,stripecustomer,currentplan,address1,address2,zipcode,city,state,country FROM %s.%s WHERE orgid = ? AND DELETED = '1970-01-01 0:00:00+0:00';"

const projectsQuery = "SELECT name,title,city,country,timezone,geolocation FROM %s.%s WHERE orgid = ?;"

const subsystemsTablename = "subsystems"

const subsystemQuery = "SELECT name,title,location FROM %s.%s WHERE orgid = ? AND project = ? AND name = ?;"

const subsystemsQuery = "SELECT name,title,location FROM %s.%s WHERE orgid = ? AND project = ?;"

const datapointsTablename = "datapoints"

const datapointQuery = "SELECT project,subsystem,name,pollinterval,datasourcetype,timetolive,proc,ttnv3,web FROM %s.%s WHERE orgid = ? AND project = ? AND subsystem = ? AND name = ?;"

const datapointsQuery = "SELECT project,subsystem,name,pollinterval,datasourcetype,timetolive,proc,ttnv3,web FROM %s.%s WHERE orgid = ? AND project = ? AND subsystem = ?;"

const plansTablename = "plans"
const plansQuery = "SELECT name,title,description,price,stripeprice,currency,active,private,start,end,maxprojects,maxcollaborators,maxdatapoints,maxstorage,minpollinterval FROM %s.%s;"

const invoicesTablename = "invoices"
const invoicesQuery = "SELECT invoicedate,plantitle,plandescription,stats,amount,currency FROM %s.%s WHERE orgid = ?;"

const paymentsTablename = "payments"
const paymentsQuery = "SELECT invoicedate,paymentdate,amount,currency FROM %s.%s WHERE orgid = ?;"

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
