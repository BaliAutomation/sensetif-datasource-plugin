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
	cassandraHosts, cassandraClient := createCassandraClient()
	kafkaClient := createKafkaClient()
	stripeClient := createStripeClient()
	clients := client.Clients{
		Cassandra: &cassandraClient,
		Kafka:     &kafkaClient,
		Stripe:    &stripeClient,
	}
	resourceHandler := ResourceHandler{
		Clients: &clients,
	}

	ds := createDatasource(&cassandraClient, cassandraHosts)
	startServing(ds, &resourceHandler)
}

func createCassandraClient() ([]string, client.CassandraClient) {
	log.DefaultLogger.Info("createCassandraClient()")
	cassandraHosts := cassandraHosts()
	cassandraClient := client.CassandraClient{}
	cassandraClient.InitializeCassandra(cassandraHosts)
	return cassandraHosts, cassandraClient
}

func createKafkaClient() client.KafkaClient {
	log.DefaultLogger.Info("createKafkaClient()")
	kafkaHosts := kafkaHosts()
	kafkaClient := client.KafkaClient{}
	clientId, err := os.Hostname()
	if err != nil {
		log.DefaultLogger.Error(fmt.Sprintf("Unable to get os.Hostname(): %s", err))
		clientId = "grafana" + strconv.FormatInt(rand.Int63(), 10)
	}
	kafkaClient.InitializeKafka(kafkaHosts, clientId)
	log.DefaultLogger.Info("kafkaClient: " + fmt.Sprintf("%+v", &kafkaClient))
	return kafkaClient
}

func createStripeClient() client.StripeClient {
	log.DefaultLogger.Info("createStripeClient()")
	stripeClient := client.StripeClient{}
	stripeKey := stripeAuthKey()
	if strings.HasPrefix(stripeKey, "sk_live") {
		log.DefaultLogger.Info("****** Stripe PRODUCTION Key is used!!!!!!!")
	} else {
		log.DefaultLogger.Info("****** Stripe TEST Key is used!!!!!!!")
	}
	stripeClient.InitializeStripe(stripeKey)
	return stripeClient
}

func stripeAuthKey() string {
	if key, ok := os.LookupEnv("STRIPE_KEY"); ok {
		return key
	}
	// If not set in environment, return the key for the Strip Test Mode.
	return "sk_test_51JZvsFBil9jp3I2LySc7piIiEpXUlDdcxpXdVERSLL10nv2AUM1dfoCjSAZIMJ2XlC8zK1tkxJw85F2KlkBh9mxE00Vne8Kp5Z"
}

func startServing(ds SensetifDatasource, resourceHandler *ResourceHandler) {
	log.DefaultLogger.Info("startServing()")
	log.DefaultLogger.Info("Kafka Client: " + fmt.Sprintf("%+v", resourceHandler.Clients.Kafka))
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
