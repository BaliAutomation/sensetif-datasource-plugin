package handler

import (
	"fmt"
	"github.com/BaliAutomation/sensetif-datasource/pkg/client"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"io/ioutil"
	"net/http"
)

func PrivacyPolicy(orgId int64, params []string, body []byte, kafka client.Kafka, cassandra client.Cassandra) (*backend.CallResourceResponse, error) {
	filename := "privacy.html"
	return returnFileContent(filename)
}

func TermsOfService(orgId int64, params []string, body []byte, kafka client.Kafka, cassandra client.Cassandra) (*backend.CallResourceResponse, error) {
	filename := "tos.html"
	return returnFileContent(filename)
}

func returnFileContent(filename string) (*backend.CallResourceResponse, error) {
	policy, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("File reading error", err)
		return &backend.CallResourceResponse{
			Status:  http.StatusNotFound,
			Headers: make(map[string][]string),
			Body:    []byte("Can not find file " + filename + ". " + err.Error()),
		}, nil
	}

	return &backend.CallResourceResponse{
		Status:  http.StatusOK,
		Headers: make(map[string][]string),
		Body:    policy,
	}, nil
}
