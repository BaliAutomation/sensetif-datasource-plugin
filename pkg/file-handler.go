package main

import (
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"io/ioutil"
	"net/http"
)

func HandleFile(path string) (*backend.CallResourceResponse, error) {

	filename := "/var/lib/grafana/plugins/sensetif-datasource/" + path
	return returnFileContent(filename)
}

func returnFileContent(filename string) (*backend.CallResourceResponse, error) {
	policy, err := ioutil.ReadFile(filename)
	if err != nil {
		log.DefaultLogger.Error("File reading error: " + err.Error())
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
