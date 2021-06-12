package main

import (
	"context"
	JSON "encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

const regexName = `[a-zA-Z][a-zA-Z0-9-_]+`

var (
	pathProject    = regexp.MustCompile(`projects/(?P<project>` + regexName + `)$`)
	pathSubsystems = regexp.MustCompile(`projects/(?P<project>` + regexName + `)/subsystems$`)

	pathDatapoints = regexp.MustCompile(`projects/(?P<project>` + regexName + `)/subsystems/(?P<subsystem>` + regexName + `)/datapoints$`)
)

type ProjectHandler struct {
	cassandraClient *CassandraClient
}

func (p ProjectHandler) CallResource(ctx context.Context, request *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	log.DefaultLogger.Info("Resource Request: " + fmt.Sprintf("%s %s", request.Method, request.Path))

	log.DefaultLogger.Info(fmt.Sprintf("URL: %s; PATH: %s", request.URL, request.Path))

	if request.URL == "projects" && http.MethodPost == request.Method {
		return p.addProject(ctx, request, sender)
	}

	if request.URL == "projects" && http.MethodGet == request.Method {
		return p.getProjects(ctx, request, sender)
	}

	if request.URL == "projects" && http.MethodPost == request.Method {
		return p.updateProject(ctx, request, sender)
	}

	if pathDatapoints.Match([]byte(request.URL)) && http.MethodGet == request.Method {
		return p.getDatapoints(ctx, request, sender)
	}

	if pathSubsystems.Match([]byte(request.URL)) && http.MethodGet == request.Method {
		return p.getSubystems(ctx, request, sender)
	}

	if pathProject.Match([]byte(request.URL)) && http.MethodGet == request.Method {
		return p.getProject(ctx, request, sender)
	}

	return p.notFound(ctx, request, sender)
}

func (p ProjectHandler) addProject(ctx context.Context, request *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	bodyRaw := request.Body
	log.DefaultLogger.Info("[addProject]")

	project := ProjectSettings{}
	if err := JSON.Unmarshal(bodyRaw, &project); err != nil {
		log.DefaultLogger.Error(fmt.Sprintf("[addProject] unmarshaling. Raw project: %s", string(bodyRaw)))
		return err
	}

	return sender.Send(
		&backend.CallResourceResponse{
			Status: http.StatusCreated,
			Body: []byte(`{
				"name":` + project.Name + `
			}`),
		},
	)
}

func (p ProjectHandler) updateProject(ctx context.Context, request *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	return nil
}

func (p ProjectHandler) getProject(ctx context.Context, request *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	match := pathProject.FindStringSubmatch(request.URL)
	log.DefaultLogger.Info(fmt.Sprintf("[getProject] %v ", match))

	responseRaw := `{
			  "name": "sbc1_malmo",
			  "title": "Brf Benzelius",
			  "city": "Lund",
			  "geolocation": "@55.884878,13.156352,13z",
			  "subsystems": []
		}`

	err := sender.Send(&backend.CallResourceResponse{
		Status:  http.StatusOK,
		Headers: make(map[string][]string),
		Body:    []byte(responseRaw),
	})

	if err != nil {
		log.DefaultLogger.Error("Unable to write projects to client.")
		return err
	}

	log.DefaultLogger.Info("Projects sent to client.")
	return nil
}
func (p ProjectHandler) getProjects(ctx context.Context, request *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {

	orgIdHeader := request.Headers["X-Grafana-Org-Id"][0]
	orgId, err := strconv.ParseInt(orgIdHeader, 10, 64)
	if err != nil {
		sender.Send(&backend.CallResourceResponse{
			Status:  http.StatusNotAcceptable,
			Headers: make(map[string][]string),
			Body:    []byte("Header X-Grafana-Org-Id is missing."),
		})
	}
	var projects []ProjectSettings
	projects = p.cassandraClient.findAllProjects(orgId)
	if projects == nil {
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
	err = sender.Send(&backend.CallResourceResponse{
		Status:  http.StatusOK,
		Headers: make(map[string][]string),
		Body:    []byte(rawJson),
	})

	if err != nil {
		log.DefaultLogger.Error("Unable to write projects to client.")
		return err
	}

	log.DefaultLogger.Info("Projects sent to client.\n" + string(rawJson[:]))
	return nil
}

func (p ProjectHandler) getSubystems(ctx context.Context, request *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	match := pathSubsystems.FindStringSubmatch(request.URL)
	log.DefaultLogger.Info(fmt.Sprintf("[getSubystems] %v ", match))
	responseRaw := `[
		{
		  "name": "sbc1_malmo project 1",
		  "title": "Project 1",
		  "city": "Lund"
		},
		{
		  "name": "sbc1_malmo project 2",
		  "title": "second Project",
		  "city": "Malmö"
		}
	  ]`

	err := sender.Send(&backend.CallResourceResponse{
		Status:  http.StatusOK,
		Headers: make(map[string][]string),
		Body:    []byte(responseRaw),
	})

	if err != nil {
		log.DefaultLogger.Error("Unable to write subsystems to client.")
		return err
	}

	log.DefaultLogger.Info("Subsystems sent to client.")
	return nil
}

func (p ProjectHandler) getDatapoints(ctx context.Context, request *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	match := pathDatapoints.FindStringSubmatch(request.URL)
	log.DefaultLogger.Info(fmt.Sprintf("[getDatapoints] %v ", match))
	responseRaw := `[
		{
			"name": "point1"
		},{
			"name": "point2"
		},{
			"name": "point3"
		},{
			"name": "point4"
		  }
	  ]`

	err := sender.Send(&backend.CallResourceResponse{
		Status:  http.StatusOK,
		Headers: make(map[string][]string),
		Body:    []byte(responseRaw),
	})

	if err != nil {
		log.DefaultLogger.Error("Unable to write datapoints to client.")
		return err
	}

	log.DefaultLogger.Info("datapoints sent to client.")
	return nil
}

func (p ProjectHandler) notFound(ctx context.Context, request *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	log.DefaultLogger.Error("Resource " + request.URL + " not found.")

	return sender.Send(&backend.CallResourceResponse{
		Status:  404,
		Headers: make(map[string][]string),
		Body:    []byte("Not found"),
	})
}
