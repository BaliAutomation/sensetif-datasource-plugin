package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-starter-datasource-backend/pkg/model"
)

type ProjectHandler struct {
	cassandraClient *CassandraClient
}

func (p ProjectHandler) CallResource(ctx context.Context, request *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	log.DefaultLogger.Info("Resource Request: " + fmt.Sprintf("%s %s", request.Method, request.Path))

	if request.URL == "projects" && http.MethodPost == request.Method {
		return p.addProject(ctx, request, sender)
	}

	if request.URL == "projects" && http.MethodGet == request.Method {
		return p.getProjects(ctx, request, sender)
	}

	if request.URL == "projects" && http.MethodPost == request.Method {
		return p.updateProject(ctx, request, sender)
	}

	return p.notFound(ctx, request, sender)
}

func (p ProjectHandler) addProject(ctx context.Context, request *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	bodyRaw := request.Body
	log.DefaultLogger.Info("[addProject]")

	project := model.Project{}
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

func (p ProjectHandler) notFound(ctx context.Context, request *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	log.DefaultLogger.Error("Resource " + request.URL + " not found.")

	return sender.Send(&backend.CallResourceResponse{
		Status:  404,
		Headers: make(map[string][]string),
		Body:    []byte("Not found"),
	})
}
