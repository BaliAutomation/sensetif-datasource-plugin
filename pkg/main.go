package main

import (
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"os"
	"strings"
)

func main() {
	hosts := cassandraHosts()
	ds := SensetifDatasource{}
	err := datasource.Serve(ds.newDatasource(hosts))
	if err != nil {
		log.DefaultLogger.Error(err.Error())
		os.Exit(1)
	}
}

func cassandraHosts() []string {
	vars := os.Environ()
	for i := 0; i < len(vars); i++ {
		v := vars[i]
		split := strings.Split(v, "=")
		if split[0] == "CASSANDRA_HOSTS" {
			return strings.Split(split[1], ",")
		}
	}
	return []string{"192.168.1.42:9042"}
}
