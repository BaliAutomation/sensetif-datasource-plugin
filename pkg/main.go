package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/BaliAutomation/sensetif-datasource/pkg/client"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

func main() {
	log.DefaultLogger.Info("Starting Sensetif plugin")
	log.DefaultLogger.Info("createCassandraClient()")
	cassandra_hosts := cassandraHosts()
	cassandraClient := client.CassandraClient{}
	cassandraClient.InitializeCassandra(cassandra_hosts)
	log.DefaultLogger.Info("createKafkaClient()")
	kafka_hosts := kafkaHosts()
	kafkaClient := client.KafkaClient{}
	clientId, err := os.Hostname()
	if err != nil {
		log.DefaultLogger.Error(fmt.Sprintf("Unable to get os.Hostname(): %s", err))
		clientId = "grafana" + strconv.FormatInt(rand.Int63(), 10)
	}
	kafkaClient.InitializeKafka(kafka_hosts, clientId)
	log.DefaultLogger.Info("createResourceHandler(): " + fmt.Sprintf("%+v", &kafkaClient))
	resourceHandler := ResourceHandler{
		cassandra: &cassandraClient,
		kafka:     &kafkaClient,
	}
	ds := createDatasource(&cassandraClient, cassandra_hosts)
	startServing(ds, &resourceHandler)
}

func startServing(ds SensetifDatasource, resourceHandler *ResourceHandler) {
	log.DefaultLogger.Info("startServing()")
	log.DefaultLogger.Info("Kafka Client: " + fmt.Sprintf("%+v", resourceHandler.kafka))
	serveOpts := datasource.ServeOpts{
		CallResourceHandler: resourceHandler,
		QueryDataHandler:    &ds,
		CheckHealthHandler:  &ds,
	}

	err := datasource.Serve(serveOpts)
	if err != nil {
		log.DefaultLogger.Error(err.Error())
		os.Exit(1)
	}
}

func createDatasource(cassandraClient *client.CassandraClient, hosts []string) SensetifDatasource {
	log.DefaultLogger.Info("createDatasource()")
	ds := SensetifDatasource{
		cassandraClient: cassandraClient,
		hosts:           hosts,
	}
	ds.initializeInstance()
	return ds
}

func cassandraHosts() []string {
	log.DefaultLogger.Info("cassandraHosts()")
	if hosts, ok := os.LookupEnv("CASSANDRA_HOSTS"); ok {
		log.DefaultLogger.Info(fmt.Sprintf("Found Cassandra Hosts:%s", hosts))
		return strings.Split(hosts, ",")
	}
	return []string{"192.168.1.42"} // Default at Niclas' lab
}

func kafkaHosts() []string {
	log.DefaultLogger.Info("kafkaHosts()")
	if hosts, ok := os.LookupEnv("KAFKA_HOSTS"); ok {
		log.DefaultLogger.Info(fmt.Sprintf("Found Kafka Hosts:%s", hosts))
		return strings.Split(hosts, ",")
	}
	return []string{"192.168.1.42:9092"} // Default at Niclas' lab
}
