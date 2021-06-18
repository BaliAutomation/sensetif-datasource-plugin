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
	"strings"
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
		if request.URL == "projects" {
			return p.getProjects(orgId, sender)
		}
		if pathProject.Match([]byte(request.URL)) {
			match := pathProject.FindStringSubmatch(request.URL)
			if match == nil {
				sender.Send(&backend.CallResourceResponse{
					Status: http.StatusBadRequest,
				})
			} else {
				projectName := match[1]
				return p.getProject(orgId, projectName, sender)
			}
		}
		if pathSubsystem.Match([]byte(request.URL)) {
			match := pathSubsystems.FindStringSubmatch(request.URL)
			if match == nil {
				sender.Send(&backend.CallResourceResponse{
					Status: http.StatusBadRequest,
				})
			} else {
				projectName := match[1]
				subsystemName := match[2]
				return p.getSubsystem(orgId, projectName, subsystemName, sender)
			}
		}
		if pathSubsystems.Match([]byte(request.URL)) {
			match := pathSubsystems.FindStringSubmatch(request.URL)
			if match == nil {
				sender.Send(&backend.CallResourceResponse{
					Status: http.StatusBadRequest,
				})
			} else {
				projectName := match[1]
				return p.getSubsystems(orgId, projectName, sender)
			}
		}
		if pathDatapoint.Match([]byte(request.URL)) {
			match := pathSubsystems.FindStringSubmatch(request.URL)
			if match == nil {
				sender.Send(&backend.CallResourceResponse{
					Status: http.StatusBadRequest,
				})
			} else {
				projectName := match[1]
				subsystemName := match[2]
				datapointName := match[3]
				log.DefaultLogger.Info("66666-->" + projectName + "/" + datapointName)
				return p.getDatapoint(orgId, projectName, subsystemName, datapointName, sender)
			}
		}
		if pathDatapoints.Match([]byte(request.URL)) {
			request.URL = strings.Trim(request.URL, " \t\n\r")
			match := pathDatapoints.FindStringSubmatch(request.URL)
			if match == nil {
				sender.Send(&backend.CallResourceResponse{
					Status: http.StatusBadRequest,
				})
			} else {
				projectName := match[1]
				subsystemName := match[2]
				return p.getDatapoints(orgId, projectName, subsystemName, sender)
			}
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
			if match == nil {
				sender.Send(&backend.CallResourceResponse{
					Status: http.StatusBadRequest,
				})
			} else {
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
		}
		if pathDatapoints.Match([]byte(request.URL)) {
			match := pathDatapoints.FindStringSubmatch(request.URL)
			if match == nil {
				sender.Send(&backend.CallResourceResponse{
					Status: http.StatusBadRequest,
				})
			} else {
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
	rawJson, err2 := JSON.Marshal(projects)
	if err2 != nil {
		log.DefaultLogger.Error("Unable to marshal json")
		return err2
	}
	if rawJson == nil || len(rawJson) == 0 {
		rawJson = []byte{'[', ']'}
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
	rawJson, err2 := JSON.Marshal(subsystems)
	if err2 != nil {
		log.DefaultLogger.Error("Unable to marshal json")
		return err2
	}
	log.DefaultLogger.Info("rawJson:\n" + string(rawJson[:]))
	if rawJson == nil || len(rawJson) == 0 {
		rawJson = []byte{'[', ']'}
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
	rawJson, err2 := JSON.Marshal(datapoints)
	if err2 != nil {
		log.DefaultLogger.Error("Unable to marshal json")
		return err2
	}
	if rawJson == nil || len(rawJson) == 0 {
		rawJson = []byte{'[', ']'}
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
