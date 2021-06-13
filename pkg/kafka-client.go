package main

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"strings"
)

type KafkaClient struct {
	p *kafka.Producer
}

func (kaf *KafkaClient) send(topic string, key string, value []byte) {
	delivery_chan := make(chan kafka.Event, 10000)
	t := topic
	err := kaf.p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &t,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(key),
		Value: value,
	}, delivery_chan)
	if err != nil {
		log.DefaultLogger.Error(fmt.Sprintf("Failed to send a message: %s\n%s : %+v\n", err, key, value))
	}
}

func (kaf *KafkaClient) initializeKafka(hosts []string, clientId string) {
	var err error
		kaf.p, err = kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": strings.Join(hosts, ","),
		"client.id":         clientId,
		"acks":              "all"})

	if err != nil {
		fmt.Printf("Failed to create producer: %s\n", err)
	}
}
