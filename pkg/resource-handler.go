package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/BaliAutomation/sensetif-datasource/pkg/client"
	"github.com/BaliAutomation/sensetif-datasource/pkg/handler"
	"github.com/BaliAutomation/sensetif-datasource/pkg/model"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

type ResourceHandler struct {
	cassandra client.Cassandra
	kafka     client.Kafka
}

func (p ResourceHandler) CallResource(ctx context.Context, request *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	orgId, err := getOrgId(request, sender)
	if err != nil {
		if sendErr := sender.Send(&backend.CallResourceResponse{
			Status:  http.StatusNotAcceptable,
			Headers: make(map[string][]string),
			Body:    []byte("Header X-Grafana-Org-Id is missing."),
		}); sendErr != nil {
			log.DefaultLogger.Error("could not write response to the client")
			return sendErr
		}
	}

	log.DefaultLogger.Info(fmt.Sprintf("URL: %s; PATH: %s, Method: %s, OrgId: %d", request.URL, request.Path, request.Method, orgId))

	var cmd model.Command
	if err := json.Unmarshal(request.Body, &cmd); err != nil {
		return p.badRequest("invalid format of the command", sender)
	}

	log.DefaultLogger.Info(fmt.Sprintf("request: [%s] [%s]", cmd.Action, cmd.Resource))
	cmd.OrgID = orgId

	if request.URL != "exec" {
		return p.notFound("", sender)
	}

	response, err := p.handle(&cmd, request)
	if err == nil {
		if sendErr := sender.Send(response); sendErr != nil {
			log.DefaultLogger.Error("could not write response to the client")
			return sendErr
		}

		return nil
	}

	if errors.Is(err, model.ErrNotFound) {
		return p.notFound(err.Error(), sender)
	}
	if errors.Is(err, model.ErrBadRequest) {
		return p.badRequest(err.Error(), sender)
	}
	if errors.Is(err, model.ErrUnprocessableEntity) {
		return p.unprocessable(err.Error(), sender)
	}

	return p.serverError(err.Error(), sender)
}

func (p ResourceHandler) handle(cmd *model.Command, request *backend.CallResourceRequest) (*backend.CallResourceResponse, error) {
	if cmd.Resource == "project" && cmd.Action == "list" {
		return handler.ListProjects(cmd, p.cassandra)
	}

	if cmd.Resource == "project" && cmd.Action == "get" {
		return handler.GetProject(cmd, p.cassandra)
	}

	if cmd.Resource == "project" && cmd.Action == "update" {
		return handler.UpdateProject(cmd, p.cassandra, p.kafka)
	}

	if cmd.Resource == "subsystem" && cmd.Action == "list" {
		return handler.ListSubsystems(cmd, p.cassandra)
	}

	if cmd.Resource == "subsystem" && cmd.Action == "update" {
		return handler.UpdateSubsystem(cmd, p.kafka)
	}

	if cmd.Resource == "datapoint" && cmd.Action == "list" {
		return handler.ListDatapoints(cmd, p.cassandra)
	}

	return nil, model.ErrNotFound
}

func getOrgId(request *backend.CallResourceRequest, sender backend.CallResourceResponseSender) (int64, error) {
	orgIdHeader := request.Headers["X-Grafana-Org-Id"][0]
	return strconv.ParseInt(orgIdHeader, 10, 64)
}

func (p ResourceHandler) notFound(message string, sender backend.CallResourceResponseSender) error {
	return sender.Send(&backend.CallResourceResponse{
		Status: http.StatusNotFound,
		Body:   p.createMessageJSON(message),
	})
}

func (p ResourceHandler) badRequest(message string, sender backend.CallResourceResponseSender) error {
	return sender.Send(&backend.CallResourceResponse{
		Status: http.StatusBadRequest,
		Body:   p.createMessageJSON(message),
	})
}

func (p ResourceHandler) unprocessable(message string, sender backend.CallResourceResponseSender) error {
	return sender.Send(&backend.CallResourceResponse{
		Status: http.StatusUnprocessableEntity,
		Body:   []byte("Unable to marshal the entity. Probably wrong format: " + message),
	})
}

func (p ResourceHandler) serverError(message string, sender backend.CallResourceResponseSender) error {
	return sender.Send(&backend.CallResourceResponse{
		Status: http.StatusInternalServerError,
		Body:   []byte(message),
	})
}

func (p ResourceHandler) createMessageJSON(message string) []byte {
	response, _ := json.Marshal(struct {
		Message string `json:"message"`
	}{
		Message: message,
	})

	return response
}
