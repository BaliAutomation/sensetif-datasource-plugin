package main

import (
	"context"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

type ProjectHandler struct {
	cassandraClient *CassandraClient
}

//func (p ProjectHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
func (p ProjectHandler) CallResource(ctx context.Context, request *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	log.DefaultLogger.Info("Resource Request: " + fmt.Sprintf("%s %s", request.Method, request.Path))
	switch request.URL {
	case "projects":
		json := `[
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
		//written, err := writer.Write([]byte(json))
		response := backend.CallResourceResponse{
			Status:  200,
			Headers: make(map[string][]string),
			Body:    []byte(json),
		}
		err := sender.Send(&response)
		if err != nil {
			log.DefaultLogger.Error("Unable to write projects to client.")
			return err
		}
		log.DefaultLogger.Info("Projects sent to client.")
	case "newProject":
		break
	case "newSubsystem":
		break
	case "newDatapoint":
		break
	default:
		log.DefaultLogger.Error("Resource " + request.URL + " not found.")
		//writer.WriteHeader(404)
		response := backend.CallResourceResponse{
			Status:  404,
			Headers: make(map[string][]string),
			Body:    []byte("Not found"),
		}
		err := sender.Send(&response)
		return err
	}
	return nil
}
