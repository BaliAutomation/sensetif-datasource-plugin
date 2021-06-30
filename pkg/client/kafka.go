package client

import (
	"fmt"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

type Kafka interface {
	Send(topic string, key string, value []byte)
}

type KafkaClient struct {
	p *kafka.Producer
}

func (kaf *KafkaClient) Send(topic string, key string, value []byte) {
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
	} else {
		data := string(value)
		log.DefaultLogger.Info(fmt.Sprintf("Sent message to key %s on topic %s. Data: %s", key, topic, data))
	}
}

func (kaf *KafkaClient) InitializeKafka(hosts []string, clientId string) {
	var err error
	config := kafka.ConfigMap{
		"bootstrap.servers":  strings.Join(hosts, ","),
		"client.id":          clientId,
		"acks":               "all",
		"enable.idempotence": "true",
	}
	kaf.p, err = kafka.NewProducer(&config)
	if err != nil {
		log.DefaultLogger.Error("Failed to create producer: " + err.Error())
	} else {
		log.DefaultLogger.Info(fmt.Sprintf("Created Kafka producer: %+v, %+v", config, kaf.p))
	}
}
