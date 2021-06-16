package main

import (
	"context"
	JSON "encoding/json"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"net/http"
	"regexp"
	"strconv"
)

const regexName = `[a-zA-Z][a-zA-Z0-9-_]+`
const configurationTopic = "_configurations"

var (
	pathProject    = regexp.MustCompile(`projects/(?P<project>` + regexName + `)$`)
	pathSubsystems = regexp.MustCompile(`projects/(?P<project>` + regexName + `)/subsystems$`)
	pathSubsystem  = regexp.MustCompile(`projects/(?P<project>` + regexName + `)/subsystems/(?P<subsystem>` + regexName + `)$`)
	pathDatapoints = regexp.MustCompile(`projects/(?P<project>` + regexName + `)/subsystems/(?P<subsystem>` + regexName + `)/datapoints$`)
	pathDatapoint  = regexp.MustCompile(`projects/(?P<project>` + regexName + `)/subsystems/(?P<subsystem>` + regexName + `)/datapoints//(?P<datapoint>` + regexName + `)$`)
)

type ProjectHandler struct {
	cassandraClient *CassandraClient
	kafkaClient     *KafkaClient
}

func (p ProjectHandler) CallResource(ctx context.Context, request *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	orgId, err := getOrgId(request, sender)
	if err != nil {
		return err
	}
	log.DefaultLogger.Info(fmt.Sprintf("URL: %s; PATH: %s, Method: %s, OrgId: %d", request.URL, request.Path, request.Method, orgId))

	if http.MethodGet == request.Method {
		log.DefaultLogger.Info("11111")
		if request.URL == "projects" {
			log.DefaultLogger.Info("22222")
			return p.getProjects(orgId, sender)
		}
		if pathProject.Match([]byte(request.URL)) {
			log.DefaultLogger.Info("33333")
			projectName := pathProject.FindStringSubmatch(request.URL)[1]
			return p.getProject(orgId, projectName, sender)
		}
		if pathSubsystem.Match([]byte(request.URL)) {
			log.DefaultLogger.Info("44444")
			match := pathSubsystems.FindStringSubmatch(request.URL)
			projectName := match[1]
			subsystemName := match[2]
			return p.getSubsystem(orgId, projectName, subsystemName, sender)
		}
		if pathSubsystems.Match([]byte(request.URL)) {
			log.DefaultLogger.Info("55555")
			projectName := pathSubsystems.FindStringSubmatch(request.URL)[1]
			return p.getSubsystems(orgId, projectName, sender)
		}
		if pathDatapoint.Match([]byte(request.URL)) {
			log.DefaultLogger.Info("66666")
			match := pathSubsystems.FindStringSubmatch(request.URL)
			projectName := match[1]
			subsystemName := match[2]
			datapointName := match[3]
			log.DefaultLogger.Info("66666-->" + projectName + "/" + datapointName)
			return p.getDatapoint(orgId, projectName, subsystemName, datapointName, sender)
		}
		if pathDatapoints.Match([]byte(request.URL)) {
			log.DefaultLogger.Info("77777")
			match := pathSubsystems.FindStringSubmatch(request.URL)
			log.DefaultLogger.Info("77777-->" + match[0])
			projectName := match[1]
			log.DefaultLogger.Info("77777-->" + projectName)
			subsystemName := match[2]
			log.DefaultLogger.Info("77777-->" + projectName + "/" + subsystemName)
			return p.getDatapoints(orgId, projectName, subsystemName, sender)
		}
	}

	if http.MethodPut == request.Method {
		bodyRaw := request.Body
		if pathProject.Match([]byte(request.URL)) {
			p.updateProject(orgId, bodyRaw)
			sendAccepted(sender)
			return nil
		}
		if pathSubsystems.Match([]byte(request.URL)) {
			match := pathSubsystems.FindStringSubmatch(request.URL)
			projectName := match[1]
			var subsystem SubsystemSettings
			err := JSON.Unmarshal(bodyRaw, &subsystem)
			if err != nil {
				return err
			}
			subsystem.Project = projectName
			bodyRaw, err = JSON.Marshal(subsystem)
			if err != nil {
				return err
			}
			p.updateSubsystem(orgId, bodyRaw)
			sendAccepted(sender)
			return nil
		}
		if pathDatapoints.Match([]byte(request.URL)) {
			match := pathDatapoints.FindStringSubmatch(request.URL)
			projectName := match[1]
			subsystemName := match[2]
			var datapoint DatapointSettings
			err := JSON.Unmarshal(bodyRaw, &datapoint)
			if err != nil {
				return err
			}
			datapoint.Project = projectName
			datapoint.Subsystem = subsystemName
			bodyRaw, err = JSON.Marshal(datapoint)
			if err != nil {
				return err
			}
			p.updateDatapoint(orgId, bodyRaw)
			sendAccepted(sender)
			return nil
		}
	}
	return p.notFound(ctx, request, sender)
}

func sendAccepted(sender backend.CallResourceResponseSender) {
	empty := []byte{}
	err := sender.Send(
		&backend.CallResourceResponse{
			Status: http.StatusAccepted,
			Body:   empty,
		},
	)
	if err != nil {
		log.DefaultLogger.Error("Unable to send ACCEPTED reply" + err.Error())
	}
}

func (p ProjectHandler) updateProject(orgId int64, body []byte) {
	// TODO: validation in Grafana datasource before hitting the rest of the backend?
	p.kafkaClient.send(configurationTopic, "updateProject:1:"+strconv.FormatInt(orgId, 10), body)
}

func (p ProjectHandler) updateSubsystem(orgId int64, body []byte) {
	// TODO: validation in Grafana datasource before hitting the rest of the backend?
	p.kafkaClient.send(configurationTopic, "updateSubsystem:1:"+strconv.FormatInt(orgId, 10), body)
}

func (p ProjectHandler) updateDatapoint(orgId int64, body []byte) {
	// TODO: validation in Grafana datasource before hitting the rest of the backend?
	p.kafkaClient.send(configurationTopic, "updateDatapoint:1:"+strconv.FormatInt(orgId, 10), body)
}

func (p ProjectHandler) getProject(orgId int64, projectName string, sender backend.CallResourceResponseSender) error {
	project := p.cassandraClient.getProject(orgId, projectName)
	if project == nil {
		project = &ProjectSettings{}
		project.Name = "sbc1_malmo"
		project.Title = "Brf Benzelius"
		project.City = "Lund"
		project.Geolocation = "@55.884878,13.156352,13z"
	}
	bytes, err := JSON.Marshal(project)
	if err != nil {
		err = sender.Send(&backend.CallResourceResponse{
			Status:  http.StatusUnprocessableEntity,
			Headers: make(map[string][]string),
			Body:    []byte("Unable to marshal the entity. Probably wrong format: " + err.Error()),
		})
	} else {
		err = sender.Send(&backend.CallResourceResponse{
			Status:  http.StatusOK,
			Headers: make(map[string][]string),
			Body:    bytes,
		})
	}
	if err != nil {
		log.DefaultLogger.Error("Unable to write project to client.")
		return err
	} else {
		log.DefaultLogger.Info("Project '" + projectName + "' sent to client.")
		return nil
	}
}

func (p ProjectHandler) getProjects(orgId int64, sender backend.CallResourceResponseSender) error {

	projects := p.cassandraClient.findAllProjects(orgId)
	if len(projects) == 0 {
		var proj1 ProjectSettings
		proj1.Name = "sbc1_malmo"
		proj1.Title = "Brf Benzelius"
		proj1.City = "Lund"
		proj1.Country = "Sverige"
		proj1.Geolocation = "@55.884878,13.156352,13z"
		var proj2 ProjectSettings
		proj2.Name = "sbc2_malmo"
		proj2.Title = "Brf Lillbragden"
		proj2.City = "Malmö"
		proj2.Country = "Sverige"
		proj2.Geolocation = "@55.884878,13.156352,13z"
		var proj3 ProjectSettings
		proj3.Name = "sbc3_malmo"
		proj3.Title = "Brf Majoren"
		proj3.City = "Malmö"
		proj3.Country = "Sverige"
		proj3.Geolocation = "@55.884878,13.156352,13z"
		var proj4 ProjectSettings
		proj4.Name = "sbc4_malmo"
		proj4.Title = "Brf Schougen"
		proj4.City = "Malmö"
		proj4.Country = "Sverige"
		proj4.Geolocation = "@55.884878,13.156352,13z"
		var proj5 ProjectSettings
		proj5.Name = "sbc5_malmo"
		proj5.Title = "Brf Eslövsgården"
		proj5.City = "Malmö"
		proj5.Country = "Sverige"
		proj5.Geolocation = "@55.884878,13.156352,13z"
		projects = append(projects, proj1, proj2, proj3, proj4, proj5)
	}

	rawJson, err2 := JSON.Marshal(projects)
	if err2 != nil {
		log.DefaultLogger.Error("Unable to marshal json")
		return err2
	}
	err := sender.Send(&backend.CallResourceResponse{
		Status:  http.StatusOK,
		Headers: make(map[string][]string),
		Body:    rawJson,
	})

	if err != nil {
		log.DefaultLogger.Error("Unable to write projects to client.")
		return err
	}

	log.DefaultLogger.Info("Projects sent to client.\n" + string(rawJson[:]))
	return nil
}

func (p ProjectHandler) getSubsystem(orgId int64, project string, subsystemName string, sender backend.CallResourceResponseSender) error {
	subsystem := p.cassandraClient.getSubsystem(orgId, project, subsystemName)
	if subsystem == nil {
		subsystem = &SubsystemSettings{}
		subsystem.Project = project
		subsystem.Name = subsystemName
		subsystem.Title = "District Heating intake"
		subsystem.Locallocation = "BV-23"
	}
	bytes, err := JSON.Marshal(subsystem)
	if err != nil {
		err = sender.Send(&backend.CallResourceResponse{
			Status:  http.StatusUnprocessableEntity,
			Headers: make(map[string][]string),
			Body:    []byte("Unable to marshal the entity. Probably wrong format: " + err.Error()),
		})
	} else {
		err = sender.Send(&backend.CallResourceResponse{
			Status:  http.StatusOK,
			Headers: make(map[string][]string),
			Body:    bytes,
		})
	}
	if err != nil {
		log.DefaultLogger.Error("Unable to write subsystem to client.")
		return err
	} else {
		log.DefaultLogger.Info("Subsystem '" + subsystemName + "' sent to client.")
		return nil
	}
}

func (p ProjectHandler) getSubsystems(orgId int64, project string, sender backend.CallResourceResponseSender) error {

	subsystems := p.cassandraClient.findAllSubsystems(orgId, project)
	if len(subsystems) == 0 {
		var sub1 SubsystemSettings
		sub1.Project = project
		sub1.Name = "5601"
		sub1.Title = "District Heating intake"
		sub1.Locallocation = "BV-23"
		var sub2 SubsystemSettings
		sub2.Project = project
		sub2.Name = "5701"
		sub2.Title = "Ventilation TA1"
		sub2.Locallocation = "3-10"
		var sub3 SubsystemSettings
		sub3.Project = project
		sub3.Name = "5701"
		sub3.Title = "Ventilation TA2"
		sub3.Locallocation = "3-50"
		subsystems = append(subsystems, sub1, sub2, sub3)
	}
	rawJson, err2 := JSON.Marshal(subsystems)
	if err2 != nil {
		log.DefaultLogger.Error("Unable to marshal json")
		return err2
	}
	err := sender.Send(&backend.CallResourceResponse{
		Status:  http.StatusOK,
		Headers: make(map[string][]string),
		Body:    rawJson,
	})
	if err != nil {
		log.DefaultLogger.Error("Unable to write subsystems to client.")
		return err
	}
	log.DefaultLogger.Info("Subsystems sent to client:\n" + string(rawJson[:]))
	return nil
}

func (p ProjectHandler) getDatapoint(orgId int64, project string, subsystem string, datapointName string, sender backend.CallResourceResponseSender) error {
	datapoint := p.cassandraClient.getDatapoint(orgId, project, subsystem, datapointName)
	if datapoint == nil {
		datapoint = &DatapointSettings{}
		datapoint.Project = project
		datapoint.Subsystem = subsystem
		datapoint.Name = datapointName
		datapoint.Interval = Thirty_minutes
		datapoint.URL = "https://api.darksky.net/forecast/615bfb2b3db89dea530f3fb6e0c9c38c/55.8794518,13.1609417"
		datapoint.AuthenticationType = none
		datapoint.Format = json
		datapoint.ValueExpression = "$.currently.temperature"
		datapoint.Unit = "ºC"
		datapoint.Scaling = fToC
		datapoint.TimeToLive = d
		datapoint.TimestampType = epochSeconds
		datapoint.TimestampExpression = "$.currently.time"
	}
	bytes, err := JSON.Marshal(datapoint)
	if err != nil {
		err = sender.Send(&backend.CallResourceResponse{
			Status:  http.StatusUnprocessableEntity,
			Headers: make(map[string][]string),
			Body:    []byte("Unable to marshal the entity. Probably wrong format: " + err.Error()),
		})
	} else {
		err = sender.Send(&backend.CallResourceResponse{
			Status:  http.StatusOK,
			Headers: make(map[string][]string),
			Body:    bytes,
		})
	}
	if err != nil {
		log.DefaultLogger.Error("Unable to write datapoint to client.")
		return err
	} else {
		log.DefaultLogger.Info("Project '" + datapointName + "' sent to client.")
		return nil
	}
}

func (p ProjectHandler) getDatapoints(orgId int64, project string, subsystem string, sender backend.CallResourceResponseSender) error {
	datapoints := p.cassandraClient.findAllDatapoints(orgId, project, subsystem)
	if len(datapoints) == 0 {
		var dp1 DatapointSettings
		dp1.Project = project
		dp1.Subsystem = subsystem
		dp1.Name = "heating"
		dp1.Interval = Thirty_minutes
		dp1.URL = "https://api.darksky.net/forecast/615bfb2b3db89dea530f3fb6e0c9c38c/55.8794518,13.1609417"
		dp1.AuthenticationType = none
		dp1.Format = json
		dp1.ValueExpression = "$.currently.temperature"
		dp1.Unit = "ºC"
		dp1.Scaling = fToC
		dp1.TimeToLive = d
		dp1.TimestampType = epochSeconds
		dp1.TimestampExpression = "$.currently.time"
		var dp2 DatapointSettings
		dp2.Project = project
		dp2.Subsystem = subsystem
		dp2.Name = "heating"
		dp2.Interval = Thirty_minutes
		dp2.URL = "https://api.darksky.net/forecast/615bfb2b3db89dea530f3fb6e0c9c38c/55.8794518,13.1609417"
		dp2.AuthenticationType = none
		dp2.Format = json
		dp2.ValueExpression = "$.currently.humidity"
		dp2.Unit = "%"
		dp2.Scaling = lin
		dp2.K = 1.0
		dp2.M = 0.0
		dp2.TimeToLive = d
		dp2.TimestampType = epochSeconds
		dp2.TimestampExpression = "$.currently.time"
		datapoints = append(datapoints, dp1, dp2)
	}
	rawJson, err2 := JSON.Marshal(datapoints)
	if err2 != nil {
		log.DefaultLogger.Error("Unable to marshal json")
		return err2
	}
	err := sender.Send(&backend.CallResourceResponse{
		Status:  http.StatusOK,
		Headers: make(map[string][]string),
		Body:    rawJson,
	})
	if err != nil {
		log.DefaultLogger.Error("Unable to write subsystems to client.")
		return err
	}
	log.DefaultLogger.Info("Datapoints sent to client:\n" + string(rawJson[:]))
	return nil
}

func getOrgId(request *backend.CallResourceRequest, sender backend.CallResourceResponseSender) (int64, error) {
	orgIdHeader := request.Headers["X-Grafana-Org-Id"][0]
	orgId, err := strconv.ParseInt(orgIdHeader, 10, 64)
	if err != nil {
		sender.Send(&backend.CallResourceResponse{
			Status:  http.StatusNotAcceptable,
			Headers: make(map[string][]string),
			Body:    []byte("Header X-Grafana-Org-Id is missing."),
		})
		return 0, err
	}
	return orgId, nil
}

func (p ProjectHandler) notFound(ctx context.Context, request *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	log.DefaultLogger.Error("Resource " + request.URL + " not found.")

	return sender.Send(&backend.CallResourceResponse{
		Status:  404,
		Headers: make(map[string][]string),
		Body:    []byte("Not found"),
	})
}
