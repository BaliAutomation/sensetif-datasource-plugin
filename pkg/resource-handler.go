package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	. "regexp"
	"strconv"
	"strings"

	"github.com/BaliAutomation/sensetif-datasource/pkg/client"
	"github.com/BaliAutomation/sensetif-datasource/pkg/handler"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

type ResourceHandler struct {
	cassandra client.Cassandra
	kafka     client.Kafka
}

type Link struct {
	Pattern *Regexp
	Method  string
	Fn      func(orgId int64, params []string, body []byte, kafka client.Kafka, cassandra client.Cassandra) (*backend.CallResourceResponse, error)
}

const regexName = `[a-zA-Z][a-zA-Z0-9-_]*`

var links = []Link{
	// Projects API
	{Method: "GET", Fn: handler.ListProjects, Pattern: MustCompile(`^_$`)},
	{Method: "GET", Fn: handler.GetProject, Pattern: MustCompile(`^(` + regexName + `)$`)},
	{Method: "PUT", Fn: handler.UpdateProject, Pattern: MustCompile(`^(` + regexName + `)$`)},
	{Method: "DELETE", Fn: handler.DeleteProject, Pattern: MustCompile(`^(` + regexName + `)$`)},
	{Method: "POST", Fn: handler.RenameProject, Pattern: MustCompile(`^(` + regexName + `)$`)},

	// Subsystems API
	{Method: "GET", Fn: handler.ListSubsystems, Pattern: MustCompile(`^(` + regexName + `)/_$`)},
	{Method: "GET", Fn: handler.GetSubsystem, Pattern: MustCompile(`^(` + regexName + `)/(` + regexName + `)$`)},
	{Method: "PUT", Fn: handler.UpdateSubsystem, Pattern: MustCompile(`^(` + regexName + `)/(` + regexName + `)$`)},
	{Method: "DELETE", Fn: handler.DeleteSubsystem, Pattern: MustCompile(`^(` + regexName + `)/(` + regexName + `)$`)},
	{Method: "POST", Fn: handler.RenameSubsystem, Pattern: MustCompile(`^(` + regexName + `)/(` + regexName + `)$`)},

	// Datapoint API
	{Method: "GET", Fn: handler.ListDatapoints, Pattern: MustCompile(`^(` + regexName + `)/(` + regexName + `)/_$`)},
	{Method: "GET", Fn: handler.GetDatapoint, Pattern: MustCompile(`^(` + regexName + `)/(` + regexName + `)/(` + regexName + `)$`)},
	{Method: "PUT", Fn: handler.UpdateDatapoint, Pattern: MustCompile(`^(` + regexName + `)/(` + regexName + `)/(` + regexName + `)$`)},
	{Method: "DELETE", Fn: handler.DeleteDatapoint, Pattern: MustCompile(`^(` + regexName + `)/(` + regexName + `)/(` + regexName + `)$`)},
	{Method: "POST", Fn: handler.RenameDatapoint, Pattern: MustCompile(`^(` + regexName + `)/(` + regexName + `)/(` + regexName + `)$`)},
}

//goland:noinspection GoUnusedParameter
func (p ResourceHandler) CallResource(ctx context.Context, request *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	log.DefaultLogger.Info(fmt.Sprintf("URL: %s; PATH: %s, Method: %s", request.URL, request.Path, request.Method))
	err2, found := handleFileRequests(request, sender)
	if found {
		return err2
	}
	orgId, err := getOrgId(request)
	if err != nil {
		if sendErr := sender.Send(&backend.CallResourceResponse{
			Status:  http.StatusNotAcceptable,
			Headers: make(map[string][]string),
			Body:    []byte("Header X-Grafana-Org-Id is missing."),
		}); sendErr != nil {
			log.DefaultLogger.Error("could not write response to the client." + err.Error())
			return sendErr
		}
	}
	log.DefaultLogger.Info(fmt.Sprintf("URL: %s; PATH: %s, Method: %s, OrgId: %d", request.URL, request.Path, request.Method, orgId))

	for _, link := range links {
		if link.Method == request.Method {
			parameters := link.Pattern.FindStringSubmatch(request.URL)
			if len(parameters) >= 1 {
				log.DefaultLogger.Info(fmt.Sprintf("Parameters: %q --> %s", strings.Join(parameters, ","), string(request.Body)))
				result, err := link.Fn(orgId, parameters, request.Body, p.kafka, p.cassandra)
				if err == nil {
					if sendErr := sender.Send(result); sendErr != nil {
						log.DefaultLogger.Error("could not write response to the client. " + sendErr.Error())
						return sendErr
					}
					return nil
				}
			}
		}
	}
	return notFound("", sender)
}

func handleFileRequests(request *backend.CallResourceRequest, sender backend.CallResourceResponseSender) (error, bool) {
	if strings.Index(request.Path, "__/") == 0 {
		filename := strings.TrimLeft(request.Path, "__/")
		content, err := HandleFile(filename)
		if err != nil {
			log.DefaultLogger.Error("Could not read file: " + filename + " : " + err.Error())
			return err, true
		}
		err = sender.Send(content)
		if err != nil {
			log.DefaultLogger.Error("could not write content to the client. " + err.Error())
			return err, true
		}
		return nil, true
	}
	return nil, false
}

func getOrgId(request *backend.CallResourceRequest) (int64, error) {
	orgIdHeader := request.Headers["X-Grafana-Org-Id"][0]
	return strconv.ParseInt(orgIdHeader, 10, 64)
}

func notFound(message string, sender backend.CallResourceResponseSender) error {
	return sender.Send(&backend.CallResourceResponse{
		Status: http.StatusNotFound,
		Body:   createMessageJSON(message),
	})
}

func createMessageJSON(message string) []byte {
	response, _ := json.Marshal(struct {
		Message string `json:"message"`
	}{
		Message: message,
	})

	return response
}
