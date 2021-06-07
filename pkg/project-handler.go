package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

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
	if err := json.Unmarshal(bodyRaw, &project); err != nil {
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
	responseRaw := `[
			{
			  "name": "sbc1_malmo",
			  "title": "Brf Benzelius",
			  "city": "Lund",
			  "geolocation": "@55.884878,13.156352,13z",
			  "subsystems": []
			},
			{
			  "name": "sbc2_malmo",
			  "title": "Brf Lillbragden",
			  "city": "Malmö",
			  "geolocation": "@55.884878,13.156352,13z",
			  "subsystems": []
			},
			{
			  "name": "sbc3_malmo",
			  "title": "Brf Majoren",
			  "city": "Malmö",
			  "geolocation": "@55.884878,13.156352,13z",
			  "subsystems": []
			},
			{
			  "name": "sbc4_malmo",
			  "title": "Brf Schougen",
			  "city": "Malmö",
			  "geolocation": "@55.884878,13.156352,13z",
			  "subsystems": []
			}
		  ]`

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
